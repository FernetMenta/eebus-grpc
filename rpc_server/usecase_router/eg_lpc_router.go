// SPDX-License-Identifier: Apache-2.0
// Copyright Pionix GmbH and Contributors to EVerest

package usecase_router

import (
	"context"

	"github.com/enbility/eebus-grpc/rpc_services/usecases/eg/lpc"
)

// EGLPCRouter implements the EnergyGuardLPCControlServer gRPC interface
// and routes calls to the correct instance based on instance_id.
type EGLPCRouter struct {
	lpc.UnimplementedEnergyGuardLPCControlServer
	registry *Registry
}

func NewEGLPCRouter(registry *Registry) *EGLPCRouter {
	return &EGLPCRouter{registry: registry}
}

func (r *EGLPCRouter) ConsumptionLimit(ctx context.Context, req *lpc.ConsumptionLimitRequest) (*lpc.ConsumptionLimitResponse, error) {
	srv, err := r.registry.GetEGLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.ConsumptionLimit(ctx, req)
}

func (r *EGLPCRouter) WriteConsumptionLimit(ctx context.Context, req *lpc.WriteConsumptionLimitRequest) (*lpc.WriteConsumptionLimitResponse, error) {
	srv, err := r.registry.GetEGLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.WriteConsumptionLimit(ctx, req)
}

func (r *EGLPCRouter) FailsafeConsumptionActivePowerLimit(ctx context.Context, req *lpc.FailsafeConsumptionActivePowerLimitRequest) (*lpc.FailsafeConsumptionActivePowerLimitResponse, error) {
	srv, err := r.registry.GetEGLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.FailsafeConsumptionActivePowerLimit(ctx, req)
}

func (r *EGLPCRouter) WriteFailsafeConsumptionActivePowerLimit(ctx context.Context, req *lpc.WriteFailsafeConsumptionActivePowerLimitRequest) (*lpc.WriteFailsafeConsumptionActivePowerLimitResponse, error) {
	srv, err := r.registry.GetEGLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.WriteFailsafeConsumptionActivePowerLimit(ctx, req)
}

func (r *EGLPCRouter) FailsafeDurationMinimum(ctx context.Context, req *lpc.FailsafeDurationMinimumRequest) (*lpc.FailsafeDurationMinimumResponse, error) {
	srv, err := r.registry.GetEGLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.FailsafeDurationMinimum(ctx, req)
}

func (r *EGLPCRouter) WriteFailsafeDurationMinimum(ctx context.Context, req *lpc.WriteFailsafeDurationMinimumRequest) (*lpc.WriteFailsafeDurationMinimumResponse, error) {
	srv, err := r.registry.GetEGLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.WriteFailsafeDurationMinimum(ctx, req)
}

func (r *EGLPCRouter) StartHeartbeat(ctx context.Context, req *lpc.StartHeartbeatRequest) (*lpc.StartHeartbeatResponse, error) {
	srv, err := r.registry.GetEGLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.StartHeartbeat(ctx, req)
}

func (r *EGLPCRouter) StopHeartbeat(ctx context.Context, req *lpc.StopHeartbeatRequest) (*lpc.StopHeartbeatResponse, error) {
	srv, err := r.registry.GetEGLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.StopHeartbeat(ctx, req)
}

func (r *EGLPCRouter) IsHeartbeatWithinDuration(ctx context.Context, req *lpc.IsHeartbeatWithinDurationRequest) (*lpc.IsHeartbeatWithinDurationResponse, error) {
	srv, err := r.registry.GetEGLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.IsHeartbeatWithinDuration(ctx, req)
}

func (r *EGLPCRouter) ConsumptionNominalMax(ctx context.Context, req *lpc.ConsumptionNominalMaxRequest) (*lpc.ConsumptionNominalMaxResponse, error) {
	srv, err := r.registry.GetEGLPC(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.ConsumptionNominalMax(ctx, req)
}
