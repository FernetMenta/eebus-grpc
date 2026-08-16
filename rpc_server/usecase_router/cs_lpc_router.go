// SPDX-License-Identifier: Apache-2.0
// Copyright Pionix GmbH and Contributors to EVerest

package usecase_router

import (
	"context"

	"github.com/enbility/eebus-grpc/rpc_services/usecases/cs/lpc"
)

// CSLPCRouter implements the ControllableSystemLPCControlServer gRPC interface
// and routes calls to the correct instance based on instance_id.
type CSLPCRouter struct {
	lpc.UnimplementedControllableSystemLPCControlServer
	registry *Registry
}

func NewCSLPCRouter(registry *Registry) *CSLPCRouter {
	return &CSLPCRouter{registry: registry}
}

func (r *CSLPCRouter) ConsumptionLimit(ctx context.Context, req *lpc.ConsumptionLimitRequest) (*lpc.ConsumptionLimitResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.ConsumptionLimit(ctx, req)
}

func (r *CSLPCRouter) SetConsumptionLimit(ctx context.Context, req *lpc.SetConsumptionLimitRequest) (*lpc.SetConsumptionLimitResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.SetConsumptionLimit(ctx, req)
}

func (r *CSLPCRouter) PendingConsumptionLimit(ctx context.Context, req *lpc.PendingConsumptionLimitRequest) (*lpc.PendingConsumptionLimitResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.PendingConsumptionLimit(ctx, req)
}

func (r *CSLPCRouter) ApproveOrDenyConsumptionLimit(ctx context.Context, req *lpc.ApproveOrDenyConsumptionLimitRequest) (*lpc.ApproveOrDenyConsumptionLimitResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.ApproveOrDenyConsumptionLimit(ctx, req)
}

func (r *CSLPCRouter) FailsafeConsumptionActivePowerLimit(ctx context.Context, req *lpc.FailsafeConsumptionActivePowerLimitRequest) (*lpc.FailsafeConsumptionActivePowerLimitResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.FailsafeConsumptionActivePowerLimit(ctx, req)
}

func (r *CSLPCRouter) SetFailsafeConsumptionActivePowerLimit(ctx context.Context, req *lpc.SetFailsafeConsumptionActivePowerLimitRequest) (*lpc.SetFailsafeConsumptionActivePowerLimitResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.SetFailsafeConsumptionActivePowerLimit(ctx, req)
}

func (r *CSLPCRouter) FailsafeDurationMinimum(ctx context.Context, req *lpc.FailsafeDurationMinimumRequest) (*lpc.FailsafeDurationMinimumResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.FailsafeDurationMinimum(ctx, req)
}

func (r *CSLPCRouter) SetFailsafeDurationMinimum(ctx context.Context, req *lpc.SetFailsafeDurationMinimumRequest) (*lpc.SetFailsafeDurationMinimumResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.SetFailsafeDurationMinimum(ctx, req)
}

func (r *CSLPCRouter) StartHeartbeat(ctx context.Context, req *lpc.StartHeartbeatRequest) (*lpc.StartHeartbeatResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.StartHeartbeat(ctx, req)
}

func (r *CSLPCRouter) StopHeartbeat(ctx context.Context, req *lpc.StopHeartbeatRequest) (*lpc.StopHeartbeatResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.StopHeartbeat(ctx, req)
}

func (r *CSLPCRouter) IsHeartbeatWithinDuration(ctx context.Context, req *lpc.IsHeartbeatWithinDurationRequest) (*lpc.IsHeartbeatWithinDurationResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.IsHeartbeatWithinDuration(ctx, req)
}

func (r *CSLPCRouter) ConsumptionNominalMax(ctx context.Context, req *lpc.ConsumptionNominalMaxRequest) (*lpc.ConsumptionNominalMaxResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.ConsumptionNominalMax(ctx, req)
}

func (r *CSLPCRouter) SetConsumptionNominalMax(ctx context.Context, req *lpc.SetConsumptionNominalMaxRequest) (*lpc.SetConsumptionNominalMaxResponse, error) {
	srv, err := r.registry.GetCSLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.SetConsumptionNominalMax(ctx, req)
}
