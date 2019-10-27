#!/usr/bin/env sh

set -e

#
# Utility functions
#

function resolve_ip {
  nslookup "${1}" 2>/dev/null | tail -1 | awk '{print $3}'
}

function fail {
  printf 'Error: %s\n' "${1}" >&2  ## Send message to stderr. Exclude >&2 if you don't want it that way.
  exit "${2-1}"                    ## Return a code specified by $2 or 1 by default.
}

#
# Arguments
#

KUMA_CONTROL_PLANE_URL=http://kuma-control-plane:5681

KUMA_EXAMPLE_APP_HOSTNAME=kuma-example-app
KUMA_EXAMPLE_APP_PUBLIC_PORT=8000
KUMA_EXAMPLE_APP_LOCAL_PORT=8000

KUMA_EXAMPLE_CLIENT_HOSTNAME=kuma-example-client
KUMA_EXAMPLE_CLIENT_PUBLIC_PORT=3000
KUMA_EXAMPLE_CLIENT_LOCAL_PORT=3000

#
# Resolve IP address allocated to the `kuma-example-app` container
#

KUMA_EXAMPLE_APP_IP_ADDRESS=$( resolve_ip ${KUMA_EXAMPLE_APP_HOSTNAME} )
if [ -z "${KUMA_EXAMPLE_APP_IP_ADDRESS}" ]; then
  fail "failed to resolve IP address allocated to the '${KUMA_EXAMPLE_APP_HOSTNAME}' container"
fi
echo "'${KUMA_EXAMPLE_APP_HOSTNAME}' has the following IP address: ${KUMA_EXAMPLE_APP_IP_ADDRESS}"

#
# Resolve IP address allocated to the `kuma-example-app` container
#

KUMA_EXAMPLE_CLIENT_IP_ADDRESS=$( resolve_ip ${KUMA_EXAMPLE_CLIENT_HOSTNAME} )
if [ -z "${KUMA_EXAMPLE_CLIENT_IP_ADDRESS}" ]; then
  fail "failed to resolve IP address allocated to the '${KUMA_EXAMPLE_CLIENT_HOSTNAME}' container"
fi
echo "'${KUMA_EXAMPLE_CLIENT_HOSTNAME}' has the following IP address: ${KUMA_EXAMPLE_CLIENT_IP_ADDRESS}"

#
# Configure `kumactl`
#

kumactl config control-planes remove --name universal 2>/dev/null || true # TODO(yskopets): eventually, replace `remove` command with `add --overwrite`
kumactl config control-planes add --name universal --address ${KUMA_CONTROL_PLANE_URL}

#
# Create Dataplane for `kuma-example-app` service
#

echo "type: Dataplane
mesh: default
name: kuma-example-app
networking:
  inbound:
  - interface: {{ IP }}:{{ PUBLIC_PORT }}:{{ LOCAL_PORT }}
    tags:
      service: kuma-example-app" | kumactl apply -f - \
  --var IP=${KUMA_EXAMPLE_APP_IP_ADDRESS} \
  --var PUBLIC_PORT=${KUMA_EXAMPLE_APP_PUBLIC_PORT} \
  --var LOCAL_PORT=${KUMA_EXAMPLE_APP_LOCAL_PORT}

#
# Create Dataplane for `kuma-example-client` service
#

echo "type: Dataplane
mesh: default
name: kuma-example-client
networking:
  inbound:
  - interface: {{ IP }}:{{ PUBLIC_PORT }}:{{ LOCAL_PORT }}
    tags:
      service: kuma-example-client
  outbound:
  - interface: :4000
    service: kuma-example-app" | kumactl apply -f - \
  --var IP=${KUMA_EXAMPLE_CLIENT_IP_ADDRESS} \
  --var PUBLIC_PORT=${KUMA_EXAMPLE_CLIENT_PUBLIC_PORT} \
  --var LOCAL_PORT=${KUMA_EXAMPLE_CLIENT_LOCAL_PORT}
