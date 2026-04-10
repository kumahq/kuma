#!/bin/bash
# policy-scenarios.sh — Phase 4 policy/KDS load scenarios
#
# Runs these scenarios against an already-deployed DP fleet:
#   baseline          — no user policies, snapshots CPU + metrics
#   spread            — N policies targeting N different Dataplane label selectors
#   single-dp         — N policies all targeting the same Dataplane label selector
#   spread-update     — update every spread policy in a tight loop
#   single-dp-update  — update every single-dp policy in a tight loop
#   delete            — bulk delete all user-origin MeshTimeout policies
#
# Assumes:
#   - a zone CP with wave1-style fake-service deployments already running
#   - DP pods have an `app=svc-<N>` label (from deploy-wave.sh)
#   - mesh has MeshService.Mode=Exclusive and MeshIdentity configured
#   - CP has KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED=true
#
# Env:
#   KUBE_CONTEXT   — kubectl context (required if not default)
#   OUT_DIR        — profile output dir (default: <script dir>/profiles)
#   ENDPOINTS      — space-separated "<name>=<url>" pairs (same format as run-waves.sh)
#   N              — policies per scenario (default: 50)
#   APP_LABELS     — distinct app label values for the SPREAD scenario
#                    (default: "svc-1 svc-2 svc-3 svc-4 svc-5 svc-6 svc-7 svc-8 svc-9 svc-10")
#   SINGLE_APP     — target label for SINGLE-DP scenario (default: svc-1)
#   CPU_SECS       — CPU profile duration (default: 60)
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT_DIR=${OUT_DIR:-"$SCRIPT_DIR/profiles"}
ENDPOINTS=${ENDPOINTS:-"cp=http://localhost:5680"}
N=${N:-50}
APP_LABELS=${APP_LABELS:-"svc-1 svc-2 svc-3 svc-4 svc-5 svc-6 svc-7 svc-8 svc-9 svc-10"}
SINGLE_APP=${SINGLE_APP:-svc-1}
CPU_SECS=${CPU_SECS:-60}

CTX_ARG=()
[[ -n "$KUBE_CONTEXT" ]] && CTX_ARG=(--context="$KUBE_CONTEXT")

COLLECT_CPU_PIDS=()

collect_cpu_bg() {
  local label=$1 secs=$2
  mkdir -p "$OUT_DIR/$label"
  COLLECT_CPU_PIDS=()
  for pair in $ENDPOINTS; do
    local name=${pair%%=*}
    local ep=${pair#*=}
    go tool pprof -proto "$ep/debug/pprof/profile?seconds=${secs}" \
      > "$OUT_DIR/$label/${name}-cpu.pb.gz" &
    COLLECT_CPU_PIDS+=($!)
  done
}

wait_cpu() {
  wait "${COLLECT_CPU_PIDS[@]}"
}

collect_heap_metrics() {
  local label=$1
  mkdir -p "$OUT_DIR/$label"
  for pair in $ENDPOINTS; do
    local name=${pair%%=*}
    local ep=${pair#*=}
    curl -sf "$ep/metrics" \
      | grep -E 'kds_delta|resources_count|xds_generation|go_goroutines|go_memstats_heap' \
      > "$OUT_DIR/$label/${name}-metrics.txt" || true
    go tool pprof -proto "$ep/debug/pprof/heap" > "$OUT_DIR/$label/${name}-heap.pb.gz"
  done
  echo "  collected $label"
}

apply_policy() {
  local name=$1 app=$2 timeout=$3
  kubectl "${CTX_ARG[@]}" apply -f - <<EOF
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: ${name}
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
    kuma.io/origin: zone
spec:
  targetRef:
    kind: Dataplane
    labels:
      app: "${app}"
  to:
  - targetRef:
      kind: Mesh
    default:
      http:
        requestTimeout: ${timeout}s
EOF
}

# ── Baseline ─────────────────────────────────────────────────────────────────
echo "=== Baseline ==="
collect_cpu_bg p4-baseline 30
wait_cpu
collect_heap_metrics p4-baseline

# ── Scenario A: SPREAD ───────────────────────────────────────────────────────
echo ""
echo "=== Scenario A: SPREAD — $N policies across distinct Dataplane label selectors ==="
collect_cpu_bg p4-spread-create "$CPU_SECS"

read -ra apps <<< "$APP_LABELS"
n_apps=${#apps[@]}
for i in $(seq 1 "$N"); do
  app=${apps[$(( (i - 1) % n_apps ))]}
  apply_policy "spread-${i}" "$app" "$((i % 10 + 1))"
done
wait_cpu
sleep 30
collect_heap_metrics p4-spread-create

# ── Scenario B: SINGLE DP ────────────────────────────────────────────────────
echo ""
echo "=== Scenario B: SINGLE DP — $N policies → app=$SINGLE_APP ==="
collect_cpu_bg p4-single-create "$CPU_SECS"
for i in $(seq 1 "$N"); do
  apply_policy "single-${i}" "$SINGLE_APP" "$((i % 10 + 1))"
done
wait_cpu
sleep 30
collect_heap_metrics p4-single-create

# ── Updates: spread ──────────────────────────────────────────────────────────
echo ""
echo "=== Updates: spread ==="
collect_cpu_bg p4-spread-update "$CPU_SECS"
for i in $(seq 1 "$N"); do
  kubectl "${CTX_ARG[@]}" -n kuma-system patch meshtimeout "spread-${i}" \
    --type merge -p '{"spec":{"to":[{"targetRef":{"kind":"Mesh"},"default":{"http":{"requestTimeout":"99s"}}}]}}'
done
wait_cpu
collect_heap_metrics p4-spread-update

# ── Updates: single-dp ───────────────────────────────────────────────────────
echo ""
echo "=== Updates: single-dp ==="
collect_cpu_bg p4-single-update "$CPU_SECS"
for i in $(seq 1 "$N"); do
  kubectl "${CTX_ARG[@]}" -n kuma-system patch meshtimeout "single-${i}" \
    --type merge -p '{"spec":{"to":[{"targetRef":{"kind":"Mesh"},"default":{"http":{"requestTimeout":"99s"}}}]}}'
done
wait_cpu
collect_heap_metrics p4-single-update

# ── Bulk delete ──────────────────────────────────────────────────────────────
echo ""
echo "=== Bulk delete ==="
collect_cpu_bg p4-delete "$CPU_SECS"
kubectl "${CTX_ARG[@]}" -n kuma-system delete meshtimeout -l kuma.io/origin=zone 2>&1 || true
wait_cpu
collect_heap_metrics p4-delete

echo ""
echo "=== Phase 4 complete. Profiles in $OUT_DIR ==="
for d in p4-baseline p4-spread-create p4-single-create p4-spread-update p4-single-update p4-delete; do
  [[ -d "$OUT_DIR/$d" ]] && echo "  $d: $(du -sh "$OUT_DIR/$d" | cut -f1)"
done
