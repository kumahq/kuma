#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  cluster-lifecycle.sh --repo-root <path> single-up [cluster-name]
  cluster-lifecycle.sh --repo-root <path> single-down [cluster-name]
  cluster-lifecycle.sh --repo-root <path> global-zone-up [global-cluster] [zone-cluster] [zone-name]
  cluster-lifecycle.sh --repo-root <path> global-zone-down [global-cluster] [zone-cluster]
  cluster-lifecycle.sh --repo-root <path> global-two-zones-up [global] [zone1] [zone2] [zone1-name] [zone2-name]
  cluster-lifecycle.sh --repo-root <path> global-two-zones-down [global] [zone1] [zone2]

Options:
  --repo-root <path>  Path to Kuma repository root (required)

Defaults:
  single-up/down cluster-name: kuma-1
  global cluster: kuma-1
  zone-1 cluster: kuma-2
  zone-2 cluster: kuma-3
  zone names: zone-1, zone-2

Performance toggles (optional env vars):
  HARNESS_BUILD_IMAGES=1      Build/tag local images once per invocation
  HARNESS_LOAD_IMAGES=1       Load images once per cluster per invocation
  HARNESS_HELM_CLEAN=0        Fast default: keep release/namespace between deploys
                              (set 1 for strict clean-state redeploy)
EOF
}

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root=""
generated_metallb_manifests=()
loaded_image_clusters=()
images_prepared=0

HARNESS_BUILD_IMAGES="${HARNESS_BUILD_IMAGES:-1}"
HARNESS_LOAD_IMAGES="${HARNESS_LOAD_IMAGES:-1}"
HARNESS_HELM_CLEAN="${HARNESS_HELM_CLEAN:-0}"

validate_toggle() {
  local name="$1"
  local value="$2"

  if [[ "${value}" != "0" && "${value}" != "1" ]]; then
    printf 'Error: %s must be 0 or 1, got: %s\n' "${name}" "${value}" >&2
    exit 1
  fi
}

validate_toggle "HARNESS_BUILD_IMAGES" "${HARNESS_BUILD_IMAGES}"
validate_toggle "HARNESS_LOAD_IMAGES" "${HARNESS_LOAD_IMAGES}"
validate_toggle "HARNESS_HELM_CLEAN" "${HARNESS_HELM_CLEAN}"

if [[ "${HARNESS_LOAD_IMAGES}" == "0" && "${HARNESS_BUILD_IMAGES}" == "1" ]]; then
  printf 'Info: HARNESS_LOAD_IMAGES=0, disabling HARNESS_BUILD_IMAGES for this run.\n'
  HARNESS_BUILD_IMAGES=0
fi

cleanup_generated_manifests() {
  local manifest

  for manifest in "${generated_metallb_manifests[@]}"; do
    rm -f "${manifest}"
  done
}

trap cleanup_generated_manifests EXIT

kubeconfig_for_cluster() {
  local cluster_name="$1"
  printf "%s/.kube/kind-%s-config" "${HOME}" "${cluster_name}"
}

cluster_exists() {
  local cluster_name="$1"
  k3d cluster list "${cluster_name}" >/dev/null 2>&1
}

is_cluster_image_loaded() {
  local cluster_name="$1"
  local loaded

  for loaded in "${loaded_image_clusters[@]}"; do
    if [[ "${loaded}" == "${cluster_name}" ]]; then
      return 0
    fi
  done

  return 1
}

prepare_local_images_once() {
  if [[ "${HARNESS_BUILD_IMAGES}" != "1" ]]; then
    return
  fi

  if [[ "${images_prepared}" == "1" ]]; then
    return
  fi

  make --directory "${repo_root}" images
  make --directory "${repo_root}" docker/tag
  images_prepared=1
}

load_images_into_cluster_once() {
  local cluster_name="$1"

  if [[ "${HARNESS_LOAD_IMAGES}" != "1" ]]; then
    return
  fi

  if is_cluster_image_loaded "${cluster_name}"; then
    return
  fi

  prepare_local_images_once
  make --directory "${repo_root}" k3d/load/images KIND_CLUSTER_NAME="${cluster_name}"
  loaded_image_clusters+=("${cluster_name}")
}

ensure_metallb_manifest() {
  local cluster_name="$1"
  local manifest_path="${repo_root}/mk/metallb-k3d-${cluster_name}.yaml"
  local octet

  if [[ -f "${manifest_path}" ]]; then
    return
  fi

  if [[ "${cluster_name}" =~ ^kuma-([0-9]+)$ ]]; then
    octet="${BASH_REMATCH[1]}"
  else
    printf 'Error: missing MetalLB manifest for cluster %s\n' "${cluster_name}" >&2
    printf 'Expected file: %s\n' "${manifest_path}" >&2
    printf 'Supported auto-generation pattern: kuma-<number>\n' >&2
    exit 1
  fi

  if (( octet < 1 || octet > 254 )); then
    printf 'Error: cluster suffix out of supported range for %s\n' "${cluster_name}" >&2
    exit 1
  fi

  cat >"${manifest_path}" <<EOF
---
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: main
  namespace: metallb-system
spec:
  addresses:
    - 172.18.${octet}.0/24
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: empty
  namespace: metallb-system
EOF

  generated_metallb_manifests+=("${manifest_path}")
}

ensure_cluster_started() {
  local cluster_name="$1"

  ensure_metallb_manifest "${cluster_name}"

  if cluster_exists "${cluster_name}"; then
    k3d cluster start "${cluster_name}" >/dev/null 2>&1 || true
    return
  fi

  make --directory "${repo_root}" k3d/start KIND_CLUSTER_NAME="${cluster_name}"
}

stop_cluster_if_exists() {
  local cluster_name="$1"

  if cluster_exists "${cluster_name}"; then
    make --directory "${repo_root}" k3d/stop KIND_CLUSTER_NAME="${cluster_name}"
    return
  fi

  printf 'Cluster %s does not exist, skip stop.\n' "${cluster_name}"
}

deploy_helm() {
  local cluster_name="$1"
  local kubeconfig_path="$2"
  local kuma_mode="$3"
  local additional_settings="${4:-}"
  local helm_clean_toggle="${HARNESS_HELM_CLEAN}"
  local -a env_cmd

  env_cmd=(
    env
    "KUBECONFIG=${kubeconfig_path}"
    "K3D_HELM_DEPLOY_NO_CNI=true"
    "KIND_CLUSTER_NAME=${cluster_name}"
    "KUMA_MODE=${kuma_mode}"
    "K3D_DONT_LOAD=1"
  )

  if [[ -n "${additional_settings}" ]]; then
    env_cmd+=("K3D_HELM_DEPLOY_ADDITIONAL_SETTINGS=${additional_settings}")
  fi

  if [[ "${helm_clean_toggle}" == "0" ]]; then
    env_cmd+=("K3D_DEPLOY_HELM_DONT_CLEAN=1")
  fi

  "${env_cmd[@]}" make --directory "${repo_root}" k3d/deploy/helm
}

resolve_global_kds_address() {
  local global_cluster="$1"
  local global_kubeconfig="$2"
  local global_node_ip
  local global_kds_port

  global_node_ip="$(docker inspect "k3d-${global_cluster}-server-0" \
    -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}')"

  global_kds_port="$(KUBECONFIG="${global_kubeconfig}" kubectl get svc \
    -n kuma-system kuma-global-zone-sync \
    -o jsonpath='{.spec.ports[?(@.name=="global-zone-sync")].nodePort}')"

  if [[ -z "${global_node_ip}" ]]; then
    printf 'Error: unable to resolve node IP for cluster %s\n' "${global_cluster}" >&2
    exit 1
  fi

  if [[ -z "${global_kds_port}" ]]; then
    printf 'Error: unable to resolve KDS NodePort from kuma-global-zone-sync service\n' >&2
    exit 1
  fi

  printf 'grpcs://%s:%s' "${global_node_ip}" "${global_kds_port}"
}

single_up() {
  local cluster_name="${1:-kuma-1}"
  local kubeconfig_path
  kubeconfig_path="$(kubeconfig_for_cluster "${cluster_name}")"

  ensure_cluster_started "${cluster_name}"
  load_images_into_cluster_once "${cluster_name}"
  deploy_helm "${cluster_name}" "${kubeconfig_path}" zone

  cat <<EOF
Single-zone cluster is ready.
Use kubeconfig:
  export KUBECONFIG="${kubeconfig_path}"
EOF
}

single_down() {
  local cluster_name="${1:-kuma-1}"
  stop_cluster_if_exists "${cluster_name}"
}

deploy_global() {
  local global_cluster="$1"
  local global_kubeconfig="$2"
  local extra_settings="${K3D_HELM_DEPLOY_ADDITIONAL_SETTINGS:-}"
  local global_settings="controlPlane.mode=global controlPlane.globalZoneSyncService.type=NodePort"

  if [[ -n "${extra_settings}" ]]; then
    global_settings="${global_settings} ${extra_settings}"
  fi

  ensure_cluster_started "${global_cluster}"
  load_images_into_cluster_once "${global_cluster}"
  deploy_helm "${global_cluster}" "${global_kubeconfig}" global "${global_settings}"
}

deploy_zone() {
  local zone_cluster="$1"
  local zone_kubeconfig="$2"
  local zone_name="$3"
  local global_kds_address="$4"
  local extra_settings="${K3D_HELM_DEPLOY_ADDITIONAL_SETTINGS:-}"
  local zone_settings="controlPlane.mode=zone controlPlane.zone=${zone_name} controlPlane.kdsGlobalAddress=${global_kds_address} controlPlane.tls.kdsZoneClient.skipVerify=true"

  if [[ -n "${extra_settings}" ]]; then
    zone_settings="${zone_settings} ${extra_settings}"
  fi

  ensure_cluster_started "${zone_cluster}"
  load_images_into_cluster_once "${zone_cluster}"
  deploy_helm "${zone_cluster}" "${zone_kubeconfig}" zone "${zone_settings}"
}

global_zone_up() {
  local global_cluster="${1:-kuma-1}"
  local zone_cluster="${2:-kuma-2}"
  local zone_name="${3:-zone-1}"
  local global_kubeconfig
  local zone_kubeconfig
  local global_kds_address

  global_kubeconfig="$(kubeconfig_for_cluster "${global_cluster}")"
  zone_kubeconfig="$(kubeconfig_for_cluster "${zone_cluster}")"

  ensure_metallb_manifest "${global_cluster}"
  ensure_metallb_manifest "${zone_cluster}"

  deploy_global "${global_cluster}" "${global_kubeconfig}"

  global_kds_address="$(resolve_global_kds_address "${global_cluster}" "${global_kubeconfig}")"

  deploy_zone "${zone_cluster}" "${zone_kubeconfig}" "${zone_name}" "${global_kds_address}"

  cat <<EOF
Multi-zone clusters are ready.
Global KDS address:
  ${global_kds_address}
Global kubeconfig:
  export KUBECONFIG_GLOBAL="${global_kubeconfig}"
Zone kubeconfig:
  export KUBECONFIG_ZONE="${zone_kubeconfig}"
EOF
}

global_zone_down() {
  local global_cluster="${1:-kuma-1}"
  local zone_cluster="${2:-kuma-2}"

  stop_cluster_if_exists "${zone_cluster}"
  stop_cluster_if_exists "${global_cluster}"
}

global_two_zones_up() {
  local global_cluster="${1:-kuma-1}"
  local zone1_cluster="${2:-kuma-2}"
  local zone2_cluster="${3:-kuma-3}"
  local zone1_name="${4:-zone-1}"
  local zone2_name="${5:-zone-2}"
  local global_kubeconfig
  local zone1_kubeconfig
  local zone2_kubeconfig
  local global_kds_address

  global_kubeconfig="$(kubeconfig_for_cluster "${global_cluster}")"
  zone1_kubeconfig="$(kubeconfig_for_cluster "${zone1_cluster}")"
  zone2_kubeconfig="$(kubeconfig_for_cluster "${zone2_cluster}")"

  ensure_metallb_manifest "${global_cluster}"
  ensure_metallb_manifest "${zone1_cluster}"
  ensure_metallb_manifest "${zone2_cluster}"

  deploy_global "${global_cluster}" "${global_kubeconfig}"

  global_kds_address="$(resolve_global_kds_address "${global_cluster}" "${global_kubeconfig}")"

  deploy_zone "${zone1_cluster}" "${zone1_kubeconfig}" "${zone1_name}" "${global_kds_address}"
  deploy_zone "${zone2_cluster}" "${zone2_kubeconfig}" "${zone2_name}" "${global_kds_address}"

  cat <<EOF
Three-cluster multi-zone setup is ready.
Global KDS address:
  ${global_kds_address}
Global kubeconfig:
  export KUBECONFIG_GLOBAL="${global_kubeconfig}"
Zone-1 kubeconfig:
  export KUBECONFIG_ZONE1="${zone1_kubeconfig}"
Zone-2 kubeconfig:
  export KUBECONFIG_ZONE2="${zone2_kubeconfig}"
EOF
}

global_two_zones_down() {
  local global_cluster="${1:-kuma-1}"
  local zone1_cluster="${2:-kuma-2}"
  local zone2_cluster="${3:-kuma-3}"

  stop_cluster_if_exists "${zone2_cluster}"
  stop_cluster_if_exists "${zone1_cluster}"
  stop_cluster_if_exists "${global_cluster}"
}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

# Parse --repo-root before subcommand
while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo-root)
      repo_root="$2"
      shift 2
      ;;
    --help)
      usage
      exit 0
      ;;
    *)
      break
      ;;
  esac
done

if [[ -z "${repo_root}" ]]; then
  repo_root="$(cd "${script_dir}/../../../.." && pwd)"
  printf 'Warning: --repo-root not specified, falling back to %s\n' "${repo_root}" >&2
fi

if [[ ! -f "${repo_root}/go.mod" ]]; then
  printf 'Error: %s does not look like a Kuma repo (no go.mod)\n' "${repo_root}" >&2
  exit 1
fi

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

command_name="$1"
shift || true

case "${command_name}" in
  single-up)
    single_up "$@"
    ;;
  single-down)
    single_down "$@"
    ;;
  global-zone-up)
    global_zone_up "$@"
    ;;
  global-zone-down)
    global_zone_down "$@"
    ;;
  global-two-zones-up)
    global_two_zones_up "$@"
    ;;
  global-two-zones-down)
    global_two_zones_down "$@"
    ;;
  *)
    usage
    exit 1
    ;;
esac
