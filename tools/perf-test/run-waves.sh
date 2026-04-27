#!/bin/bash
# run-waves.sh — run a progressive load ramp (Phase 3) and collect profiles per wave
#
# Deploys waves 1..N incrementally, waits for stabilization after each, and captures
# pprof profiles + metrics snapshots from one or more CPs.
#
# Env:
#   KUBE_CONTEXT     — kubectl context (required if not default)
#   OUT_DIR          — profile output directory (default: <script dir>/profiles)
#   ENDPOINTS        — space-separated "<name>=<url>" pairs (e.g.
#                      "zone=http://localhost:5680 global=http://34.10.179.32:5680")
#                      Defaults to "cp=http://localhost:5680".
#   STABILIZE_SECS   — seconds to wait after each wave before profiling (default: 300)
#   WAVES            — override the default wave definition. Space-separated entries
#                      of "wave_num:num_ns:svc_per_ns:replicas".
#                      Default: "1:5:5:2 2:10:5:2 3:20:5:3 4:30:10:3 5:20:10:5"
#
# Example:
#   KUBE_CONTEXT=gke_my-project_us-central1_perf \
#   ENDPOINTS="zone=http://localhost:5680 global=http://34.10.179.32:5680" \
#     ./run-waves.sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT_DIR=${OUT_DIR:-"$SCRIPT_DIR/profiles"}
STABILIZE_SECS=${STABILIZE_SECS:-300}
ENDPOINTS=${ENDPOINTS:-"cp=http://localhost:5680"}
WAVES=${WAVES:-"1:5:5:2 2:10:5:2 3:20:5:3 4:30:10:3 5:20:10:5"}

export KUBE_CONTEXT OUT_DIR ENDPOINTS

# Sanity check: every endpoint must respond
for pair in $ENDPOINTS; do
  name=${pair%%=*}
  ep=${pair#*=}
  curl -sf --connect-timeout 3 "$ep/debug/pprof/" > /dev/null || {
    echo "ERROR: $name ($ep) is not reachable" >&2
    exit 1
  }
done

mkdir -p "$OUT_DIR"

echo "=== Phase 3: Progressive Load Ramp ==="
echo "Waves: $WAVES"
echo "Endpoints: $ENDPOINTS"
echo ""

for wave_entry in $WAVES; do
  IFS=':' read -r wave ns_count svc_count replicas <<< "$wave_entry"

  echo ""
  echo "--- Wave $wave ($ns_count ns × $svc_count svc × $replicas replicas) ---"
  bash "$SCRIPT_DIR/deploy-wave.sh" "$wave" "$ns_count" "$svc_count" "$replicas"

  echo "Stabilizing ${STABILIZE_SECS}s..."
  sleep "$STABILIZE_SECS"

  bash "$SCRIPT_DIR/collect-profiles.sh" "wave${wave}"
done

echo ""
echo "=== All waves complete. Profiles in $OUT_DIR ==="
