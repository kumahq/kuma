#!/usr/bin/env bash
#
# Keep e2e node image layers hot in the page cache: the init-hang flake is cold binary page-in (State-D filemap_fault) off the throughput-capped runner disk during the pod-startup wave (kumahq/kuma#17426), and re-reading the imported overlay layers keeps them resident so container-exec faults become cache hits.
# Re-warms continuously on purpose (NOT once): multizone imports layers into several clusters over time and pages can be evicted over a ~20min run, so a single pass leaves later waves cold. Best effort, must never fail the job.

INTERVAL="${E2E_PREWARM_INTERVAL:-30}"

node_containers() {
  docker ps --format '{{.Names}}' 2>/dev/null \
    | grep -E '(-server-[0-9]+|-agent-[0-9]+|-control-plane$|-worker[0-9]*$)' \
    | grep -vE 'serverlb|registry'
}

warm() {
  docker exec "$1" sh -c '
    for d in /var/lib/rancher/k3s/agent/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots \
             /var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots; do
      [ -d "$d" ] && find "$d" -type f -print0 2>/dev/null | xargs -0 -r cat >/dev/null 2>&1
    done
  ' 2>/dev/null
}

while true; do
  for n in $(node_containers); do warm "$n"; done
  sleep "$INTERVAL"
done
