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

function create_dataplane {
  DATAPLANE_HOSTNAME="$1"
  DATAPLANE_PUBLIC_PORT=$2
  DATAPLANE_LOCAL_PORT=$3
  DATAPLANE_RESOURCE="$4"
  DATAPLANE_NAME="${DATAPLANE_HOSTNAME}"

  #
  # Resolve IP address allocated to the "${DATAPLANE_NAME}" container
  #

  DATAPLANE_IP_ADDRESS=$( resolve_ip ${DATAPLANE_HOSTNAME} )
  if [ -z "${DATAPLANE_IP_ADDRESS}" ]; then
    fail "failed to resolve IP address allocated to the '${DATAPLANE_HOSTNAME}' container"
  fi
  echo "'${DATAPLANE_HOSTNAME}' has the following IP address: ${DATAPLANE_IP_ADDRESS}"

  #
  # Create Dataplane for "${DATAPLANE_NAME}"
  #

  echo "${DATAPLANE_RESOURCE}" | kumactl apply -f - \
    --var IP=${DATAPLANE_IP_ADDRESS} \
    --var PUBLIC_PORT=${DATAPLANE_PUBLIC_PORT} \
    --var LOCAL_PORT=${DATAPLANE_LOCAL_PORT}

  #
  # Create token for "${DATAPLANE_NAME}"
  #

  kumactl generate dataplane-token --dataplane=${DATAPLANE_NAME} > /${DATAPLANE_NAME}/token
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

KUMA_EXAMPLE_WEB_HOSTNAME=kuma-example-web
KUMA_EXAMPLE_WEB_PUBLIC_PORT=6060
KUMA_EXAMPLE_WEB_LOCAL_PORT=6060

KUMA_EXAMPLE_BACKEND_V1_HOSTNAME=kuma-example-backend-v1
KUMA_EXAMPLE_BACKEND_V1_PUBLIC_PORT=7070
KUMA_EXAMPLE_BACKEND_V1_LOCAL_PORT=7070

KUMA_EXAMPLE_BACKEND_V2_HOSTNAME=kuma-example-backend-v2
KUMA_EXAMPLE_BACKEND_V2_PUBLIC_PORT=7070
KUMA_EXAMPLE_BACKEND_V2_LOCAL_PORT=7070

#
# Configure `kumactl`
#

kumactl config control-planes add --name universal --address ${KUMA_CONTROL_PLANE_URL} --admin-client-cert /certs/client/cert.pem --admin-client-key /certs/client/cert.key --overwrite

#
# Create Dataplane for `kuma-example-app` service
#

create_dataplane "${KUMA_EXAMPLE_APP_HOSTNAME}" ${KUMA_EXAMPLE_APP_PUBLIC_PORT} ${KUMA_EXAMPLE_APP_LOCAL_PORT} "
type: Dataplane
mesh: default
name: kuma-example-app
networking:
  inbound:
  - interface: {{ IP }}:{{ PUBLIC_PORT }}:{{ LOCAL_PORT }}
    tags:
      service: kuma-example-app
      protocol: http
"

#
# Create Dataplane for `kuma-example-client` service
#

create_dataplane "${KUMA_EXAMPLE_CLIENT_HOSTNAME}" ${KUMA_EXAMPLE_CLIENT_PUBLIC_PORT} ${KUMA_EXAMPLE_CLIENT_LOCAL_PORT} "
type: Dataplane
mesh: default
name: kuma-example-client
networking:
  inbound:
  - interface: {{ IP }}:{{ PUBLIC_PORT }}:{{ LOCAL_PORT }}
    tags:
      service: kuma-example-client
  outbound:
  - interface: :4000
    service: kuma-example-app"

#
# Create Dataplane for `kuma-example-web` service
#

create_dataplane "${KUMA_EXAMPLE_WEB_HOSTNAME}" ${KUMA_EXAMPLE_WEB_PUBLIC_PORT} ${KUMA_EXAMPLE_WEB_LOCAL_PORT} "
type: Dataplane
mesh: default
name: kuma-example-web
networking:
  inbound:
  - interface: {{ IP }}:{{ PUBLIC_PORT }}:{{ LOCAL_PORT }}
    tags:
      service: kuma-example-web
      version: v8
      env: prod
  outbound:
  - interface: :5000
    service: kuma-example-backend"

#
# Create Dataplane v1 for `kuma-example-backend` service
#

create_dataplane "${KUMA_EXAMPLE_BACKEND_V1_HOSTNAME}" ${KUMA_EXAMPLE_BACKEND_V1_PUBLIC_PORT} ${KUMA_EXAMPLE_BACKEND_V1_LOCAL_PORT} "
type: Dataplane
mesh: default
name: kuma-example-backend-v1
networking:
  inbound:
  - interface: {{ IP }}:{{ PUBLIC_PORT }}:{{ LOCAL_PORT }}
    tags:
      service: kuma-example-backend
      protocol: http
      version: v1
      env: prod"

#
# Create Dataplane v2 for `kuma-example-backend` service
#

create_dataplane "${KUMA_EXAMPLE_BACKEND_V2_HOSTNAME}" ${KUMA_EXAMPLE_BACKEND_V2_PUBLIC_PORT} ${KUMA_EXAMPLE_BACKEND_V2_LOCAL_PORT} "
type: Dataplane
mesh: default
name: kuma-example-backend-v2
networking:
  inbound:
  - interface: {{ IP }}:{{ PUBLIC_PORT }}:{{ LOCAL_PORT }}
    tags:
      service: kuma-example-backend
      protocol: http
      version: v2
      env: intg"
