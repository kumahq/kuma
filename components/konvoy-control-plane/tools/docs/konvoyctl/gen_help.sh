#!/bin/bash

if [ $# -ne 1 ]; then
    echo 'This script must be called with exactly one parameter - path to `konvoyctl` binary' >&2
    exit 1
fi

# associate a full path with the tool name
declare -a tools
tools["konvoyctl"]=$1

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

gen_help konvoyctl
gen_help konvoyctl config
gen_help konvoyctl config view
gen_help konvoyctl config control-planes
gen_help konvoyctl config control-planes list
gen_help konvoyctl config control-planes add
gen_help konvoyctl install
gen_help konvoyctl get
gen_help konvoyctl get meshes
gen_help konvoyctl get proxytemplates
gen_help konvoyctl inspect
gen_help konvoyctl inspect dataplanes