// SPDX-License-Identifier: Apache-2.0
// Copyright Pionix GmbH and Contributors to EVerest

package lpp

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/enbility/eebus-go/api"
	"github.com/enbility/eebus-go/service"
	uc_eg_lpp "github.com/enbility/eebus-go/usecases/eg/lpp"
	"github.com/enbility/eebus-grpc/rpc_server/usecase_server"
	"github.com/enbility/eebus-grpc/rpc_services/usecases/eg/lpp"
	"github.com/enbility/eebus-grpc/utils/type_conversions"
	spineapi "github.com/enbility/spine-go/api"
	"github.com/enbility/spine-go/model"
	"google.golang.org/grpc"
)

type Server struct {
	eebusService *service.Service
	uclpp        *uc_eg_lpp.LPP

	lpp.UnimplementedEnergyGuardLPPControlServer
	usecase_server.BaseUseCaseServer

	grpcServer *grpc.Server
}

func NewServer(eebusService *service.Service, localEntity spineapi.EntityLocalInterface) *Server {
	server := &Server{
		eebusService:      eebusService,
		uclpp:             nil,
		BaseUseCaseServer: *usecase_server.NewBaseUseCaseServer(),
		grpcServer:        nil,
	}
	server.uclpp = uc_eg_lpp.NewLPP(localEntity, server.OnUseCaseEvent)
	eebusService.AddUseCase(server.uclpp)
	return server
}

func (h *Server) GetUseCase() api.UseCaseInterface {
	return h.uclpp
}

// resolveRemoteEntity looks up the remote device and entity for the given SKI and entity address.
// Returns an error if either is not found.
func (h *Server) resolveRemoteEntity(ski string, entityAddress []model.AddressEntityType) (spineapi.EntityRemoteInterface, error) {
	remoteDevice := h.eebusService.LocalDevice().RemoteDeviceForSki(ski)
	if remoteDevice == nil {
		return nil, fmt.Errorf("no remote device found for SKI %s", ski)
	}
	remoteEntity := remoteDevice.Entity(entityAddress)
	if remoteEntity == nil {
		return nil, fmt.Errorf("no remote entity found for address %v", entityAddress)
	}
	return remoteEntity, nil
}

func (h *Server) ProductionLimit(_ context.Context, req *lpp.ProductionLimitRequest) (*lpp.ProductionLimitResponse, error) {
	remoteEntity, err := h.resolveRemoteEntity(req.GetRemoteSki(), type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))
	if err != nil {
		return nil, err
	}

	eebus_limit, resultErr := h.uclpp.ProductionLimit(remoteEntity)
	rpc_limit := type_conversions.ConvertEEBUSLoadLimit(&eebus_limit)
	return &lpp.ProductionLimitResponse{
		Limit: &rpc_limit,
	}, resultErr
}

func (h *Server) WriteProductionLimit(_ context.Context, req *lpp.WriteProductionLimitRequest) (*lpp.WriteProductionLimitResponse, error) {
	remoteEntity, err := h.resolveRemoteEntity(req.GetRemoteSki(), type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))
	if err != nil {
		return nil, err
	}

	resultCh := make(chan model.ResultDataType)
	resultCB := func(result model.ResultDataType, msgCounter model.MsgCounterType) {
		resultCh <- result
	}

	msgCounter, resultErr := h.uclpp.WriteProductionLimit(
		remoteEntity,
		type_conversions.ConvertRPCLoadLimit(req.GetLimit()),
		resultCB,
	)

	result := <-resultCh
	errorNumber := uint64(*result.ErrorNumber)

	return &lpp.WriteProductionLimitResponse{
		ErrorNumber: &errorNumber,
		Description: (*string)(result.Description),
		MsgCounter:  uint64(*msgCounter),
	}, resultErr
}

func (h *Server) FailsafeProductionActivePowerLimit(_ context.Context, req *lpp.FailsafeProductionActivePowerLimitRequest) (*lpp.FailsafeProductionActivePowerLimitResponse, error) {
	remoteEntity, err := h.resolveRemoteEntity(req.GetRemoteSki(), type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))
	if err != nil {
		return nil, err
	}

	limit, resultErr := h.uclpp.FailsafeProductionActivePowerLimit(remoteEntity)

	return &lpp.FailsafeProductionActivePowerLimitResponse{
		Limit: limit,
	}, resultErr
}

func (h *Server) WriteFailsafeProductionActivePowerLimit(_ context.Context, req *lpp.WriteFailsafeProductionActivePowerLimitRequest) (*lpp.WriteFailsafeProductionActivePowerLimitResponse, error) {
	remoteEntity, err := h.resolveRemoteEntity(req.GetRemoteSki(), type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))
	if err != nil {
		return nil, err
	}

	msgCounter, resultErr := h.uclpp.WriteFailsafeProductionActivePowerLimit(
		remoteEntity,
		req.GetLimit(),
	)

	if msgCounter == nil {
		return &lpp.WriteFailsafeProductionActivePowerLimitResponse{}, resultErr
	}
	return &lpp.WriteFailsafeProductionActivePowerLimitResponse{
		MsgCounter: uint64(*msgCounter),
	}, resultErr
}

func (h *Server) FailsafeDurationMinimum(_ context.Context, req *lpp.FailsafeDurationMinimumRequest) (*lpp.FailsafeDurationMinimumResponse, error) {
	remoteEntity, err := h.resolveRemoteEntity(req.GetRemoteSki(), type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))
	if err != nil {
		return nil, err
	}

	duration, resultErr := h.uclpp.FailsafeDurationMinimum(remoteEntity)

	return &lpp.FailsafeDurationMinimumResponse{
		DurationNanoseconds: uint64(duration.Nanoseconds()),
	}, resultErr
}

func (h *Server) WriteFailsafeDurationMinimum(_ context.Context, req *lpp.WriteFailsafeDurationMinimumRequest) (*lpp.WriteFailsafeDurationMinimumResponse, error) {
	remoteEntity, err := h.resolveRemoteEntity(req.GetRemoteSki(), type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))
	if err != nil {
		return nil, err
	}

	msgCounter, resultErr := h.uclpp.WriteFailsafeDurationMinimum(
		remoteEntity,
		time.Duration(req.GetDurationNanoseconds())*time.Nanosecond,
	)

	if msgCounter == nil {
		return &lpp.WriteFailsafeDurationMinimumResponse{}, resultErr
	}
	return &lpp.WriteFailsafeDurationMinimumResponse{
		MsgCounter: uint64(*msgCounter),
	}, resultErr
}

func (h *Server) StartHeartbeat(_ context.Context, req *lpp.StartHeartbeatRequest) (*lpp.StartHeartbeatResponse, error) {
	h.uclpp.StartHeartbeat()

	return &lpp.StartHeartbeatResponse{}, nil
}

func (h *Server) StopHeartbeat(_ context.Context, req *lpp.StopHeartbeatRequest) (*lpp.StopHeartbeatResponse, error) {
	h.uclpp.StopHeartbeat()

	return &lpp.StopHeartbeatResponse{}, nil
}

func (h *Server) IsHeartbeatWithinDuration(_ context.Context, req *lpp.IsHeartbeatWithinDurationRequest) (*lpp.IsHeartbeatWithinDurationResponse, error) {
	remoteEntity, err := h.resolveRemoteEntity(req.GetRemoteSki(), type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))
	if err != nil {
		return nil, err
	}

	withinDuration := h.uclpp.IsHeartbeatWithinDuration(remoteEntity)

	return &lpp.IsHeartbeatWithinDurationResponse{
		WithinDuration: withinDuration,
	}, nil
}

func (h *Server) ProductionNominalMax(_ context.Context, req *lpp.ProductionNominalMaxRequest) (*lpp.ProductionNominalMaxResponse, error) {
	remoteEntity, err := h.resolveRemoteEntity(req.GetRemoteSki(), type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))
	if err != nil {
		return nil, err
	}

	nominalMax, resultErr := h.uclpp.ProductionNominalMax(remoteEntity)

	return &lpp.ProductionNominalMaxResponse{
		ProductionNominalMax: nominalMax,
	}, resultErr
}

func (h *Server) Start(ipv4Addr string, port *int) (int, error) {
	var res int
	if port != nil {
		return -1, fmt.Errorf("port is specifiedm but not allowed")
	}
	if h.grpcServer != nil {
		return -1, fmt.Errorf("grpc server is already running")
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:0", ipv4Addr))
	res = lis.Addr().(*net.TCPAddr).Port
	log.Printf("server running at port %d\n", res)

	if err != nil {
		return -1, fmt.Errorf("failed to listen: %v", err)
	}
	h.grpcServer = grpc.NewServer()
	lpp.RegisterEnergyGuardLPPControlServer(h.grpcServer, h)
	go func() {
		if err := h.grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v\n", err)
		}
	}()

	return res, nil
}

func (h *Server) Stop() error {
	if h.grpcServer == nil {
		return fmt.Errorf("grpc server is not running")
	}
	h.CloseEvents()
	h.grpcServer.Stop()
	h.Debug("use case server stopped\n")
	h.grpcServer = nil
	return nil
}
