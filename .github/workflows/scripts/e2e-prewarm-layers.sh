#!/usr/bin/env bash
#
# Warm imported node image layers into the page cache to avoid the cold binary page-in stall during the concurrent pod-startup wave (kumahq/kuma#17426).
# Immutable layers are warmed at most once per node so there is no steady-state I/O; it polls because k3d/kind import images incrementally as the cluster starts. Best effort, must never fail the job.

set -u

INTERVAL="${E2E_PREWARM_INTERVAL:-10}"
SEEN_DIR="$(mktemp -d)"

cleanup() { rm -rf "${SEEN_DIR}"; }
trap cleanup EXIT
trap 'exit 0' INT TERM

node_containers() {
  docker ps --format '{{.Names}}' 2>/dev/null \
    | grep -E '(-server-[0-9]+|-agent-[0-9]+|-control-plane$|-worker[0-9]*$)' \
    | grep -vE 'serverlb|registry'
}

list_layer_files() {
  docker exec "$1" sh -c '
    for d in /var/lib/rancher/k3s/agent/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots \
             /var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots; do
      [ -d "$d" ] && find "$d" -type f
    done 2>/dev/null
  ' 2>/dev/null | sort
}

warm_node() {
  local node="$1"
  local seen="${SEEN_DIR}/${node}.seen"
  local current new
  current="$(list_layer_files "$node")"
  [ -n "${current}" ] || return 0
  new="$(comm -23 <(printf '%s\n' "${current}") <(sort "${seen}" 2>/dev/null))"
  [ -n "${new}" ] || return 0
  printf '%s\n' "${new}" | docker exec -i "$node" xargs cat >/dev/null 2>&1
  printf '%s\n' "${current}" > "${seen}"
}

while true; do
  for node in $(node_containers); do
    warm_node "${node}"
  done
  sleep "${INTERVAL}"
done
