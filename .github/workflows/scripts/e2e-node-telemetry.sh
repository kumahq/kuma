#!/usr/bin/env bash
#
# Background node-runtime telemetry sampler for e2e jobs.
#
# Why: the "pod stuck initializing" flake self-heals, so by the time a test
# dumps state the evidence is gone. This samples host/node pressure and the
# kernel state of watched init processes while they are stuck, so the next
# occurrence leaves a trace (IO-wait vs CPU starvation vs container-runtime
# stall). It is best effort and must never fail the job.

OUT="${1:?usage: e2e-node-telemetry.sh <output-dir>}"
INTERVAL="${E2E_TELEMETRY_INTERVAL:-3}"
SLOW_EVERY="${E2E_TELEMETRY_SLOW_EVERY:-4}"
WATCH_CONTAINERS_RE='kuma-init|kuma-validation|kuma-sidecar'

ts() { date -u +%Y-%m-%dT%H:%M:%SZ; }

node_containers() {
  docker ps --format '{{.Names}}' 2>/dev/null \
    | grep -E '(-server-[0-9]+|-agent-[0-9]+|-control-plane$|-worker[0-9]*$)' \
    | grep -vE 'serverlb|registry'
}

crictl_exec() {
  local node="$1"
  shift
  docker exec "$node" sh -c "crictl $* 2>/dev/null || k3s crictl $* 2>/dev/null"
}

sample_host() {
  {
    echo "=== $(ts) ==="
    echo "# loadavg"
    cat /proc/loadavg 2>/dev/null
    for resource in cpu io memory; do
      echo "# pressure/${resource}"
      cat "/proc/pressure/${resource}" 2>/dev/null
    done
    echo "# meminfo"
    grep -E '^(MemTotal|MemFree|MemAvailable|Dirty|Writeback):' /proc/meminfo 2>/dev/null
    echo "# disk free"
    df -h / /var/lib/docker 2>/dev/null
  } >> "${OUT}/host.log"
}

sample_docker_stats() {
  {
    echo "=== $(ts) ==="
    docker stats --no-stream \
      --format 'table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.BlockIO}}\t{{.PIDs}}' 2>/dev/null
  } >> "${OUT}/docker-stats.log"
}

sample_watched_procs() {
  local node="$1"
  local ids id pid
  ids=$(crictl_exec "$node" ps -a | grep -E "${WATCH_CONTAINERS_RE}" | awk '{print $1}')
  for id in ${ids}; do
    pid=$(crictl_exec "$node" inspect "$id" | grep -m1 '"pid"' | grep -oE '[0-9]+' | head -1)
    [ -n "${pid}" ] && [ "${pid}" != "0" ] || continue
    {
      echo "=== $(ts) node=${node} ctr=${id} pid=${pid} ==="
      docker exec "$node" sh -c "grep -E '^(Name|State|Threads|VmRSS|voluntary_ctxt_switches|nonvoluntary_ctxt_switches):' /proc/${pid}/status 2>/dev/null"
      echo "# wchan"
      docker exec "$node" sh -c "cat /proc/${pid}/wchan 2>/dev/null; echo"
      echo "# stack"
      docker exec "$node" sh -c "cat /proc/${pid}/stack 2>/dev/null"
    } >> "${OUT}/node-${node}-watched-procs.log"
  done
}

sample_nodes() {
  local node
  for node in $(node_containers); do
    {
      echo "=== $(ts) node=${node} ==="
      crictl_exec "$node" ps -a
    } >> "${OUT}/node-${node}-crictl.log"
    sample_watched_procs "$node"
  done
}

final_capture() {
  local node
  sudo dmesg --ctime 2>/dev/null | tail -500 > "${OUT}/host-dmesg.log" 2>/dev/null
  for node in $(node_containers); do
    docker logs --tail 5000 "$node" > "${OUT}/node-${node}-runtime.log" 2>&1
    crictl_exec "$node" ps -a > "${OUT}/node-${node}-crictl-final.log" 2>/dev/null
    docker exec "$node" sh -c 'dmesg --ctime 2>/dev/null | tail -300' > "${OUT}/node-${node}-dmesg.log" 2>/dev/null
  done
}

stop() {
  final_capture
  exit 0
}
trap stop TERM INT

mkdir -p "${OUT}"
echo "node telemetry sampler started at $(ts), interval=${INTERVAL}s" > "${OUT}/sampler.log"

tick=0
while true; do
  mkdir -p "${OUT}"
  sample_host
  if [ $((tick % SLOW_EVERY)) -eq 0 ]; then
    sample_docker_stats
    sample_nodes
  fi
  tick=$((tick + 1))
  sleep "${INTERVAL}"
done
