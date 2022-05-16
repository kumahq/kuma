#!/bin/sh

if [ "$1" = "--version" ];
then
    printf "\nvCoreDNS-1.8.3\ndarwin/amd64, go1.17.3, 1.4.0-rc1-13-g7f3938a7f-dirty\n"
    exit 0
fi

# Send logs for Cmd#Wait to finish
while true;
do
  echo "Log"
  sleep 0.1
done

