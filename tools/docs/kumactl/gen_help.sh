#!/bin/bash

if [ $# -ne 1 ]; then
    echo 'This script must be called with exactly one parameter - path to `kumactl` binary' >&2
    exit 1
fi

# associate a full path with the tool name
declare -a tools
tools["kumactl"]=$1

function gen_help() {
    command=$@
    level=$#
    # map tool name onto a full path to it
    binary=${tools[$1]}
    shift # drop the first argument (tool name)
    subcommand=$@
    for i in `seq 1 ${level}`; do printf "#"; done; echo " ${command}"
    echo
    echo '```'
    ${binary} ${subcommand} --help
    echo '```'
    echo
}

gen_help kumactl
gen_help kumactl apply
gen_help kumactl config
gen_help kumactl config view
gen_help kumactl config control-planes
gen_help kumactl config control-planes list
gen_help kumactl config control-planes add
gen_help kumactl config control-planes remove
gen_help kumactl config control-planes switch
gen_help kumactl install
gen_help kumactl install control-plane
gen_help kumactl install postgres-schema
gen_help kumactl get
gen_help kumactl get meshes
gen_help kumactl get dataplanes
gen_help kumactl get proxytemplates
gen_help kumactl get traffic-permissions
gen_help kumactl inspect
gen_help kumactl inspect dataplanes
gen_help kumactl version