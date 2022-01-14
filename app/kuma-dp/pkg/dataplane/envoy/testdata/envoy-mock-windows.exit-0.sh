#!/bin/sh

if [ "$1" = "--version" ];
then
  printf "\r\nC:\\\Users\\\kuma\\\envoy\\\versions\\\1.19.0\\\bin\\\envoy.exe  version: 68fe53a889416fd8570506232052b06f5a531541/1.19.0/Modified/RELEASE/BoringSSL\r\n\r\n"
  exit 0
fi
# print arguments to verify in the test
echo "$@"
