#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "usage: $0 <N>   (positive integer for cluster suffix kuma-N)" >&2
  exit 1
fi

N="$1"
if ! [[ "$N" =~ ^[1-9][0-9]*$ ]]; then
  echo "error: N must be a positive integer, got: $N" >&2
  exit 1
fi

CLUSTER="kuma-$N"
KUBECONFIG_PATH="$HOME/.kube/k3d-${CLUSTER}.yaml"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$REPO_ROOT"

echo "==> 1. Creating k3d cluster ${CLUSTER}"
# Use k3d directly (no makefile) so we don't inherit the project's "kuma"
# docker network, which is often pre-created with a /24 subnet that MetalLB
# rejects. k3d will create its own network (k3d-${CLUSTER}) from Docker's
# default address pool.
k3d cluster create "${CLUSTER}" \
  --k3s-arg "--disable=traefik@server:0" \
  --k3s-arg "--disable=servicelb@server:0" \
  --k3s-arg "--disable=metrics-server@server:0" \
  --wait

echo "==> 1a. Writing kubeconfig to ${KUBECONFIG_PATH}"
mkdir -p "$(dirname "${KUBECONFIG_PATH}")"
k3d kubeconfig get "${CLUSTER}" > "${KUBECONFIG_PATH}"
export KUBECONFIG="$KUBECONFIG_PATH"

echo "==> 2. Loading images into ${CLUSTER}"
make k3d/cluster/load CLUSTER="${CLUSTER}"

echo "==> 3. Installing control plane via helm"
helm upgrade --install kuma deployments/charts/kuma \
  --set 'meshes[0].name=default' \
  --set 'meshes[0].ingress.enabled=true' \
  --set 'meshes[0].egress.enabled=true' \
  --set 'dataPlane.features.unifiedResourceNaming=true' \
  -n kuma-system \
  --create-namespace

echo "==> 3a. Waiting for control plane to be ready"
kubectl -n kuma-system rollout status deploy/kuma-control-plane --timeout=300s
kubectl -n kuma-system wait --for=condition=available deploy/kuma-control-plane --timeout=300s

echo "==> 4. Installing demo"
kubectl apply -f build/k8s/001-with-mtls.yaml

echo "==> 4a. Waiting for kuma-demo namespace and workloads"
for _ in $(seq 1 60); do
  if kubectl get ns kuma-demo >/dev/null 2>&1; then
    break
  fi
  sleep 2
done

if kubectl get ns kuma-demo >/dev/null 2>&1; then
  deploys=$(kubectl -n kuma-demo get deploy -o name 2>/dev/null || true)
  for d in $deploys; do
    kubectl -n kuma-demo rollout status "$d" --timeout=300s || true
  done
fi

echo "==> 5. Removing spec.mtls from default mesh and deleting allow-all MeshTrafficPermission"
kubectl patch mesh default --type=json -p='[{"op":"remove","path":"/spec/mtls"}]' || \
  echo "(spec.mtls already absent)"

# The default allow-all MTP created by the chart
kubectl -n kuma-system delete meshtrafficpermission allow-all-default --ignore-not-found

echo "==> 6. Applying MeshIdentity"
cat <<'EOF' | kubectl apply -f -
apiVersion: kuma.io/v1alpha1
kind: MeshIdentity
metadata:
  name: identity
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  selector:
    dataplane:
      matchLabels: {}
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      autogenerate:
        enabled: true
EOF

echo "==> 7. Applying MeshTrafficPermission"
cat <<'EOF' | kubectl apply -f -
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  labels:
    kuma.io/origin: zone
  name: kv
  namespace: kuma-system
spec:
  rules:
  - default:
      allow:
      - spiffeID:
          type: Prefix
          value: spiffe://default.default.mesh.local
EOF

echo "==> 8. Applying MeshExternalService"
cat <<'EOF' | kubectl apply -f -
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: httpbin
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: httpbin.org
      port: 80
EOF

echo "==> 8a. Giving the CP a moment to roll out new config"
sleep 10

echo "==> 8b. Applying MeshAccessLog"
cat <<'EOF' | kubectl apply -f -
apiVersion: kuma.io/v1alpha1
kind: MeshAccessLog
metadata:
  name: backend-access-log
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  rules:
    - matches:
        - sni:
            type: Exact
            value: sni.extsvc.default.default.kuma-system.httpbin.80
      default:
        backends:
          - type: File
            file:
              format:
                type: Plain
                plain: "sni=httpbin uri_san=[%DYNAMIC_METADATA(kuma.mesh_access_log:uri_san)%]"
              path: /dev/stdout
    - matches:
        - spiffeID:
            type: Exact
            value: "spiffe://default.default.mesh.local/ns/kuma-demo/sa/demo-app"
      default:
        backends:
          - type: File
            file:
              format:
                type: Plain
                plain: "spiffe=demo-app uri_san=[%DYNAMIC_METADATA(kuma.mesh_access_log:uri_san)%]"
              path: /dev/stdout
EOF

echo "==> 8c. Giving the CP a moment to roll out MAL config"
sleep 5

echo "==> 9. Testing curl httpbin.extsvc.mesh.local"

# Run curl from an ephemeral debug container (netshoot) sharing the target
# pod's network namespace, since the app containers don't ship curl.
curl_from_pod() {
  local ns="$1" pod="$2"
  local target
  target=$(kubectl -n "$ns" get pod "$pod" -o jsonpath='{.spec.containers[0].name}')
  echo "--> curl from $ns/$pod (debug container, target=$target)"
  kubectl -n "$ns" debug -q -it "$pod" \
    --image=nicolaka/netshoot \
    --target="$target" \
    -- curl -sS -o /dev/null -w "HTTP %{http_code}\n" httpbin.extsvc.mesh.local
}

# Both demo-app and kv pods live in the kuma-demo namespace, labelled
# app=demo-app and app=kv respectively.
find_pod() {
  local selector="$1"
  kubectl -n kuma-demo get pod -l "$selector" \
    -o jsonpath='{.items[?(@.status.phase=="Running")].metadata.name}' 2>/dev/null | awk '{print $1}'
}

for app in demo-app kv; do
  POD=$(find_pod "app=$app" || true)
  if [[ -n "$POD" ]]; then
    curl_from_pod kuma-demo "$POD"
  else
    echo "WARN: no running pod with app=$app in kuma-demo namespace" >&2
  fi
done

echo "==> Done."
