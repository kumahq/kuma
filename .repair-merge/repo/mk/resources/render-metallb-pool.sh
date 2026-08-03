#!/usr/bin/env bash
# Render a MetalLB IPAddressPool + L2Advertisement manifest for a k3d cluster.
#
# Derives a /24 subnet from the Docker network backing the cluster:
#   - first two octets come from the network's IPv4 CIDR
#   - third octet is CLUSTER_NUMBER+1 (avoids the gateway at x.x.0.1)
#
# The network must be at least /16 so that per-cluster /24 subnets fit.
#
# Usage: render-metallb-pool.sh <docker-network> <cluster-number> <output> [namespace]
set -euo pipefail

network="${1:?usage: $0 <docker-network> <cluster-number> <output> [namespace]}"
cluster_number="${2:?usage: $0 <docker-network> <cluster-number> <output> [namespace]}"
output="${3:?usage: $0 <docker-network> <cluster-number> <output> [namespace]}"
namespace="${4:-metallb-system}"

parent_subnet=$(
  docker network inspect "$network" --format json \
    | jq --raw-output '.[0].IPAM.Config[]?.Subnet' \
    | grep --invert-match ':' \
    | head -n 1
)

if [ -z "$parent_subnet" ]; then
  echo "Docker network '$network' has no IPv4 subnet configured" >&2
  exit 1
fi

prefix_len="${parent_subnet##*/}"
if [ "$prefix_len" -gt 16 ]; then
  echo "Docker network '$network' subnet $parent_subnet is narrower than /16; MetalLB needs at least /16" >&2
  exit 1
fi

network_prefix=$(echo "$parent_subnet" | cut -d'.' -f1-2)
subnet="${network_prefix}.$((cluster_number + 1)).0/24"

mkdir -p "$(dirname "$output")"
cat > "$output" <<EOF
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: main
  namespace: ${namespace}
spec:
  addresses:
    - ${subnet}
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: empty
  namespace: ${namespace}
EOF
