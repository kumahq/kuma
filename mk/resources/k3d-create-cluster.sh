#!/usr/bin/env bash
# Create a k3d cluster while holding a lock around host-port prefix allocation.
#
# The lock must stay held through `k3d cluster create`, otherwise parallel
# starts can observe the same free 3XX prefix before Docker binds the ports.
#
# Usage:
#   k3d-create-cluster.sh <docker-network> <lock-dir> -- <k3d cluster create ...>
set -euo pipefail

network="${1:?usage: $0 <docker-network> <lock-dir> -- <k3d cluster create ...>}"
lock_dir="${2:?usage: $0 <docker-network> <lock-dir> -- <k3d cluster create ...>}"
separator="${3:-}"

if [ "$separator" != "--" ]; then
  echo "usage: $0 <docker-network> <lock-dir> -- <k3d cluster create ...>" >&2
  exit 1
fi

shift 3

if [ "$#" -eq 0 ]; then
  echo "usage: $0 <docker-network> <lock-dir> -- <k3d cluster create ...>" >&2
  exit 1
fi

lock_timeout="${K3D_LOCK_TIMEOUT:-120}"

mkdir -p "$(dirname "$lock_dir")"
waited=0
while ! mkdir "$lock_dir" 2>/dev/null; do
  if [ "$waited" -ge "$lock_timeout" ]; then
    echo "Timed out after ${lock_timeout}s waiting for lock: $lock_dir" >&2
    echo "A previous run may have been killed. Remove the directory manually to unblock." >&2
    exit 1
  fi
  sleep 1
  waited=$((waited + 1))
done
trap 'rmdir "$lock_dir"' EXIT INT TERM

# Scan ALL containers' host-bound ports and extract the 3-digit prefix
# of each (e.g. 30080 -> 300). Docker port output looks like:
#   80/tcp, 0.0.0.0:30080-30099->30080-30099/tcp, [::]:30080->30080/tcp
# We split on commas, extract the host port (before "->"), strip the
# IP binding, and collect prefixes. All done in awk to avoid sed
# character-class portability issues.
used_prefixes=$(
  docker ps --format '{{.Ports}}' \
    | awk -F'[, ]+' '{
        for (i = 1; i <= NF; i++) {
          s = $i
          # skip entries without "->" (no host binding, e.g. "80/tcp")
          idx = index(s, "->")
          if (idx == 0) continue
          # isolate the host side (before "->")
          s = substr(s, 1, idx - 1)
          # strip IP prefix: remove everything up to and including the last ":"
          n = split(s, parts, ":")
          s = parts[n]
          # expand ranges and extract 3-digit prefixes
          if (s ~ /^[0-9]+-[0-9]+$/) {
            split(s, r, "-")
            for (p = r[1]; p <= r[2]; p++) print substr(p, 1, 3)
          } else if (s ~ /^[0-9]+$/) {
            print substr(s, 1, 3)
          }
        }
      }' \
    | sort -u
)

port_prefix=""
for candidate in $(seq 300 399); do
  if ! printf '%s\n' "$used_prefixes" | grep -qx "$candidate"; then
    port_prefix="$candidate"
    break
  fi
done

if [ -z "$port_prefix" ]; then
  echo "No free 3XX host port prefix on network '${network}'" >&2
  exit 1
fi

"$@" --port "${port_prefix}80-${port_prefix}99:30080-30099@server:0"
