// SPDX-License-Identifier: Apache-2.0
// Copyright Pionix GmbH and Contributors to EVerest

package control_service_server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/enbility/eebus-go/api"
	"github.com/enbility/eebus-go/service"
	"github.com/enbility/eebus-grpc/eebus_service"
	"github.com/enbility/eebus-grpc/rpc_server"
	"github.com/enbility/eebus-grpc/rpc_server/usecase_server"
	cs_lpc_server "github.com/enbility/eebus-grpc/rpc_server/usecase_server/cs/lpc"
	cs_lpp_server "github.com/enbility/eebus-grpc/rpc_server/usecase_server/cs/lpp"
	eg_lpc_server "github.com/enbility/eebus-grpc/rpc_server/usecase_server/eg/lpc"
	eg_lpp_server "github.com/enbility/eebus-grpc/rpc_server/usecase_server/eg/lpp"
	"github.com/enbility/eebus-grpc/rpc_services/control_service"
	shipapi "github.com/enbility/ship-go/api"
	"github.com/enbility/ship-go/cert"
	"github.com/enbility/spine-go/model"
	"github.com/enbility/spine-go/spine"
	"github.com/looplab/fsm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	// Use cases - CEM

	// Use cases - CS

	// Use cases - EG

	// Use cases - MA

	// Conversion functions

	log "github.com/enbility/eebus-grpc/utils/logging"
	"github.com/enbility/eebus-grpc/utils/type_conversions"
)

type useCaseEntryType struct {
	entityAddress []model.AddressEntityType
	useCase       *control_service.UseCase
	useCaseServer usecase_server.UseCaseServer
}

func (h *useCaseEntryType) Equals(other *useCaseEntryType) bool {
	// compare entity address
	if len(h.entityAddress) != len(other.entityAddress) {
		return false
	}
	for i, address := range h.entityAddress {
		if address != other.entityAddress[i] {
			return false
		}
	}
	// compare use case
	if h.useCase.GetActor() != other.useCase.GetActor() {
		return false
	}
	if h.useCase.GetName() != other.useCase.GetName() {
		return false
	}
	return true
}

type Server struct {
	control_service.UnsafeControlServiceServer
	rpc_server.RPCServer
	eebusService     *eebus_service.Service
	stateMachine     *fsm.FSM
	grpcServer       *grpc.Server
	grpcHealthServer *health.Server
	grpcIpv4Addr     string
	certificate      *tls.Certificate

	mu              sync.Mutex
	useCaseRegistry []useCaseEntryType

	log.Logger
}

func NewServer(eebusService *eebus_service.Service, certificate *tls.Certificate) *Server {
	return &Server{
		eebusService: eebusService,
		stateMachine: NewStateMachine(),
		grpcServer:   nil,
		certificate:  certificate,
		grpcIpv4Addr: "",
	}
}

func (h *Server) StartService(_ context.Context, _ *control_service.EmptyRequest) (*control_service.EmptyResponse, error) {
	h.Debug("Received start command")
	if h.eebusService.IsRunning() {
		return &control_service.EmptyResponse{}, fmt.Errorf("Service is already running")
	}

	h.mu.Lock()
	if h.stateMachine.Cannot(EventStart) {
		h.mu.Unlock()
		return &control_service.EmptyResponse{}, fmt.Errorf("Service is not ready")
	}
	h.mu.Unlock()

	h.eebusService.Start()
	if !h.eebusService.IsRunning() {
		return &control_service.EmptyResponse{}, fmt.Errorf("Failed to start service")
	}

	h.mu.Lock()
	h.stateMachine.Event(context.Background(), EventStart)
	h.mu.Unlock()

	return &control_service.EmptyResponse{}, nil
}

func (h *Server) StopService(_ context.Context, _ *control_service.EmptyRequest) (*control_service.EmptyResponse, error) {
	h.Debug("Received stop command")
	if !h.eebusService.IsRunning() {
		return &control_service.EmptyResponse{}, fmt.Errorf("Service is not running")
	}

	h.mu.Lock()
	if h.stateMachine.Cannot(EventStop) {
		h.mu.Unlock()
		return &control_service.EmptyResponse{}, fmt.Errorf("Service is not running")
	}
	h.mu.Unlock()

	h.eebusService.Shutdown()
	if h.eebusService.IsRunning() {
		return &control_service.EmptyResponse{}, fmt.Errorf("Failed to stop service")
	}

	h.mu.Lock()
	h.stateMachine.Event(context.Background(), EventStop)
	h.mu.Unlock()

	return &control_service.EmptyResponse{}, nil
}

func (h *Server) ResetService(_ context.Context, _ *control_service.EmptyRequest) (*control_service.EmptyResponse, error) {
	h.Debug("Received reset command")
	h.eebusService.Shutdown()

	h.mu.Lock()
	h.stateMachine = NewStateMachine()

	var useCaseServer usecase_server.UseCaseServer = nil
	for _, entry := range h.useCaseRegistry {
		useCaseServer = entry.useCaseServer
		if useCaseServer != nil {
			useCaseServer.Stop()
		}
	}
	h.useCaseRegistry = []useCaseEntryType{}
	h.mu.Unlock()

	return &control_service.EmptyResponse{}, nil
}

func (h *Server) SetConfig(_ context.Context, req *control_service.SetConfigRequest) (*control_service.EmptyResponse, error) {
	h.Debug("Received set config command")

	h.mu.Lock()
	if !h.stateMachine.Is(StateSetup) {
		h.mu.Unlock()
		h.Error("Service is not in setup state")
		return &control_service.EmptyResponse{}, fmt.Errorf("Service is not in setup state")
	}
	h.mu.Unlock()
	var device_categories []shipapi.DeviceCategoryType
	for _, category := range req.GetDeviceCategories() {
		device_categories = append(device_categories, shipapi.DeviceCategoryType(category))
	}

	var entity_types []model.EntityTypeType
	for _, entity_type := range req.GetEntityTypes() {
		entity_types = append(entity_types, type_conversions.ConvertRPCEntityType(entity_type))
	}
	configuration, err := api.NewConfiguration(
		req.GetVendorCode(),
		req.GetDeviceBrand(),
		req.GetDeviceModel(),
		req.GetSerialNumber(),
		device_categories,
		type_conversions.ConvertRPCDeviceType(req.DeviceType),
		entity_types,
		int(req.GetPort()),
		*h.certificate,
		time.Duration(req.GetHeartbeatTimeoutSeconds())*time.Second,
		nil,
		nil,
	)
	if err != nil {
		h.Errorf("Failed to create configuration: %v", err)
		return &control_service.EmptyResponse{}, fmt.Errorf("Failed to create configuration: %v", err)
	}
	h.eebusService.Service = *service.NewService(configuration, h.eebusService)
	h.eebusService.SetLogging(h)

	h.Debug("Successfully set configuration")
	return &control_service.EmptyResponse{}, nil
}

func (h *Server) StartSetup(_ context.Context, _ *control_service.EmptyRequest) (*control_service.EmptyResponse, error) {
	h.Debug("Received start setup command")

	h.mu.Lock()
	if h.stateMachine.Cannot(EventReady) {
		h.mu.Unlock()
		return &control_service.EmptyResponse{}, fmt.Errorf("Service is not in setup state")
	}
	h.mu.Unlock()

	var err error
	err = h.eebusService.Setup()
	if err != nil {
		return &control_service.EmptyResponse{}, fmt.Errorf("Failed to setup service: %v", err)
	}

	h.mu.Lock()
	h.stateMachine.Event(context.Background(), EventReady)
	h.mu.Unlock()

	return &control_service.EmptyResponse{}, nil
}

func (h *Server) AddEntity(_ context.Context, req *control_service.AddEntityRequest) (*control_service.EmptyResponse, error) {
	h.Debug("Received add entity command")

	h.mu.Lock()
	if h.stateMachine.Is(StateSetup) {
		h.mu.Unlock()
		return &control_service.EmptyResponse{}, fmt.Errorf("Service is not in ready state")
	}
	h.mu.Unlock()

	entityAddress := type_conversions.ConvertRPCEntityAddress(req.GetAddress())
	res := h.eebusService.LocalDevice().Entity(entityAddress)
	if res != nil {
		return &control_service.EmptyResponse{}, fmt.Errorf("Entity address already used")
	}

	entity := spine.NewEntityLocal(
		h.eebusService.LocalDevice(),
		model.EntityTypeType(req.GetType().String()),
		entityAddress,
		h.eebusService.Configuration().HeartbeatTimeout(),
	)
	h.eebusService.LocalDevice().AddEntity(entity)

	return &control_service.EmptyResponse{}, nil
}

func (h *Server) RemoveEntity(_ context.Context, req *control_service.RemoveEntityRequest) (*control_service.EmptyResponse, error) {
	h.Debug("Received remove entity command")

	h.mu.Lock()
	if h.stateMachine.Is(StateSetup) {
		h.mu.Unlock()
		return &control_service.EmptyResponse{}, fmt.Errorf("Service is not in ready state")
	}
	h.mu.Unlock()

	entityAddress := type_conversions.ConvertRPCEntityAddress(req.GetAddress())
	entity := h.eebusService.LocalDevice().Entity(entityAddress)
	if entity == nil {
		return &control_service.EmptyResponse{}, fmt.Errorf("Entity not found")
	}
	h.eebusService.LocalDevice().RemoveEntity(entity)

	return &control_service.EmptyResponse{}, nil
}

func (h *Server) RegisterRemoteSki(_ context.Context, req *control_service.RegisterRemoteSkiRequest) (*control_service.EmptyResponse, error) {
	h.Debug("Received register remote SKI command")

	h.mu.Lock()
	if !h.stateMachine.Is(StateReady) && !h.stateMachine.Is(StateRunning) {
		h.mu.Unlock()
		return &control_service.EmptyResponse{}, fmt.Errorf("Service must be in ready or running state")
	}
	h.mu.Unlock()

	h.eebusService.RegisterRemoteService(shipapi.ServiceIdentity{SKI: req.GetRemoteSki()})

	return &control_service.EmptyResponse{}, nil
}

func (h *Server) AddUseCase(_ context.Context, req *control_service.AddUseCaseRequest) (*control_service.AddUseCaseResponse, error) {
	h.Debug("Received add use case command")

	entityAddress := type_conversions.ConvertRPCEntityAddress(req.GetEntityAddress())
	h.Debugf("h.eebusService: %v", h.eebusService)
	h.Debugf("h.eebusService.LocalDevice(): %v", h.eebusService.LocalDevice())
	h.Debugf("entityAddress: %v", entityAddress)
	entity := h.eebusService.LocalDevice().Entity(entityAddress)
	h.Debugf("entity: %v", entity)

	actor := model.UseCaseActorType(req.GetUseCase().GetActor().String())
	useCaseName := model.UseCaseNameType(req.GetUseCase().GetName().String())

	filter := model.UseCaseFilterType{
		Actor:       actor,
		UseCaseName: useCaseName,
	}

	if entity.HasUseCaseSupport(filter) {
		return &control_service.AddUseCaseResponse{}, fmt.Errorf("Use case already supported")
	}

	var useCaseEntry useCaseEntryType
	var useCaseServer usecase_server.UseCaseServer

	switch actor {
	case model.UseCaseActorTypeCEM:
		switch useCaseName {
		case model.UseCaseNameTypeCoordinatedEVCharging:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		case model.UseCaseNameTypeEVCommissioningAndConfiguration:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		case model.UseCaseNameTypeMeasurementOfElectricityDuringEVCharging:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		case model.UseCaseNameTypeEVSECommissioningAndConfiguration:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		case model.UseCaseNameTypeEVStateOfCharge:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		case model.UseCaseNameTypeOverloadProtectionByEVChargingCurrentCurtailment:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		case model.UseCaseNameTypeOptimizationOfSelfConsumptionDuringEVCharging:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		case model.UseCaseNameTypeVisualizationOfAggregatedBatteryData:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		case model.UseCaseNameTypeVisualizationOfAggregatedPhotovoltaicData:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		default:
			return &control_service.AddUseCaseResponse{}, fmt.Errorf("Unknown use case '%v' for role CEM", req.GetUseCase().GetName().String())
		}
	case model.UseCaseActorTypeControllableSystem:
		switch useCaseName {
		case model.UseCaseNameTypeLimitationOfPowerConsumption:
			useCaseServer = cs_lpc_server.NewServer(
				&h.eebusService.Service,
				entity,
			)
			useCaseEntry = useCaseEntryType{
				entityAddress: entityAddress,
				useCase:       req.GetUseCase(),
				useCaseServer: useCaseServer,
			}

		case model.UseCaseNameTypeLimitationOfPowerProduction:
			useCaseServer = cs_lpp_server.NewServer(
				&h.eebusService.Service,
				entity,
			)
			useCaseEntry = useCaseEntryType{
				entityAddress: entityAddress,
				useCase:       req.GetUseCase(),
				useCaseServer: useCaseServer,
			}
		default:
			return &control_service.AddUseCaseResponse{}, fmt.Errorf("Unknown use case '%v' for role CS", req.GetUseCase().GetName().String())
		}
	case model.UseCaseActorTypeEnergyGuard:
		switch useCaseName {
		case model.UseCaseNameTypeLimitationOfPowerConsumption:
			useCaseServer = eg_lpc_server.NewServer(
				&h.eebusService.Service,
				entity,
			)
			useCaseEntry = useCaseEntryType{
				entityAddress: entityAddress,
				useCase:       req.GetUseCase(),
				useCaseServer: useCaseServer,
			}
		case model.UseCaseNameTypeLimitationOfPowerProduction:
			useCaseServer = eg_lpp_server.NewServer(
				&h.eebusService.Service,
				entity,
			)
			useCaseEntry = useCaseEntryType{
				entityAddress: entityAddress,
				useCase:       req.GetUseCase(),
				useCaseServer: useCaseServer,
			}
		default:
			return &control_service.AddUseCaseResponse{}, fmt.Errorf("Unknown use case '%v' for role EG", req.GetUseCase().GetName().String())
		}
	case model.UseCaseActorTypeMonitoringAppliance:
		switch useCaseName {
		case model.UseCaseNameTypeMonitoringOfGridConnectionPoint:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		case model.UseCaseNameTypeMonitoringOfPowerConsumption:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		default:
			return &control_service.AddUseCaseResponse{}, fmt.Errorf("Unknown use case '%v' for role MA", req.GetUseCase().GetName().String())
		}
	default:
		return &control_service.AddUseCaseResponse{}, fmt.Errorf("Unknown role '%v'", req.GetUseCase().GetActor().String())
	}
	h.mu.Lock()
	h.useCaseRegistry = append(h.useCaseRegistry, useCaseEntry)
	h.mu.Unlock()

	port, err := useCaseServer.Start(h.grpcIpv4Addr, nil)
	if err != nil {
		return &control_service.AddUseCaseResponse{}, fmt.Errorf("Failed to start use case server: %v", err)
	}

	return &control_service.AddUseCaseResponse{
		Endpoint: fmt.Sprintf("%s:%v", h.grpcIpv4Addr, port),
	}, nil
}

func (h *Server) SubscribeUseCaseEvents(req *control_service.SubscribeUseCaseEventsRequest, stream control_service.ControlService_SubscribeUseCaseEventsServer) error {
	entityAddress := type_conversions.ConvertRPCEntityAddress(req.GetEntityAddress())

	var useCaseServer usecase_server.UseCaseServer = nil
	referenceEntry := useCaseEntryType{
		entityAddress: entityAddress,
		useCase:       req.GetUseCase(),
		useCaseServer: nil,
	}

	h.mu.Lock()
	for _, entry := range h.useCaseRegistry {
		if entry.Equals(&referenceEntry) {
			useCaseServer = entry.useCaseServer
			break
		}
	}
	h.mu.Unlock()

	if useCaseServer == nil {
		return fmt.Errorf("Use case not found")
	}
	for event := range useCaseServer.GetEvents() {
		err := stream.Send(event)
		if err != nil {
			useCaseServer.OnSubscriberDisconnected()
			return fmt.Errorf("Failed to send event: %v", err)
		}
	}
	return nil
}

func (h *Server) SubscribeDiscoveryEvents(
	_ *control_service.SubscribeDiscoveryEventsRequest,
	stream control_service.ControlService_SubscribeDiscoveryEventsServer,
) error {
	h.mu.Lock()
	if !h.stateMachine.Is(StateReady) && !h.stateMachine.Is(StateRunning) {
		h.mu.Unlock()
		return fmt.Errorf("Service must be in ready or running state")
	}
	h.mu.Unlock()

	snapshot, ch, cancel := h.eebusService.SubscribeDiscoveries()
	defer cancel()

	for _, evt := range snapshot {
		if err := stream.Send(evt); err != nil {
			return err
		}
	}

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case evt, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(evt); err != nil {
				return err
			}
		}
	}
}

func (h *Server) GetCertificate(_ context.Context, _ *control_service.GetCertificateRequest) (*control_service.GetCertificateResponse, error) {
	h.Debug("Received get certificate command")

	if h.certificate == nil {
		return nil, fmt.Errorf("no certificate available")
	}

	// PEM-encode the certificate
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: h.certificate.Certificate[0],
	})

	// Compute SKI from the certificate
	leaf, err := x509.ParseCertificate(h.certificate.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	ski, err := cert.SkiFromCertificate(leaf)
	if err != nil {
		return nil, fmt.Errorf("failed to compute SKI: %v", err)
	}

	return &control_service.GetCertificateResponse{
		CertificatePem: string(pemData),
		Ski:            ski,
	}, nil
}

func (h *Server) Start(ipv4Addr string, port *int) (int, error) {
	h.grpcIpv4Addr = ipv4Addr
	if port == nil {
		return -1, fmt.Errorf("no port specified")
	}
	if h.grpcServer != nil {
		return -1, fmt.Errorf("grpc server is already running")
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ipv4Addr, *port))
	if err != nil {
		return -1, fmt.Errorf("failed to listen: %v", err)
	}
	h.grpcServer = grpc.NewServer()
	control_service.RegisterControlServiceServer(h.grpcServer, h)

	// Enbable health server
	h.grpcHealthServer = health.NewServer()
	grpc_health_v1.RegisterHealthServer(h.grpcServer, h.grpcHealthServer)
	h.grpcHealthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		if err := h.grpcServer.Serve(lis); err != nil {
			fmt.Errorf("failed to serve: %v", err)
		}
	}()

	return *port, nil
}

func (h *Server) Stop() error {
	if h.grpcServer == nil {
		return fmt.Errorf("grpc server is not running")
	}
	if h.grpcHealthServer != nil {
		h.grpcHealthServer.Shutdown()
		h.grpcHealthServer = nil
	}
	h.grpcServer.Stop()
	h.grpcServer = nil

	return nil
}
