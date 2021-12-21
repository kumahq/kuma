#!/usr/bin/env sh

set -e

#
# Utility functions
#

resolve_ip() {
  getent hosts "${DATAPLANE_HOSTNAME}" 2>/dev/null | awk -e '{ print $1 }'
}

fail() {
  printf 'Error: %s\n' "${1}" >&2  ## Send message to stderr. Exclude >&2 if you don't want it that way.
  exit "${2-1}"                    ## Return a code specified by $2 or 1 by default.
}

create_dataplane() {
  DATAPLANE_HOSTNAME="$1"
  DATAPLANE_PUBLIC_PORT=$2
  DATAPLANE_LOCAL_PORT=$3
  DATAPLANE_RESOURCE="$4"
  DATAPLANE_NAME="${DATAPLANE_HOSTNAME}"

  #
  # Resolve IP address allocated to the "${DATAPLANE_NAME}" container
  #

  DATAPLANE_IP_ADDRESS=$( resolve_ip "${DATAPLANE_HOSTNAME}" )
  if [ -z "${DATAPLANE_IP_ADDRESS}" ]; then
    fail "failed to resolve IP address allocated to the '${DATAPLANE_HOSTNAME}' container"
  fi
  echo "'${DATAPLANE_HOSTNAME}' has the following IP address: ${DATAPLANE_IP_ADDRESS}"

  #
  # Create Dataplane for "${DATAPLANE_NAME}"
  #

  echo "${DATAPLANE_RESOURCE}" | kumactl apply -f - \
    --var IP="${DATAPLANE_IP_ADDRESS}" \
    --var PUBLIC_PORT="${DATAPLANE_PUBLIC_PORT}" \
    --var LOCAL_PORT="${DATAPLANE_LOCAL_PORT}"

  #
  # Create token for "${DATAPLANE_NAME}"
  #

  kumactl generate dataplane-token --name="${DATAPLANE_NAME}" > /"${DATAPLANE_NAME}"/token
}

#
# Arguments
#

KUMA_CONTROL_PLANE_URL=https://kuma-control-plane:5682

KUMA_EXAMPLE_APP_HOSTNAME=kuma-example-app
KUMA_EXAMPLE_APP_PUBLIC_PORT=8000
KUMA_EXAMPLE_APP_LOCAL_PORT=80

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

kumactl config control-planes add --name universal --address ${KUMA_CONTROL_PLANE_URL} --ca-cert-file /certs/server/cert.pem --client-cert-file /certs/client/cert.pem --client-key-file /certs/client/cert.key --overwrite

#
# Create Dataplane for `kuma-example-app` service
#

create_dataplane "${KUMA_EXAMPLE_APP_HOSTNAME}" ${KUMA_EXAMPLE_APP_PUBLIC_PORT} ${KUMA_EXAMPLE_APP_LOCAL_PORT} "
type: Dataplane
mesh: default
name: kuma-example-app
networking:
  address: {{ IP }}
  inbound:
  - port: {{ PUBLIC_PORT }}
    servicePort: {{ LOCAL_PORT }}
    tags:
      kuma.io/service: kuma-example-app
      kuma.io/protocol: http
"

#
# Create Dataplane for `kuma-example-client` service
#

create_dataplane "${KUMA_EXAMPLE_CLIENT_HOSTNAME}" ${KUMA_EXAMPLE_CLIENT_PUBLIC_PORT} ${KUMA_EXAMPLE_CLIENT_LOCAL_PORT} "
type: Dataplane
mesh: default
name: kuma-example-client
networking:
  address: {{ IP }}
  inbound:
  - port: {{ PUBLIC_PORT }}
    servicePort: {{ LOCAL_PORT }}
    tags:
      kuma.io/service: kuma-example-client
  outbound:
  - port: 4000
    service: kuma-example-app"

#
# Create Dataplane for `kuma-example-web` service
#

create_dataplane "${KUMA_EXAMPLE_WEB_HOSTNAME}" ${KUMA_EXAMPLE_WEB_PUBLIC_PORT} ${KUMA_EXAMPLE_WEB_LOCAL_PORT} "
type: Dataplane
mesh: default
name: kuma-example-web
networking:
  address: {{ IP }}
  inbound:
  - port: {{ PUBLIC_PORT }}
    servicePort: {{ LOCAL_PORT }}
    tags:
      kuma.io/service: kuma-example-web
      version: v8
      env: prod
  outbound:
  - port: 5000
    service: kuma-example-backend"

#
# Create Dataplane v1 for `kuma-example-backend` service
#

create_dataplane "${KUMA_EXAMPLE_BACKEND_V1_HOSTNAME}" ${KUMA_EXAMPLE_BACKEND_V1_PUBLIC_PORT} ${KUMA_EXAMPLE_BACKEND_V1_LOCAL_PORT} "
type: Dataplane
mesh: default
name: kuma-example-backend-v1
networking:
  address: {{ IP }}
  inbound:
  - port: {{ PUBLIC_PORT }}
    servicePort: {{ LOCAL_PORT }}
    tags:
      kuma.io/service: kuma-example-backend
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
  address: {{ IP }}
  inbound:
  - port: {{ PUBLIC_PORT }}
    servicePort: {{ LOCAL_PORT }}
    tags:
      kuma.io/service: kuma-example-backend
      version: v2
      env: intg"
