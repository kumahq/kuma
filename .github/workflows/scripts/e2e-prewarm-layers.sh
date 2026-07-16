#!/usr/bin/env bash
#
# Warm imported node image layers into the page cache to avoid the cold binary page-in stall during the concurrent pod-startup wave (kumahq/kuma#17426).
# Immutable layers are warmed at most once per node so there is no steady-state I/O; it polls because k3d/kind import images incrementally as the cluster starts. Best effort, must never fail the job.

set -u

INTERVAL="${E2E_PREWARM_INTERVAL:-10}"
case "${INTERVAL}" in
  ''|*[!0-9]*|0) INTERVAL=10 ;;
esac

# Stop polling after this many consecutive intervals with nothing new to warm,
# so the background loop doesn't keep scanning the snapshot trees for the
# rest of the e2e step once the cluster is fully warmed.
IDLE_LIMIT=6

SEEN_DIR="$(mktemp -d)" || exit 0
[ -d "${SEEN_DIR}" ] || exit 0

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
  echo warmed
}

idle=0
while true; do
  warmed_any=""
  for node in $(node_containers); do
    [ -z "$(warm_node "${node}")" ] || warmed_any=1
  done
  if [ -n "${warmed_any}" ]; then
    idle=0
  else
    idle=$((idle + 1))
    [ "${idle}" -lt "${IDLE_LIMIT}" ] || exit 0
  fi
  sleep "${INTERVAL}"
done
