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
	uc_cs_lpp "github.com/enbility/eebus-go/usecases/cs/lpp"
	"github.com/enbility/eebus-grpc/rpc_server/usecase_server"
	"github.com/enbility/eebus-grpc/rpc_services/common_types"
	"github.com/enbility/eebus-grpc/rpc_services/usecases/cs/lpp"
	"github.com/enbility/eebus-grpc/utils/type_conversions"
	spineapi "github.com/enbility/spine-go/api"
	"github.com/enbility/spine-go/model"
	"google.golang.org/grpc"
)

type Server struct {
	eebusService *service.Service
	uclpp        *uc_cs_lpp.LPP

	lpp.UnimplementedControllableSystemLPPControlServer
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
	server.uclpp = uc_cs_lpp.NewLPP(localEntity, server.OnUseCaseEvent)
	eebusService.AddUseCase(server.uclpp)
	return server
}

func (h *Server) GetUseCase() api.UseCaseInterface {
	return h.uclpp
}

func (h *Server) ProductionLimit(_ context.Context, req *lpp.ProductionLimitRequest) (*lpp.ProductionLimitResponse, error) {
	eebus_limit, resultErr := h.uclpp.ProductionLimit()
	rpc_limit := type_conversions.ConvertEEBUSLoadLimit(&eebus_limit)

	res := &lpp.ProductionLimitResponse{
		LoadLimit: &rpc_limit,
	}
	return res, resultErr
}

func (h *Server) SetProductionLimit(_ context.Context, req *lpp.SetProductionLimitRequest) (*lpp.SetProductionLimitResponse, error) {
	eebus_limit := type_conversions.ConvertRPCLoadLimit(req.GetLoadLimit())
	h.uclpp.IsHeartbeatWithinDuration()
	resultErr := h.uclpp.SetProductionLimit(eebus_limit)

	return &lpp.SetProductionLimitResponse{}, resultErr
}

func (h *Server) PendingProductionLimit(_ context.Context, req *lpp.PendingProductionLimitRequest) (*lpp.PendingProductionLimitResponse, error) {
	eebus_load_limits := h.uclpp.PendingProductionLimits()
	rpc_load_limits := make(map[uint64]*common_types.LoadLimit)
	for eebus_id, eebus_limit := range eebus_load_limits {
		rpc_limit := type_conversions.ConvertEEBUSLoadLimit(&eebus_limit)
		rpc_id := uint64(eebus_id)
		rpc_load_limits[rpc_id] = &rpc_limit
	}

	res := &lpp.PendingProductionLimitResponse{
		LoadLimits: rpc_load_limits,
	}
	return res, nil
}

func (h *Server) ApproveOrDenyProductionLimit(_ context.Context, req *lpp.ApproveOrDenyProductionLimitRequest) (*lpp.ApproveOrDenyProductionLimitResponse, error) {
	h.uclpp.ApproveOrDenyProductionLimit(
		model.MsgCounterType(req.GetMsgCounter()),
		req.GetApprove(),
		req.GetReason(),
	)
	return &lpp.ApproveOrDenyProductionLimitResponse{}, nil
}

func (h *Server) FailsafeProductionActivePowerLimit(_ context.Context, req *lpp.FailsafeProductionActivePowerLimitRequest) (*lpp.FailsafeProductionActivePowerLimitResponse, error) {
	limit, isChangeable, resultErr := h.uclpp.FailsafeProductionActivePowerLimit()

	res := &lpp.FailsafeProductionActivePowerLimitResponse{
		Limit:        limit,
		IsChangeable: isChangeable,
	}

	return res, resultErr
}

func (h *Server) SetFailsafeProductionActivePowerLimit(_ context.Context, req *lpp.SetFailsafeProductionActivePowerLimitRequest) (*lpp.SetFailsafeProductionActivePowerLimitResponse, error) {
	resultErr := h.uclpp.SetFailsafeProductionActivePowerLimit(
		req.GetValue(),
		req.GetIsChangeable(),
	)
	return &lpp.SetFailsafeProductionActivePowerLimitResponse{}, resultErr
}

func (h *Server) FailsafeDurationMinimum(_ context.Context, req *lpp.FailsafeDurationMinimumRequest) (*lpp.FailsafeDurationMinimumResponse, error) {
	duration, isChangeable, resultErr := h.uclpp.FailsafeDurationMinimum()

	res := &lpp.FailsafeDurationMinimumResponse{
		DurationNanoseconds: duration.Nanoseconds(),
		IsChangeable:        isChangeable,
	}

	return res, resultErr
}

func (h *Server) SetFailsafeDurationMinimum(_ context.Context, req *lpp.SetFailsafeDurationMinimumRequest) (*lpp.SetFailsafeDurationMinimumResponse, error) {
	resultErr := h.uclpp.SetFailsafeDurationMinimum(
		time.Duration(req.GetDurationNanoseconds())*time.Nanosecond,
		req.GetIsChangeable(),
	)

	return &lpp.SetFailsafeDurationMinimumResponse{}, resultErr
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
	is_within_duration := h.uclpp.IsHeartbeatWithinDuration()
	res := &lpp.IsHeartbeatWithinDurationResponse{
		IsWithinDuration: is_within_duration,
	}
	return res, nil
}

func (h *Server) ProductionNominalMax(_ context.Context, req *lpp.ProductionNominalMaxRequest) (*lpp.ProductionNominalMaxResponse, error) {
	value, resultErr := h.uclpp.ProductionNominalMax()

	res := &lpp.ProductionNominalMaxResponse{
		Value: value,
	}

	return res, resultErr
}

func (h *Server) SetProductionNominalMax(_ context.Context, req *lpp.SetProductionNominalMaxRequest) (*lpp.SetProductionNominalMaxResponse, error) {
	resultErr := h.uclpp.SetProductionNominalMax(req.GetValue())

	return &lpp.SetProductionNominalMaxResponse{}, resultErr
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
	lpp.RegisterControllableSystemLPPControlServer(h.grpcServer, h)
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
