#!/usr/bin/env bash

NAMESPACE=${1:-kuma-system}

kubectl create namespace "$NAMESPACE" || true

kubectl patch namespace/"$NAMESPACE" \
  --type merge \
  --patch '{ "metadata": { "labels": { "kuma.io/system-namespace": "true" } } }'
