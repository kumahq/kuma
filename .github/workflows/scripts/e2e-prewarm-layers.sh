#!/usr/bin/env bash
#
# Background page-cache prewarmer for e2e nodes.
#
# Why: the init-hang flake is cold binary page-in (State-D filemap_fault) off a
# saturated disk during the pod-startup wave. Reading the imported image overlay
# layers keeps them resident in the shared host page cache, so container exec
# faults become cache hits instead of disk reads. Best effort; must never fail
# the job. Prototype for kumahq/kuma#17426.

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
