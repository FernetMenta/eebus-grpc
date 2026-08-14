#!/bin/sh
set -e

# GRPC_ENDPOINT (host:port) takes precedence over IPV4_ADDR + GRPC_PORT
if [ -n "${GRPC_ENDPOINT}" ]; then
    ADDR="${GRPC_ENDPOINT%%:*}"
    PORT="${GRPC_ENDPOINT##*:}"
else
    ADDR="${IPV4_ADDR:-0.0.0.0}"
    PORT="${GRPC_PORT:-50051}"
fi

exec /usr/local/bin/eebus-grpc \
    -certificate-path "${CRT_PATH}" \
    -private-key-path "${KEY_PATH}" \
    -port "${PORT}" \
    -ipv4Addr "${ADDR}" \
    -log-level "${LOG_LEVEL:-info}" \
    "$@"

