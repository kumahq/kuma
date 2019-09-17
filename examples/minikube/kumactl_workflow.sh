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

run kumactl config view

run kumactl config control-planes list

# Install kubectl
run curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/${GOOS:-linux}/${GOARCH:-amd64}/kubectl
run chmod +x kubectl

run export PATH=.:$PATH

# Forward CP API server from k8s onto localhost
run kubectl port-forward -n kuma-system $(kubectl get pods -n kuma-system -l app=kuma-control-plane -o=jsonpath='{.items[0].metadata.name}') 15681:5681 &

# Give port-forward 10 seconds to come alive -- else you won't be able to connect to the control plane
run curl --retry 10 --retry-delay 1 --retry-connrefused http://localhost:15681

# Add the CP to the config
run kumactl config control-planes add --name demo-kubectl-port-forward --address http://localhost:15681

run kumactl config view

run kumactl config control-planes list

run kumactl get meshes

run kumactl get dataplanes

run kumactl get proxytemplates

run kumactl get traffic-permissions

run kumactl get traffic-loggings

run kumactl inspect dataplanes

run kumactl inspect dataplanes -otable

run kumactl inspect dataplanes -oyaml

run kumactl inspect dataplanes -ojson

run test $(run kumactl inspect dataplanes | tail +4 | grep -v '<<<<<' | grep -v -e '^$' | wc -l) -eq 2

# Kill the port-forward
run killall kubectl

# Forward CP API server from k8s onto localhost
run kubectl proxy &

# Give the proxy 10 seconds to come alive -- else you won't be able to connect to the control plane
run curl --retry 10 --retry-delay 1 --retry-connrefused http://localhost:8001

# Add the CP to the config
run kumactl config control-planes add --name demo-kubectl-proxy --address http://localhost:8001/api/v1/namespaces/kuma-system/services/kuma-control-plane:5681/proxy

run kumactl config view

run kumactl config control-planes list

run kumactl get meshes

run kumactl get dataplanes

run kumactl get proxytemplates

run kumactl get traffic-permissions

run kumactl inspect dataplanes

run kumactl inspect dataplanes -otable

run kumactl inspect dataplanes -oyaml

run kumactl inspect dataplanes -ojson

run kumactl get traffic-permissions

run test $(run kumactl inspect dataplanes | tail +4 | grep -v '<<<<<' | grep -v -e '^$' | wc -l) -eq 2

# Kill the proxy
run killall kubectl

