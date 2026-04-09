#!/bin/bash
# dump-prometheus.sh — pull Kuma CP metrics from Prometheus for offline analysis
#
# Grabs a set of Go runtime, xDS, KDS, and store metrics as Prometheus
# range queries and saves them as JSON files. Assumes Prometheus is reachable
# at $PROM_URL (default: http://localhost:9090 — start a port-forward first).
#
# Env:
#   PROM_URL       — Prometheus base URL (default: http://localhost:9090)
#   OUT_DIR        — output dir (default: ./profiles/prometheus)
#   WINDOW_MINUTES — how far back to query (default: 150)
#   STEP           — range query step in seconds (default: 15)
#   NAMESPACE      — Kuma CP namespace for container/pod metrics (default: kuma-system)
set -e

PROM_URL=${PROM_URL:-http://localhost:9090}
OUT_DIR=${OUT_DIR:-./profiles/prometheus}
WINDOW_MINUTES=${WINDOW_MINUTES:-150}
STEP=${STEP:-15}
NAMESPACE=${NAMESPACE:-kuma-system}

END=$(date +%s)
START=$((END - WINDOW_MINUTES * 60))

mkdir -p "$OUT_DIR"
echo "Window: $START → $END (${WINDOW_MINUTES}min) | step=${STEP}s | prom=$PROM_URL"

# Verify Prometheus is reachable
curl -sf "$PROM_URL/api/v1/query?query=up" > /dev/null || {
  echo "ERROR: Prometheus at $PROM_URL not reachable" >&2
  exit 1
}

CP_METRIC_FILTER="namespace=\"${NAMESPACE}\",container=\"control-plane\""

metrics=(
  # Go runtime
  "go_goroutines"
  "go_memstats_heap_inuse_bytes"
  "go_memstats_heap_alloc_bytes"
  "go_memstats_heap_objects"
  "go_memstats_sys_bytes"
  "go_memstats_stack_inuse_bytes"
  "go_gc_duration_seconds_sum"
  "process_cpu_seconds_total"
  "rate(process_cpu_seconds_total[1m])"

  # xDS
  "xds_generation_count"
  "xds_generation_sum"
  "xds_generation_errors"
  "xds_streams_active"
  "xds_requests_received"
  "xds_responses_sent"

  # KDS
  "kds_delta_generation_count"
  "kds_delta_generation_sum"
  "kds_delta_generation_errors"
  "kds_delta_streams_active"
  "kds_delta_requests_received"
  "kds_delta_responses_sent"

  # Store
  "store_count"
  "store_sum"
  "store_conflicts"
  "resources_count"

  # Histogram quantiles (p50/p99)
  "histogram_quantile(0.50, sum(rate(xds_generation_bucket[1m])) by (le))"
  "histogram_quantile(0.99, sum(rate(xds_generation_bucket[1m])) by (le))"
  "histogram_quantile(0.99, sum(rate(kds_delta_generation_bucket[1m])) by (le))"
  "histogram_quantile(0.99, sum(rate(store_bucket[1m])) by (le, operation, resource_type))"

  # Container-level (correlates with pod restarts / scheduling)
  "container_memory_working_set_bytes{${CP_METRIC_FILTER}}"
  "container_memory_rss{${CP_METRIC_FILTER}}"
  "rate(container_cpu_usage_seconds_total{${CP_METRIC_FILTER}}[1m])"
  "kube_pod_container_status_restarts_total{namespace=\"${NAMESPACE}\"}"
  "kube_pod_container_status_ready{namespace=\"${NAMESPACE}\"}"
)

for metric in "${metrics[@]}"; do
  fname=$(echo "$metric" | tr '(){}=",[]' '_________' | tr -s '_' | tr -d ' ')
  curl -sG "$PROM_URL/api/v1/query_range" \
    --data-urlencode "query=$metric" \
    --data-urlencode "start=$START" \
    --data-urlencode "end=$END" \
    --data-urlencode "step=$STEP" \
    > "$OUT_DIR/${fname}.json"

  n=$(python3 -c "import json; print(len(json.load(open('$OUT_DIR/${fname}.json'))['data']['result']))" 2>/dev/null || echo "err")
  echo "  $metric: $n series"
done

echo ""
count=$(find "$OUT_DIR" -maxdepth 1 -type f | wc -l | tr -d ' ')
echo "Dumped $count files → $OUT_DIR"
du -sh "$OUT_DIR"
