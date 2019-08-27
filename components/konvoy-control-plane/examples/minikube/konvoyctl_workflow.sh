#!/bin/sh

set -e

function run() {
    command=$@
    echo ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
    echo '$' $command
    echo
    $command
    echo "<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
    echo
}

# Killing the kubectl port-forward at the end of the script -- regardless of exit status
trap "killall kubectl && rm $HOME/kubectl" EXIT

run konvoyctl config view

run konvoyctl config control-planes list

# Install kubectl
run curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/${GOOS:-linux}/${GOARCH:-amd64}/kubectl
run chmod +x kubectl

run export PATH=.:$PATH

# Forward CP API server from k8s onto localhost
run kubectl port-forward -n konvoy-system $(kubectl get pods -n konvoy-system -l app=konvoy-control-plane -o=jsonpath='{.items[0].metadata.name}') 15681:5681 &

# Give port-forward 10 seconds to come alive -- else you won't be able to connect to the control plane
run curl --retry 10 --retry-delay 1 --retry-connrefused http://localhost:15681

# Add the CP to the config
run konvoyctl config control-planes add --name demo-kubectl-port-forward --api-server-url http://localhost:15681

run konvoyctl config view

run konvoyctl config control-planes list

run konvoyctl inspect dataplanes

run konvoyctl inspect dataplanes -otable

run konvoyctl inspect dataplanes -oyaml

run konvoyctl inspect dataplanes -ojson

run test $(run konvoyctl inspect dataplanes | tail +4 | grep -v '<<<<<' | grep -v -e '^$' | wc -l) -eq 2

# Kill the port-forward
run killall kubectl

# Forward CP API server from k8s onto localhost
run kubectl proxy &

# Give the proxy 10 seconds to come alive -- else you won't be able to connect to the control plane
run curl --retry 10 --retry-delay 1 --retry-connrefused http://localhost:8001

# Add the CP to the config
run konvoyctl config control-planes add --name demo-kubectl-proxy --api-server-url http://localhost:8001/api/v1/namespaces/konvoy-system/services/konvoy-control-plane:5681/proxy

run konvoyctl config view

run konvoyctl config control-planes list

run konvoyctl inspect dataplanes

run konvoyctl inspect dataplanes -otable

run konvoyctl inspect dataplanes -oyaml

run konvoyctl inspect dataplanes -ojson

run konvoyctl get traffic-permissions

run test $(run konvoyctl inspect dataplanes | tail +4 | grep -v '<<<<<' | grep -v -e '^$' | wc -l) -eq 2

# Kill the proxy
run killall kubectl

