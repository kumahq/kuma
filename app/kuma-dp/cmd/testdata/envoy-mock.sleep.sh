#!/bin/sh

echo $$ >${ENVOY_MOCK_PID_FILE}

touch ${ENVOY_MOCK_CMDLINE_FILE}
for arg in "$@"; do
    echo $arg >> ${ENVOY_MOCK_CMDLINE_FILE}
done

sleep 86400
