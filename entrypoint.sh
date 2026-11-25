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

/usr/local/bin/eebus-grpc \
    -certificate-path "${CRT_PATH}" \
    -private-key-path "${KEY_PATH}" \
    -port "${PORT}" \
    -ipv4Addr "${ADDR}" \
    -log-level "${LOG_LEVEL:-info}" \
    -eebus-addr "${EEBUS_ADDR:-}" \
    "$@" &
MAIN_PID=$!

# Forward termination signals to the main process
trap 'kill -TERM "$MAIN_PID" 2>/dev/null' TERM INT

WATCHDOG_PID=""

# Watchdog: probes the gRPC health endpoint and kills the main process after
# WATCHDOG_FAILURES consecutive failed probes, so the orchestrator's restart
# policy can restart the container. Disable with WATCHDOG_ENABLED=false.
watchdog() {
    probe_addr="${GRPC_ENDPOINT:-${ADDR}:${PORT}}"
    interval="${WATCHDOG_INTERVAL:-60}"
    timeout="${WATCHDOG_TIMEOUT:-10}"
    max_failures="${WATCHDOG_FAILURES:-3}"
    failures=0

    echo "watchdog: enabled (addr=${probe_addr} interval=${interval}s timeout=${timeout}s failures=${max_failures})"
    sleep "${WATCHDOG_START_PERIOD:-10}"

    while kill -0 "$MAIN_PID" 2>/dev/null; do
        if grpc-health-probe -addr="${probe_addr}" \
            -connect-timeout="${timeout}s" -rpc-timeout="${timeout}s" >/dev/null 2>&1; then
            failures=0
        else
            failures=$((failures + 1))
            echo "watchdog: health probe failed (${failures}/${max_failures})"
            if [ "$failures" -ge "$max_failures" ]; then
                echo "watchdog: killing main process (pid ${MAIN_PID})"
                kill -TERM "$MAIN_PID" 2>/dev/null || true
                sleep 5
                kill -KILL "$MAIN_PID" 2>/dev/null || true
                return 0
            fi
        fi
        sleep "$interval"
    done
}

case "$(echo "${WATCHDOG_ENABLED:-true}" | tr '[:upper:]' '[:lower:]')" in
    false|0|no|off)
        echo "watchdog: disabled"
        ;;
    *)
        watchdog &
        WATCHDOG_PID=$!
        ;;
esac

# wait returns early (>128) when a trapped signal arrives; keep waiting until
# the main process has really exited.
STATUS=0
while kill -0 "$MAIN_PID" 2>/dev/null; do
    set +e
    wait "$MAIN_PID"
    STATUS=$?
    set -e
done

[ -n "$WATCHDOG_PID" ] && kill "$WATCHDOG_PID" 2>/dev/null || true
exit "$STATUS"
