// SPDX-License-Identifier: Apache-2.0
// Copyright Pionix GmbH and Contributors to EVerest

package rpc_server

type RPCServer interface {
	Start(ipv4Addr string, port *int) (int, error)
	Stop() error
}
