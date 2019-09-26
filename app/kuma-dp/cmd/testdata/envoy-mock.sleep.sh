#!/bin/sh

touch ${ENVOY_MOCK_CMDLINE_FILE}
for arg in "$@"; do
    echo $arg >> ${ENVOY_MOCK_CMDLINE_FILE}
done

echo $$ >${ENVOY_MOCK_PID_FILE}

sleep 86400
