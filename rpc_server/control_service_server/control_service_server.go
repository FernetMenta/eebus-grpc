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
	"sync/atomic"
	"time"

	"github.com/enbility/eebus-go/api"
	"github.com/enbility/eebus-go/service"
	"github.com/enbility/eebus-grpc/eebus_service"
	"github.com/enbility/eebus-grpc/rpc_server"
	"github.com/enbility/eebus-grpc/rpc_server/usecase_router"
	cs_lpc_server "github.com/enbility/eebus-grpc/rpc_server/usecase_server/cs/lpc"
	cs_lpp_server "github.com/enbility/eebus-grpc/rpc_server/usecase_server/cs/lpp"
	eg_lpc_server "github.com/enbility/eebus-grpc/rpc_server/usecase_server/eg/lpc"
	eg_lpp_server "github.com/enbility/eebus-grpc/rpc_server/usecase_server/eg/lpp"
	"github.com/enbility/eebus-grpc/rpc_services/control_service"
	cs_lpc_pb "github.com/enbility/eebus-grpc/rpc_services/usecases/cs/lpc"
	cs_lpp_pb "github.com/enbility/eebus-grpc/rpc_services/usecases/cs/lpp"
	eg_lpc_pb "github.com/enbility/eebus-grpc/rpc_services/usecases/eg/lpc"
	eg_lpp_pb "github.com/enbility/eebus-grpc/rpc_services/usecases/eg/lpp"
	shipapi "github.com/enbility/ship-go/api"
	"github.com/enbility/ship-go/cert"
	"github.com/enbility/spine-go/model"
	"github.com/enbility/spine-go/spine"
	"github.com/looplab/fsm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	log "github.com/enbility/eebus-grpc/utils/logging"
	"github.com/enbility/eebus-grpc/utils/type_conversions"
)

type Server struct {
	control_service.UnsafeControlServiceServer
	rpc_server.RPCServer
	eebusService     *eebus_service.Service
	stateMachine     *fsm.FSM
	grpcServer       *grpc.Server
	grpcHealthServer *health.Server
	certificate      *tls.Certificate

	mu       sync.Mutex
	registry *usecase_router.Registry
	// interfaces restricts mDNS/SHIP to these network interfaces (all if empty).
	interfaces []string

	// instanceCounter is used to generate unique instance IDs.
	instanceCounter atomic.Uint64

	log.Logger
}

func NewServer(eebusService *eebus_service.Service, certificate *tls.Certificate) *Server {
	return &Server{
		eebusService: eebusService,
		stateMachine: NewStateMachine(),
		grpcServer:   nil,
		certificate:  certificate,
		registry:     usecase_router.NewRegistry(),
	}
}

// SetInterfaces restricts mDNS and SHIP to the given network interface names.
func (h *Server) SetInterfaces(ifaces []string) {
	h.interfaces = ifaces
}

// nextInstanceID generates a unique instance ID.
func (h *Server) nextInstanceID() string {
	id := h.instanceCounter.Add(1)
	return fmt.Sprintf("uc-%d", id)
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
	h.mu.Unlock()

	h.registry.ShutdownAll()

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
	if len(h.interfaces) > 0 {
		configuration.SetInterfaces(h.interfaces)
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
	entity := h.eebusService.LocalDevice().Entity(entityAddress)

	actor := model.UseCaseActorType(req.GetUseCase().GetActor().String())
	useCaseName := model.UseCaseNameType(req.GetUseCase().GetName().String())

	filter := model.UseCaseFilterType{
		Actor:       actor,
		UseCaseName: useCaseName,
	}

	if entity.HasUseCaseSupport(filter) {
		return &control_service.AddUseCaseResponse{}, fmt.Errorf("Use case already supported")
	}

	instanceID := h.nextInstanceID()

	switch actor {
	case model.UseCaseActorTypeCEM:
		switch useCaseName {
		case model.UseCaseNameTypeCoordinatedEVCharging,
			model.UseCaseNameTypeEVCommissioningAndConfiguration,
			model.UseCaseNameTypeMeasurementOfElectricityDuringEVCharging,
			model.UseCaseNameTypeEVSECommissioningAndConfiguration,
			model.UseCaseNameTypeEVStateOfCharge,
			model.UseCaseNameTypeOverloadProtectionByEVChargingCurrentCurtailment,
			model.UseCaseNameTypeOptimizationOfSelfConsumptionDuringEVCharging,
			model.UseCaseNameTypeVisualizationOfAggregatedBatteryData,
			model.UseCaseNameTypeVisualizationOfAggregatedPhotovoltaicData:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		default:
			return &control_service.AddUseCaseResponse{}, fmt.Errorf("Unknown use case '%v' for role CEM", req.GetUseCase().GetName().String())
		}
	case model.UseCaseActorTypeControllableSystem:
		switch useCaseName {
		case model.UseCaseNameTypeLimitationOfPowerConsumption:
			srv := cs_lpc_server.NewServer(&h.eebusService.Service, entity)
			h.registry.RegisterCSLPC(instanceID, srv)

		case model.UseCaseNameTypeLimitationOfPowerProduction:
			srv := cs_lpp_server.NewServer(&h.eebusService.Service, entity)
			h.registry.RegisterCSLPP(instanceID, srv)
		default:
			return &control_service.AddUseCaseResponse{}, fmt.Errorf("Unknown use case '%v' for role CS", req.GetUseCase().GetName().String())
		}
	case model.UseCaseActorTypeEnergyGuard:
		switch useCaseName {
		case model.UseCaseNameTypeLimitationOfPowerConsumption:
			srv := eg_lpc_server.NewServer(&h.eebusService.Service, entity)
			h.registry.RegisterEGLPC(instanceID, srv)
		case model.UseCaseNameTypeLimitationOfPowerProduction:
			srv := eg_lpp_server.NewServer(&h.eebusService.Service, entity)
			h.registry.RegisterEGLPP(instanceID, srv)
		default:
			return &control_service.AddUseCaseResponse{}, fmt.Errorf("Unknown use case '%v' for role EG", req.GetUseCase().GetName().String())
		}
	case model.UseCaseActorTypeMonitoringAppliance:
		switch useCaseName {
		case model.UseCaseNameTypeMonitoringOfGridConnectionPoint,
			model.UseCaseNameTypeMonitoringOfPowerConsumption:
			return &control_service.AddUseCaseResponse{}, errors.New("Not implemented")
		default:
			return &control_service.AddUseCaseResponse{}, fmt.Errorf("Unknown use case '%v' for role MA", req.GetUseCase().GetName().String())
		}
	default:
		return &control_service.AddUseCaseResponse{}, fmt.Errorf("Unknown role '%v'", req.GetUseCase().GetActor().String())
	}

	return &control_service.AddUseCaseResponse{
		InstanceId: instanceID,
	}, nil
}

func (h *Server) SubscribeUseCaseEvents(req *control_service.SubscribeUseCaseEventsRequest, stream control_service.ControlService_SubscribeUseCaseEventsServer) error {
	instanceID := req.GetInstanceId()

	useCaseServer, ok := h.registry.Get(instanceID)
	if !ok {
		return fmt.Errorf("no use case instance with id %q", instanceID)
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

	// Register the control service
	control_service.RegisterControlServiceServer(h.grpcServer, h)

	// Register all use case routers on the single server
	cs_lpc_pb.RegisterControllableSystemLPCControlServer(h.grpcServer, usecase_router.NewCSLPCRouter(h.registry))
	cs_lpp_pb.RegisterControllableSystemLPPControlServer(h.grpcServer, usecase_router.NewCSLPPRouter(h.registry))
	eg_lpc_pb.RegisterEnergyGuardLPCControlServer(h.grpcServer, usecase_router.NewEGLPCRouter(h.registry))
	eg_lpp_pb.RegisterEnergyGuardLPPControlServer(h.grpcServer, usecase_router.NewEGLPPRouter(h.registry))

	// Enable health server
	h.grpcHealthServer = health.NewServer()
	grpc_health_v1.RegisterHealthServer(h.grpcServer, h.grpcHealthServer)
	h.grpcHealthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		if err := h.grpcServer.Serve(lis); err != nil {
			log.Errorf("failed to serve: %v", err)
		}
	}()

	return *port, nil
}

func (h *Server) Stop() error {
	if h.grpcServer == nil {
		return fmt.Errorf("grpc server is not running")
	}
	h.registry.ShutdownAll()
	if h.grpcHealthServer != nil {
		h.grpcHealthServer.Shutdown()
		h.grpcHealthServer = nil
	}
	h.grpcServer.Stop()
	h.grpcServer = nil

	return nil
}
