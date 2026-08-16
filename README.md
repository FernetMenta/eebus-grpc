# EEBUS GRPC

This repository contains the EEBUS GRPC api. It is based on the `enbility/eebus-go` library.
It works as a bridge between a generic EEBUS go application and application code in other languages.
It is designed to be started/stoped and configured via gRPC calls.

## Start

```bash
go run cmd/main.go \
  -certificate-path <path_to_certificate> \
  -private-key-path <path_to_private_key> \
  -port <rpc-port>
```

To start the server, you need to provide the path to the certificate and the private key. A single gRPC server will listen on the specified port and host all services (control service + all use case services).

## Structure

### eebus_service

Implements the `eebus-go/service` interface.

### internal

Contains the generation instructions of the gRPC go glue code generated with `protoc`.

The generation can be done with the following command:

```bash
go generate ./...
```

The generated code is checked in to the repository. So you don't need to run the generation command unless you change the `.proto` files.

### protobuf

Contains the `.proto` files that define the gRPC API. Which is as analogous as possible to the `eebus-go` api.

### rpc_server

Implements the gRPC server interfaces. There is a single gRPC server that hosts:
- The `control_service` responsible for starting, stopping and configuring the `eebus_service`, and for managing use case instances.
- Use case routers (in `usecase_router/`) that dispatch incoming gRPC calls to the correct use case instance based on the `instance_id` field in each request.
- Concrete use case implementations (in `usecase_server/`) that hold the business logic for each actor/use case combination (cs/lpc, cs/lpp, eg/lpc, eg/lpp).

When a use case is added via `AddUseCase`, it returns an `instance_id`. All subsequent calls to that use case's gRPC service must include this `instance_id` so the router can dispatch to the correct instance.

### rpc_services

Contains, the generated gRPC glue code. The code is generated with the `protoc` command.

### utils

Contains utility functions.

## Helper Commands

- [Create Certificate Script](cmd/create_cert/README.md)
- [Dummy CEM](cmd/dummy_cem/README.md)


## Build image

Build and push multi-platform image (amd64 + arm64):

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t fernetmenta/iobroker.eebus-grpc:2.0.0 \
  -t fernetmenta/iobroker.eebus-grpc:latest \
  --push .
```

Note: This requires a buildx builder with the `docker-container` driver (for multi-platform support). If your current builder doesn't support arm64, set one up:

```bash
docker buildx create --name multiarch --use
docker buildx inspect --bootstrap
```

Local testing (single platform, loaded into local docker):

```bash
docker buildx build --platform linux/amd64 -t eebus-grpc:amd64 --load .
docker buildx build --platform linux/arm64 -t eebus-grpc:arm64 --load .
```

## Run commands
Run grpc server (certs will be created automatically in the specified certs directory if not exist!)
```bash
mkdir certs
docker run --rm -it \
  --network=host \
  -v "$PWD/certs:/certs" \
  -e LOG_LEVEL=debug \
  fernetmenta/iobroker.eebus-grpc:latest
```

The entrypoint uses these environment variables (with their Dockerfile defaults):

| Variable | Default | Description |
|----------|---------|-------------|
| `GRPC_ENDPOINT` | *(empty)* | If set (host:port), takes precedence over IPV4_ADDR + GRPC_PORT |
| `IPV4_ADDR` | `0.0.0.0` | gRPC server bind address |
| `CRT_PATH` | `/certs/myhems_cert` | Path to certificate |
| `KEY_PATH` | `/certs/myhems_key` | Path to private key |
| `GRPC_PORT` | `50051` | gRPC server port |
| `LOG_LEVEL` | `info` | Log level: trace, debug, info, error |
## Docker Compose
An example [docker-compose.yml](docker-compose.yml) file is provided. You need to adopt the values in that file to your needs. It shows how to use the eebus-grpc container with another iobroker container. Both containers are using Docker Macvlan driver to be able to communicate directly with other devices in the LAN. The Docker Bridge network is used to for the GRPC communication between the containers.

``` bash
mkdir -p certs
mkdir -p iobroker
docker-compose up -d
```




