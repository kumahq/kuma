#!/bin/sh

if [ "$1" = "--version" ];
then
  printf "\nenvoy  version: 50ef0945fa2c5da4bff7627c3abf41fdd3b7cffd/1.15.0/clean-getenvoy-2aa564b-envoy/RELEASE/BoringSSL\n\n"
  exit 0
fi

# print arguments to verify in the test
echo "$@"
