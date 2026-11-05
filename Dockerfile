# syntax=docker/dockerfile:1.7

# Base images are parameterized so CI can pin/track them (e.g. the weekly
# base-image refresh resolves RUNTIME_BASE to a digest and rebuilds on change).
ARG BUILDER_BASE=golang:1.26-alpine3.24
ARG RUNTIME_BASE=alpine:3.24

FROM --platform=$BUILDPLATFORM ${BUILDER_BASE} AS builder

ARG TARGETOS
ARG TARGETARCH

# Dynamically switch Go compiler version based on go.mod directive
ENV GOTOOLCHAIN=auto

WORKDIR /src

COPY --link cmd/ cmd/
COPY --link internal/ internal/
COPY --link protobuf/ protobuf/
COPY --link rpc_server/ rpc_server/
COPY --link rpc_services/ rpc_services/
COPY --link utils/ utils/
COPY --link eebus_service/ eebus_service/
COPY --link go.mod go.sum ./

RUN apk add --no-cache git

# Mount Go module cache so downloaded modules & toolchains persist across builds
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go install github.com/grpc-ecosystem/grpc-health-probe@v0.4.52 && \
    cp $(go env GOPATH)/bin/${TARGETOS}_${TARGETARCH}/grpc-health-probe /out/grpc-health-probe || \
    cp $(go env GOPATH)/bin/grpc-health-probe /out/grpc-health-probe

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/eebus-grpc \
        ./cmd/main.go

FROM --platform=$TARGETPLATFORM ${RUNTIME_BASE}

# Record the runtime base image reference (and its digest when CI pins one) so
# the weekly refresh workflow can detect when the base image has changed.
ARG RUNTIME_BASE=alpine:3.24
LABEL org.opencontainers.image.base.name="${RUNTIME_BASE}"

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/grpc-health-probe /usr/local/bin/
COPY --from=builder /out/eebus-grpc /usr/local/bin/
COPY --chmod=755 entrypoint.sh /usr/local/bin/

RUN addgroup -S app && \
    adduser -S \
        -G app \
        -h /home/app \
        -s /sbin/nologin \
        app && \
    mkdir -p /certs /app && \
    chown -R app:app /certs /app

USER app

WORKDIR /app

# GRPC Port
EXPOSE 50051

VOLUME /certs

# GRPC_ENDPOINT (host:port) takes precedence over IPV4_ADDR + GRPC_PORT
ENV GRPC_ENDPOINT=""
ENV IPV4_ADDR="0.0.0.0"
ENV CRT_PATH="/certs/myhems_cert"
ENV KEY_PATH="/certs/myhems_key"
ENV GRPC_PORT="50051"
ENV LOG_LEVEL="info"

HEALTHCHECK --interval=60s --timeout=10s --retries=3 --start-period=10s \
    CMD grpc-health-probe -addr=${GRPC_ENDPOINT:-${IPV4_ADDR}:${GRPC_PORT}}

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
