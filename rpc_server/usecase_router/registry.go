// SPDX-License-Identifier: Apache-2.0
// Copyright Pionix GmbH and Contributors to EVerest

package usecase_router

import (
	"fmt"
	"sync"

	"github.com/enbility/eebus-grpc/rpc_server/usecase_server"
	cs_lpc_server "github.com/enbility/eebus-grpc/rpc_server/usecase_server/cs/lpc"
	cs_lpp_server "github.com/enbility/eebus-grpc/rpc_server/usecase_server/cs/lpp"
	eg_lpc_server "github.com/enbility/eebus-grpc/rpc_server/usecase_server/eg/lpc"
	eg_lpp_server "github.com/enbility/eebus-grpc/rpc_server/usecase_server/eg/lpp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Registry holds all use case instances keyed by their instance ID.
// It is safe for concurrent use.
type Registry struct {
	mu        sync.RWMutex
	instances map[string]usecase_server.UseCaseServer

	// Typed references for gRPC routing.
	csLPC map[string]*cs_lpc_server.Server
	csLPP map[string]*cs_lpp_server.Server
	egLPC map[string]*eg_lpc_server.Server
	egLPP map[string]*eg_lpp_server.Server
}

func NewRegistry() *Registry {
	return &Registry{
		instances: make(map[string]usecase_server.UseCaseServer),
		csLPC:     make(map[string]*cs_lpc_server.Server),
		csLPP:     make(map[string]*cs_lpp_server.Server),
		egLPC:     make(map[string]*eg_lpc_server.Server),
		egLPP:     make(map[string]*eg_lpp_server.Server),
	}
}

func (r *Registry) RegisterCSLPC(id string, server *cs_lpc_server.Server) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.instances[id] = server
	r.csLPC[id] = server
}

func (r *Registry) RegisterCSLPP(id string, server *cs_lpp_server.Server) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.instances[id] = server
	r.csLPP[id] = server
}

func (r *Registry) RegisterEGLPC(id string, server *eg_lpc_server.Server) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.instances[id] = server
	r.egLPC[id] = server
}

func (r *Registry) RegisterEGLPP(id string, server *eg_lpp_server.Server) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.instances[id] = server
	r.egLPP[id] = server
}

func (r *Registry) Get(id string) (usecase_server.UseCaseServer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	srv, ok := r.instances[id]
	return srv, ok
}

func (r *Registry) GetCSLPC(id string) (*cs_lpc_server.Server, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	srv, ok := r.csLPC[id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no CS LPC instance with id %q", id)
	}
	return srv, nil
}

func (r *Registry) GetCSLPP(id string) (*cs_lpp_server.Server, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	srv, ok := r.csLPP[id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no CS LPP instance with id %q", id)
	}
	return srv, nil
}

func (r *Registry) GetEGLPC(id string) (*eg_lpc_server.Server, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	srv, ok := r.egLPC[id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no EG LPC instance with id %q", id)
	}
	return srv, nil
}

func (r *Registry) GetEGLPP(id string) (*eg_lpp_server.Server, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	srv, ok := r.egLPP[id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no EG LPP instance with id %q", id)
	}
	return srv, nil
}

// ShutdownAll shuts down all registered use case instances and clears the registry.
func (r *Registry) ShutdownAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, srv := range r.instances {
		srv.Shutdown()
	}
	r.instances = make(map[string]usecase_server.UseCaseServer)
	r.csLPC = make(map[string]*cs_lpc_server.Server)
	r.csLPP = make(map[string]*cs_lpp_server.Server)
	r.egLPC = make(map[string]*eg_lpc_server.Server)
	r.egLPP = make(map[string]*eg_lpp_server.Server)
}

// Remove removes and shuts down a single instance.
func (r *Registry) Remove(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	srv, ok := r.instances[id]
	if !ok {
		return fmt.Errorf("instance %q not found", id)
	}
	srv.Shutdown()
	delete(r.instances, id)
	delete(r.csLPC, id)
	delete(r.csLPP, id)
	delete(r.egLPC, id)
	delete(r.egLPP, id)
	return nil
}
