#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright Pionix GmbH and Contributors to EVerest

set -e

PROTO_PATH="../protobuf"
OUT="../rpc_services"
mkdir -p "$OUT"

# Common Types
protoc \
    --plugin=protoc-gen-go=$(go tool -n protoc-gen-go) \
    --plugin=protoc-gen-go-grpc=$(go tool -n protoc-gen-go-grpc) \
    --proto_path="$PROTO_PATH" \
    --go_out="$OUT" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$OUT" \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_PATH/common_types/types.proto"

# Control Service
protoc \
    --plugin=protoc-gen-go=$(go tool -n protoc-gen-go) \
    --plugin=protoc-gen-go-grpc=$(go tool -n protoc-gen-go-grpc) \
    --proto_path="$PROTO_PATH" \
    --go_out="$OUT" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$OUT" \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_PATH/control_service/control_service.proto" \
    "$PROTO_PATH/control_service/messages.proto" \
    "$PROTO_PATH/control_service/types.proto"


# EG LPC Service
protoc \
    --plugin=protoc-gen-go=$(go tool -n protoc-gen-go) \
    --plugin=protoc-gen-go-grpc=$(go tool -n protoc-gen-go-grpc) \
    --proto_path="$PROTO_PATH" \
    --go_out="$OUT" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$OUT" \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_PATH/usecases/eg/lpc/service.proto" \
    "$PROTO_PATH/usecases/eg/lpc/messages.proto"

# CS LPC Service
protoc \
    --plugin=protoc-gen-go=$(go tool -n protoc-gen-go) \
    --plugin=protoc-gen-go-grpc=$(go tool -n protoc-gen-go-grpc) \
    --proto_path="$PROTO_PATH" \
    --go_out="$OUT" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$OUT" \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_PATH/usecases/cs/lpc/service.proto" \
    "$PROTO_PATH/usecases/cs/lpc/messages.proto"

# CS LPP Service
protoc \
    --plugin=protoc-gen-go=$(go tool -n protoc-gen-go) \
    --plugin=protoc-gen-go-grpc=$(go tool -n protoc-gen-go-grpc) \
    --proto_path="$PROTO_PATH" \
    --go_out="$OUT" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$OUT" \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_PATH/usecases/cs/lpp/service.proto" \
    "$PROTO_PATH/usecases/cs/lpp/messages.proto"

# EG LPP Service
protoc \
    --plugin=protoc-gen-go=$(go tool -n protoc-gen-go) \
    --plugin=protoc-gen-go-grpc=$(go tool -n protoc-gen-go-grpc) \
    --proto_path="$PROTO_PATH" \
    --go_out="$OUT" \
    --go_opt=paths=source_relative \
    --go-grpc_out="$OUT" \
    --go-grpc_opt=paths=source_relative \
    "$PROTO_PATH/usecases/eg/lpp/service.proto" \
    "$PROTO_PATH/usecases/eg/lpp/messages.proto"
