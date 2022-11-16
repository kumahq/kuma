#!/usr/bin/env bash

versionInfo=$("${BASH_SOURCE%/*}"/version.sh)

IFS=' ' read -r -a values <<< "$versionInfo"

fields=( "version" "gitTag" "gitCommit" "buildDate" "Envoy" )

for index in "${!fields[@]}"
do
   echo -n "-X github.com/kumahq/kuma/pkg/version.${fields[index]}=${values[index]} "
done
