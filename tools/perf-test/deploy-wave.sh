#!/bin/bash
# deploy-wave.sh — deploy a wave of fake-service workloads into a Kuma-injected cluster
#
# Usage: deploy-wave.sh <wave_num> <num_namespaces> <services_per_ns> <replicas>
#
# Env:
#   KUBE_CONTEXT  — kubectl context (default: current context)
#   NS_PREFIX     — namespace prefix (default: perf-wave)
#   IMAGE         — workload image (default: nicholasjackson/fake-service:v0.26.2)
#
# Example:
#   KUBE_CONTEXT=gke_my-project_us-central1_perf deploy-wave.sh 1 5 5 2   # 50 DPs
set -e

WAVE=$1
NS_COUNT=$2
SVC_COUNT=$3
REPLICAS=$4

CTX_ARG=()
[[ -n "$KUBE_CONTEXT" ]] && CTX_ARG=(--context="$KUBE_CONTEXT")
NS_PREFIX=${NS_PREFIX:-perf-wave}
IMAGE=${IMAGE:-nicholasjackson/fake-service:v0.26.2}

if [[ -z "$WAVE" || -z "$NS_COUNT" || -z "$SVC_COUNT" || -z "$REPLICAS" ]]; then
  cat <<USAGE
Usage: $0 <wave_num> <num_namespaces> <services_per_ns> <replicas>

Reference wave sizing:
  $0 1 5  5 2   # 50 DPs
  $0 2 10 5 2   # 100 DPs (150 cumulative)
  $0 3 20 5 3   # 300 DPs (450 cumulative)
  $0 4 30 10 3  # 900 DPs (1350 cumulative)
  $0 5 20 10 5  # 1000 DPs (2350 cumulative)
USAGE
  exit 1
fi

render_workload() {
  local j=$1
  cat <<YAML
apiVersion: apps/v1
kind: Deployment
metadata:
  name: svc-${j}
spec:
  replicas: $REPLICAS
  selector:
    matchLabels:
      app: svc-${j}
  template:
    metadata:
      labels:
        app: svc-${j}
    spec:
      containers:
      - name: fake-service
        image: ${IMAGE}
        ports:
        - containerPort: 9090
        env:
        - name: LISTEN_ADDR
          value: "0.0.0.0:9090"
        - name: NAME
          value: "svc-${j}"
        resources:
          requests:
            cpu: 10m
            memory: 32Mi
          limits:
            memory: 64Mi
---
apiVersion: v1
kind: Service
metadata:
  name: svc-${j}
spec:
  selector:
    app: svc-${j}
  ports:
  - port: 9090
YAML
}

echo "Deploying wave $WAVE: ${NS_COUNT} namespaces × ${SVC_COUNT} services × ${REPLICAS} replicas"

for i in $(seq 1 "$NS_COUNT"); do
  NS="${NS_PREFIX}${WAVE}-${i}"
  kubectl "${CTX_ARG[@]}" create namespace "$NS" --dry-run=client -o yaml | kubectl "${CTX_ARG[@]}" apply -f -
  kubectl "${CTX_ARG[@]}" label namespace "$NS" kuma.io/sidecar-injection=enabled --overwrite

  for j in $(seq 1 "$SVC_COUNT"); do
    render_workload "$j" | kubectl "${CTX_ARG[@]}" -n "$NS" apply -f -
  done
done

echo "Wave $WAVE deployed. Waiting for pods to be ready..."
for i in $(seq 1 "$NS_COUNT"); do
  NS="${NS_PREFIX}${WAVE}-${i}"
  kubectl "${CTX_ARG[@]}" -n "$NS" wait --for=condition=available deployment --all --timeout=300s 2>/dev/null || true
done
echo "Wave $WAVE ready."
