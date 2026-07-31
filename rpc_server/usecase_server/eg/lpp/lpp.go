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

func (h *Server) ProductionLimit(_ context.Context, req *lpp.ProductionLimitRequest) (*lpp.ProductionLimitResponse, error) {
	remoteDevice := h.eebusService.LocalDevice().RemoteDeviceForSki(req.GetRemoteSki())
	remoteEntity := remoteDevice.Entity(type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))
	eebus_limit, resultErr := h.uclpp.ProductionLimit(remoteEntity)
	rpc_limit := type_conversions.ConvertEEBUSLoadLimit(&eebus_limit)
	return &lpp.ProductionLimitResponse{
		Limit: &rpc_limit,
	}, resultErr
}

func (h *Server) WriteProductionLimit(_ context.Context, req *lpp.WriteProductionLimitRequest) (*lpp.WriteProductionLimitResponse, error) {
	remoteDevice := h.eebusService.LocalDevice().RemoteDeviceForSki(req.GetRemoteSki())
	remoteEntity := remoteDevice.Entity(type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))

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
	remoteDevice := h.eebusService.LocalDevice().RemoteDeviceForSki(req.GetRemoteSki())
	remoteEntity := remoteDevice.Entity(type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))

	limit, resultErr := h.uclpp.FailsafeProductionActivePowerLimit(remoteEntity)

	return &lpp.FailsafeProductionActivePowerLimitResponse{
		Limit: limit,
	}, resultErr
}

func (h *Server) WriteFailsafeProductionActivePowerLimit(_ context.Context, req *lpp.WriteFailsafeProductionActivePowerLimitRequest) (*lpp.WriteFailsafeProductionActivePowerLimitResponse, error) {
	remoteDevice := h.eebusService.LocalDevice().RemoteDeviceForSki(req.GetRemoteSki())
	remoteEntity := remoteDevice.Entity(type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))

	msgCounter, resultErr := h.uclpp.WriteFailsafeProductionActivePowerLimit(
		remoteEntity,
		req.GetLimit(),
	)

	return &lpp.WriteFailsafeProductionActivePowerLimitResponse{
		MsgCounter: uint64(*msgCounter),
	}, resultErr
}

func (h *Server) FailsafeDurationMinimum(_ context.Context, req *lpp.FailsafeDurationMinimumRequest) (*lpp.FailsafeDurationMinimumResponse, error) {
	remoteDevice := h.eebusService.LocalDevice().RemoteDeviceForSki(req.GetRemoteSki())
	remoteEntity := remoteDevice.Entity(type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))

	duration, resultErr := h.uclpp.FailsafeDurationMinimum(remoteEntity)

	return &lpp.FailsafeDurationMinimumResponse{
		DurationNanoseconds: uint64(duration.Nanoseconds()),
	}, resultErr
}

func (h *Server) WriteFailsafeDurationMinimum(_ context.Context, req *lpp.WriteFailsafeDurationMinimumRequest) (*lpp.WriteFailsafeDurationMinimumResponse, error) {
	remoteDevice := h.eebusService.LocalDevice().RemoteDeviceForSki(req.GetRemoteSki())
	remoteEntity := remoteDevice.Entity(type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))

	msgCounter, resultErr := h.uclpp.WriteFailsafeDurationMinimum(
		remoteEntity,
		time.Duration(req.GetDurationNanoseconds())*time.Nanosecond,
	)

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
	remoteDevice := h.eebusService.LocalDevice().RemoteDeviceForSki(req.GetRemoteSki())
	remoteEntity := remoteDevice.Entity(type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))

	withinDuration := h.uclpp.IsHeartbeatWithinDuration(remoteEntity)

	return &lpp.IsHeartbeatWithinDurationResponse{
		WithinDuration: withinDuration,
	}, nil
}

func (h *Server) ProductionNominalMax(_ context.Context, req *lpp.ProductionNominalMaxRequest) (*lpp.ProductionNominalMaxResponse, error) {
	remoteDevice := h.eebusService.LocalDevice().RemoteDeviceForSki(req.GetRemoteSki())
	remoteEntity := remoteDevice.Entity(type_conversions.ConvertRPCEntityAddress(req.GetRemoteEntityAddress()))

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
	h.grpcServer.Stop()
	h.Debug("use case server stopped\n")
	h.grpcServer = nil
	return nil
}
