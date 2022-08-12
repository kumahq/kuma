#!/bin/bash

LABEL_TO_CHECK="$1"
CURL_OUTPUT=$(curl -s --fail -H "Accept: application/vnd.github+json" https://api.github.com/repos/kumahq/kuma/pulls/"$CIRCLE_PR_NUMBER")

if [[ $CURL_OUTPUT != "" ]]; then
    LABELS=$(jq '.labels[].name' <<< "$CURL_OUTPUT")
    if echo "$LABELS" | grep -q "$LABEL_TO_CHECK"; then
        echo "true"
    else
        echo "false"
    fi
else
    echo "curl call failed"
fi
