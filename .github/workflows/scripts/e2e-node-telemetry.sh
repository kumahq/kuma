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
  local out
  out=$(docker exec "$node" sh -c '
    for p in /proc/[0-9]*; do
      [ -r "$p/cmdline" ] || continue
      cmd=$(tr "\0" " " < "$p/cmdline" 2>/dev/null)
      case "$cmd" in
        *kumactl*|*"kuma-dp run"*|*"/usr/bin/envoy"*) : ;;
        *) continue ;;
      esac
      pid=${p#/proc/}
      echo "--- pid=$pid cmd=$cmd"
      grep -E "^(State|VmRSS|Threads|voluntary_ctxt_switches|nonvoluntary_ctxt_switches):" "$p/status" 2>/dev/null
      printf "wchan="; cat "$p/wchan" 2>/dev/null; echo
      echo "stack:"; cat "$p/stack" 2>/dev/null
    done
  ' 2>/dev/null)
  if [ -n "$out" ]; then
    { echo "=== $(ts) node=${node} ==="; echo "$out"; } >> "${OUT}/node-${node}-watched-procs.log"
  fi
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
