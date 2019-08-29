#!/usr/bin/env bash

set -e

# Redirect all traffic into Envoy
/usr/local/bin/istio-iptables.sh -p 15001 -u 1337 -g 1337 -m REDIRECT -b '*' -i '*'

IP=$(hostname --ip-address)
PORT=8080

# Simulate application that accepts HTTP requests and makes outgoing calls
while true ; do nc -l -p ${PORT} -c "curl http://mockbin.org/request" >/dev/null 2>/dev/null ; done &

# Run Envoy
su -c "/usr/local/bin/envoy -c /etc/envoy/envoy.yaml --config-yaml \"node: {metadata: {'IPS': '${IP}', 'PORTS': '${PORT}'}}\"" envoy
