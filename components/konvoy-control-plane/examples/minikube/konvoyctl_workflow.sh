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

run konvoyctl config view

run konvoyctl config control-planes list

run konvoyctl config control-planes add k8s --name demo

run konvoyctl config view

run konvoyctl config control-planes list

run konvoyctl get dataplanes

run konvoyctl get dataplanes -otable

run konvoyctl get dataplanes -oyaml

run konvoyctl get dataplanes -ojson

run test $(run konvoyctl get dataplanes | tail +4 | grep -v '<<<<<' | grep -v -e '^$' | wc -l) -eq 2
