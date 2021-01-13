#!/usr/bin/env sh

missing_crds="$(kumactl install crds --not-installed)"

if [ -n "${missing_crds}" ]; then
 echo "${missing_crds}" | kubectl create -f -
fi
