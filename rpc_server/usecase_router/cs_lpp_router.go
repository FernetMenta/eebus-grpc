// SPDX-License-Identifier: Apache-2.0
// Copyright Pionix GmbH and Contributors to EVerest

package usecase_router

import (
	"context"

	"github.com/enbility/eebus-grpc/rpc_services/usecases/cs/lpp"
)

// CSLPPRouter implements the ControllableSystemLPPControlServer gRPC interface
// and routes calls to the correct instance based on instance_id.
type CSLPPRouter struct {
	lpp.UnimplementedControllableSystemLPPControlServer
	registry *Registry
}

func NewCSLPPRouter(registry *Registry) *CSLPPRouter {
	return &CSLPPRouter{registry: registry}
}

func (r *CSLPPRouter) ProductionLimit(ctx context.Context, req *lpp.ProductionLimitRequest) (*lpp.ProductionLimitResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.ProductionLimit(ctx, req)
}

func (r *CSLPPRouter) SetProductionLimit(ctx context.Context, req *lpp.SetProductionLimitRequest) (*lpp.SetProductionLimitResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.SetProductionLimit(ctx, req)
}

func (r *CSLPPRouter) PendingProductionLimit(ctx context.Context, req *lpp.PendingProductionLimitRequest) (*lpp.PendingProductionLimitResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.PendingProductionLimit(ctx, req)
}

func (r *CSLPPRouter) ApproveOrDenyProductionLimit(ctx context.Context, req *lpp.ApproveOrDenyProductionLimitRequest) (*lpp.ApproveOrDenyProductionLimitResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.ApproveOrDenyProductionLimit(ctx, req)
}

func (r *CSLPPRouter) FailsafeProductionActivePowerLimit(ctx context.Context, req *lpp.FailsafeProductionActivePowerLimitRequest) (*lpp.FailsafeProductionActivePowerLimitResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.FailsafeProductionActivePowerLimit(ctx, req)
}

func (r *CSLPPRouter) SetFailsafeProductionActivePowerLimit(ctx context.Context, req *lpp.SetFailsafeProductionActivePowerLimitRequest) (*lpp.SetFailsafeProductionActivePowerLimitResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.SetFailsafeProductionActivePowerLimit(ctx, req)
}

func (r *CSLPPRouter) FailsafeDurationMinimum(ctx context.Context, req *lpp.FailsafeDurationMinimumRequest) (*lpp.FailsafeDurationMinimumResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.FailsafeDurationMinimum(ctx, req)
}

func (r *CSLPPRouter) SetFailsafeDurationMinimum(ctx context.Context, req *lpp.SetFailsafeDurationMinimumRequest) (*lpp.SetFailsafeDurationMinimumResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.SetFailsafeDurationMinimum(ctx, req)
}

func (r *CSLPPRouter) StartHeartbeat(ctx context.Context, req *lpp.StartHeartbeatRequest) (*lpp.StartHeartbeatResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.StartHeartbeat(ctx, req)
}

func (r *CSLPPRouter) StopHeartbeat(ctx context.Context, req *lpp.StopHeartbeatRequest) (*lpp.StopHeartbeatResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.StopHeartbeat(ctx, req)
}

func (r *CSLPPRouter) IsHeartbeatWithinDuration(ctx context.Context, req *lpp.IsHeartbeatWithinDurationRequest) (*lpp.IsHeartbeatWithinDurationResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.IsHeartbeatWithinDuration(ctx, req)
}

func (r *CSLPPRouter) ProductionNominalMax(ctx context.Context, req *lpp.ProductionNominalMaxRequest) (*lpp.ProductionNominalMaxResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.ProductionNominalMax(ctx, req)
}

func (r *CSLPPRouter) SetProductionNominalMax(ctx context.Context, req *lpp.SetProductionNominalMaxRequest) (*lpp.SetProductionNominalMaxResponse, error) {
	srv, err := r.registry.GetCSLPP(req.GetInstanceId())
	if err != nil {
		return nil, err
	}
	return srv.SetProductionNominalMax(ctx, req)
}
