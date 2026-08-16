// SPDX-License-Identifier: Apache-2.0
// Copyright Pionix GmbH and Contributors to EVerest

package usecase_router

import (
	"context"

	"github.com/enbility/eebus-grpc/rpc_services/usecases/eg/lpp"
)

// EGLPPRouter implements the EnergyGuardLPPControlServer gRPC interface
// and routes calls to the correct instance based on instance_id.
type EGLPPRouter struct {
	lpp.UnimplementedEnergyGuardLPPControlServer
	registry *Registry
}

func NewEGLPPRouter(registry *Registry) *EGLPPRouter {
	return &EGLPPRouter{registry: registry}
}

func (r *EGLPPRouter) ProductionLimit(ctx context.Context, req *lpp.ProductionLimitRequest) (*lpp.ProductionLimitResponse, error) {
	srv, err := r.registry.GetEGLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.ProductionLimit(ctx, req)
}

func (r *EGLPPRouter) WriteProductionLimit(ctx context.Context, req *lpp.WriteProductionLimitRequest) (*lpp.WriteProductionLimitResponse, error) {
	srv, err := r.registry.GetEGLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.WriteProductionLimit(ctx, req)
}

func (r *EGLPPRouter) FailsafeProductionActivePowerLimit(ctx context.Context, req *lpp.FailsafeProductionActivePowerLimitRequest) (*lpp.FailsafeProductionActivePowerLimitResponse, error) {
	srv, err := r.registry.GetEGLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.FailsafeProductionActivePowerLimit(ctx, req)
}

func (r *EGLPPRouter) WriteFailsafeProductionActivePowerLimit(ctx context.Context, req *lpp.WriteFailsafeProductionActivePowerLimitRequest) (*lpp.WriteFailsafeProductionActivePowerLimitResponse, error) {
	srv, err := r.registry.GetEGLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.WriteFailsafeProductionActivePowerLimit(ctx, req)
}

func (r *EGLPPRouter) FailsafeDurationMinimum(ctx context.Context, req *lpp.FailsafeDurationMinimumRequest) (*lpp.FailsafeDurationMinimumResponse, error) {
	srv, err := r.registry.GetEGLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.FailsafeDurationMinimum(ctx, req)
}

func (r *EGLPPRouter) WriteFailsafeDurationMinimum(ctx context.Context, req *lpp.WriteFailsafeDurationMinimumRequest) (*lpp.WriteFailsafeDurationMinimumResponse, error) {
	srv, err := r.registry.GetEGLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.WriteFailsafeDurationMinimum(ctx, req)
}

func (r *EGLPPRouter) StartHeartbeat(ctx context.Context, req *lpp.StartHeartbeatRequest) (*lpp.StartHeartbeatResponse, error) {
	srv, err := r.registry.GetEGLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.StartHeartbeat(ctx, req)
}

func (r *EGLPPRouter) StopHeartbeat(ctx context.Context, req *lpp.StopHeartbeatRequest) (*lpp.StopHeartbeatResponse, error) {
	srv, err := r.registry.GetEGLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.StopHeartbeat(ctx, req)
}

func (r *EGLPPRouter) IsHeartbeatWithinDuration(ctx context.Context, req *lpp.IsHeartbeatWithinDurationRequest) (*lpp.IsHeartbeatWithinDurationResponse, error) {
	srv, err := r.registry.GetEGLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.IsHeartbeatWithinDuration(ctx, req)
}

func (r *EGLPPRouter) ProductionNominalMax(ctx context.Context, req *lpp.ProductionNominalMaxRequest) (*lpp.ProductionNominalMaxResponse, error) {
	srv, err := r.registry.GetEGLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.ProductionNominalMax(ctx, req)
}
