#!/bin/bash
# collect-profiles.sh — collect pprof profiles and a metrics snapshot from Kuma CP(s)
#
# Usage: collect-profiles.sh <label> [endpoint_url]
#
# Env:
#   OUT_DIR     — base directory for profiles (default: ./profiles)
#   ENDPOINTS   — space-separated list of "<name>=<url>" pairs.
#                 If set, overrides the positional endpoint argument and
#                 collects from every endpoint in parallel. Each profile is
#                 prefixed with its name (e.g. zone-cpu.pb.gz, global-cpu.pb.gz).
#   CPU_SECS    — CPU profile duration (default: 30)
#   TRACE_SECS  — execution trace duration (default: 5)
#
# Examples:
#   # single endpoint
#   collect-profiles.sh wave1 http://localhost:5680
#
#   # multi-CP (zone + global)
#   ENDPOINTS="zone=http://localhost:5680 global=http://34.10.179.32:5680" \
#     collect-profiles.sh wave1
set -e

LABEL=${1:-manual}
SINGLE_EP=${2:-http://localhost:5680}
OUT_DIR=${OUT_DIR:-./profiles}
CPU_SECS=${CPU_SECS:-30}
TRACE_SECS=${TRACE_SECS:-5}

OUT="$OUT_DIR/$LABEL"
mkdir -p "$OUT"

collect_one() {
  local name=$1 ep=$2 prefix=$3
  echo "  [$name] metrics snapshot..."
  curl -sf "$ep/metrics" \
    | grep -E 'xds_generation|kds_delta|store_bucket|go_goroutines|go_memstats_heap|resources_count|process_cpu' \
    > "$OUT/${prefix}metrics.txt"

  echo "  [$name] CPU profile (${CPU_SECS}s)..."
  go tool pprof -proto "$ep/debug/pprof/profile?seconds=${CPU_SECS}" > "$OUT/${prefix}cpu.pb.gz"

  echo "  [$name] heap..."
  go tool pprof -proto "$ep/debug/pprof/heap" > "$OUT/${prefix}heap.pb.gz"

  echo "  [$name] goroutine..."
  go tool pprof -proto "$ep/debug/pprof/goroutine" > "$OUT/${prefix}goroutine.pb.gz"

  echo "  [$name] mutex..."
  go tool pprof -proto "$ep/debug/pprof/mutex" > "$OUT/${prefix}mutex.pb.gz"

  echo "  [$name] block..."
  go tool pprof -proto "$ep/debug/pprof/block" > "$OUT/${prefix}block.pb.gz"

  echo "  [$name] trace (${TRACE_SECS}s)..."
  curl -sf -o "$OUT/${prefix}trace.out" "$ep/debug/pprof/trace?seconds=${TRACE_SECS}"
}

echo "Collecting profiles: $LABEL → $OUT"

if [[ -n "$ENDPOINTS" ]]; then
  pids=()
  for pair in $ENDPOINTS; do
    name=${pair%%=*}
    ep=${pair#*=}
    collect_one "$name" "$ep" "${name}-" &
    pids+=($!)
  done
  wait "${pids[@]}"
else
  collect_one cp "$SINGLE_EP" ""
fi

echo "Done."
ls -lh "$OUT/"
