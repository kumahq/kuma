#!/bin/sh

if [ ! -e "${ENVOY_MOCK_CMDLINE_FILE}" ]; then
    touch "${ENVOY_MOCK_CMDLINE_FILE}"
fi

if [ "$1" = "--version" ];
then
    echo "$1" >> "${ENVOY_MOCK_CMDLINE_FILE}"
    printf "\nversion: 50ef0945fa2c5da4bff7627c3abf41fdd3b7cffd/1.15.0/clean-getenvoy-2aa564b-envoy/RELEASE/BoringSSL\n"
    exit 0
else
    for arg in "$@"; do
        echo "$arg" >> "${ENVOY_MOCK_CMDLINE_FILE}"
    done
fi

echo $$ >"${ENVOY_MOCK_PID_FILE}"

sleep 86400
