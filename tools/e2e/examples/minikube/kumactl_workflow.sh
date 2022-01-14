#!/usr/bin/env bash

set -e

run() {
    command=$*
    echo ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>"
    echo '$' "$command"
    echo
    $command
    echo "<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<"
    echo
}

workflow() {

    run kumactl config view

    run kumactl config control-planes list

    run kumactl get meshes
    run kumactl get meshes -otable
    run kumactl get meshes -oyaml
    run kumactl get meshes -ojson

    run kumactl get dataplanes
    run kumactl get dataplanes -otable
    run kumactl get dataplanes -oyaml
    run kumactl get dataplanes -ojson

    run kumactl get healthchecks
    run kumactl get healthchecks -otable
    run kumactl get healthchecks -oyaml
    run kumactl get healthchecks -ojson

    run kumactl get retries
    run kumactl get retries -otable
    run kumactl get retries -oyaml
    run kumactl get retries -ojson

    run kumactl get proxytemplates
    run kumactl get proxytemplates -otable
    run kumactl get proxytemplates -oyaml
    run kumactl get proxytemplates -ojson

    run kumactl get traffic-logs
    run kumactl get traffic-logs -otable
    run kumactl get traffic-logs -oyaml
    run kumactl get traffic-logs -ojson

    run kumactl get traffic-permissions
    run kumactl get traffic-permissions -otable
    run kumactl get traffic-permissions -oyaml
    run kumactl get traffic-permissions -ojson

    run kumactl get traffic-routes
    run kumactl get traffic-routes -otable
    run kumactl get traffic-routes -oyaml
    run kumactl get traffic-routes -ojson

    run kumactl get traffic-traces
    run kumactl get traffic-traces -otable
    run kumactl get traffic-traces -oyaml
    run kumactl get traffic-traces -ojson

    run kumactl inspect dataplanes
    run kumactl inspect dataplanes -otable
    run kumactl inspect dataplanes -oyaml
    run kumactl inspect dataplanes -ojson

    run test "$(run kumactl inspect dataplanes | tail +4 | grep -v '<<<<<' | grep -v -e '^$' -c) -eq 3"
}

# Killing the kubectl port-forward at the end of the script -- regardless of exit status
trap 'killall kubectl && rm $HOME/kubectl' EXIT

# Print version
run kumactl version

# Dump config
run kumactl config view

# List control planes
run kumactl config control-planes list

# Install kubectl
run curl -LO https://storage.googleapis.com/kubernetes-release/release/"$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)"/bin/"${GOOS:-linux}"/"${GOARCH:-amd64}"/kubectl
run chmod +x kubectl

run export PATH=.:"$PATH"

# Forward CP API server from k8s onto localhost
run kubectl port-forward -n kuma-system "$(kubectl get pods -n kuma-system -l app=kuma-control-plane -o=jsonpath='{.items[0].metadata.name}')" 15681:5681 >/dev/null 2>&1 &

# Give port-forward 10 seconds to come alive -- else you won't be able to connect to the control plane
run curl --retry 10 --retry-delay 1 --retry-connrefused http://localhost:15681

# Add the CP to the config
run kumactl config control-planes add --name demo-kubectl-port-forward --address http://localhost:15681

# Run a sequence of commands
workflow

# Kill the port-forward
run killall kubectl

# Forward CP API server from k8s onto localhost
run kubectl proxy >/dev/null 2>&1 &

# Give the proxy 10 seconds to come alive -- else you won't be able to connect to the control plane
run curl --retry 10 --retry-delay 1 --retry-connrefused http://localhost:8001

# Add the CP to the config
run kumactl config control-planes add --name demo-kubectl-proxy --address http://localhost:8001/api/v1/namespaces/kuma-system/services/kuma-control-plane:5681/proxy

# Run a sequence of commands
workflow

# Kill the proxy
run killall kubectl
