#!/usr/bin/env bash

LABEL_TO_CHECK="$1"
# CIRCLE_PULL_REQUEST looks like: https://github.com/kumahq/kuma/pull/2949 the api is at: https://api.github.com/kumahq/kuma/pulls/2949
URL=$(echo "${CIRCLE_PULL_REQUEST}" | sed 's:github.com:api.github.com/repos:' | sed 's:pull:pulls:')
AUTH=""
if [[ "$GITHUB_TOKEN" != "" ]]; then
  AUTH="Authorization: bearer $GITHUB_TOKEN"
fi

curl -s --fail -H "$AUTH" -H "Accept: application/vnd.github+json" "$URL" | jq --arg l "$LABEL_TO_CHECK" '.labels | any(.name == $l)'
