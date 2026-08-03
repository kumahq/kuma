# Kuma Service Mesh

## Overview

CNCF service mesh (Envoy-based) for K8s/VMs. L4-L7 connectivity, security, observability, multi-zone/multi-mesh.

**Tech Stack:** Go 1.23+, Envoy, Kubernetes/VMs, Ginkgo/Gomega, protobuf, mise

**Purpose:** Universal service mesh supporting both Kubernetes and VMs with automatic mTLS, traffic control, observability, and multi-zone deployments. Built by Kong, donated to CNCF.

---

## Project Structure

```
/app/                    # Main apps (kuma-cp, kuma-dp, kumactl, cni, kuma-ui)
/pkg/
  core/                  # Resources, managers, validators, plugins
  xds/                   # Envoy xDS (cache, envoy/, server)
  kds/                   # Multi-zone sync (Kuma Discovery Service)
  plugins/policies/      # 20+ policies (api/, plugin/, xds/)
  config/api-server/dp-server/  # Config, APIs, servers
  test/                  # Test utilities, matchers, builders
/test/                   # E2E and integration tests
/docs/madr/              # Architecture Decision Records
```

**Key Modules:**
- **Control Plane** (`pkg/core/`, `app/kuma-cp/`) - xDS, mTLS, multi-zone
- **Policies** (`pkg/plugins/policies/<name>/`) - api/v1alpha1/, plugin/, xds/
- **XDS** (`pkg/xds/`) - Envoy config generation, dual-tier cache

**Code Generation:** Run `make generate` after changes to `.proto`, `pkg/plugins/policies/*/api/`, resource definitions. Generated: `zz_generated.*`, `*.pb.go`

---

## Architecture & Domain

### Components

- **kuma-cp** - Control plane (xDS, mTLS, multi-zone coordinator)
- **kuma-dp** - Data plane (Envoy wrapper)
- **kumactl** - CLI
- **kuma-cni** - CNI plugin

**Multi-zone:** Global CP → Zone CPs via KDS • Zone Ingress/Egress for cross-zone • K8s + Universal (VMs)

### Core Concepts

| Term | Definition | Usage |
|------|------------|-------|
| **Dataplane** | Envoy proxy instance per workload | Sidecar for each pod/VM |
| **Mesh** | Isolated namespace for resources | `default` mesh, can have multiple |
| **ZoneTag** | `kuma.io/zone` auto-added by CP | Marks dataplane origin zone |
| **TargetRef** | Policy selector | `Mesh`, `MeshService`, `MeshSubset` |
| **FromRules** | Inbound policies | TrafficPermission from services |
| **ToRules** | Outbound policies | RetryPolicy to dependencies |
| **MeshContext** | Cached config for xDS | Dual-tier cache (short TTL + hash) |
| **xDS** | Envoy discovery protocol | CDS/EDS/LDS/RDS |
| **KDS** | Kuma Discovery Service | Multi-zone sync |
| **Configurer** | Strategy for Envoy config | `TimeoutConfigurer`, `CircuitBreakerConfigurer` |
| **MatchedPolicies** | Policies for a dataplane | Based on TargetRef selectors |

### Policy Structure

```
pkg/plugins/policies/<policy-name>/
  api/v1alpha1/           # <policy>_types.go, validator.go
  plugin/v1alpha1/        # plugin.go (application logic)
  xds/                    # configurer.go (Envoy config)
```

**Example:**
```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
spec:
  targetRef: {kind: MeshService, name: backend}
  to:
    - targetRef: {kind: MeshService, name: db}
      default: {connectionTimeout: 5s, idleTimeout: 60s}
```

### xDS Protocol

**CDS** (Clusters) - Backend services
**EDS** (Endpoints) - Service endpoints
**LDS** (Listeners) - Inbound/outbound listeners
**RDS** (Routes) - HTTP routes

---

## Commands

### Development

```bash
make install              # mise install
make build                # Build all
make check                # Lint/format (auto-fix)
make generate             # Gen code (proto, policies, resources)
make test                 # Run tests
make test/e2e             # E2E (slow)

make build/kuma-cp        # Build control plane
make test TEST_PKG_LIST=./pkg/xds/...    # Test specific package
UPDATE_GOLDEN_FILES=true make test       # Update golden files

make k3d/restart && skaffold dev          # Dev environment
```

### Git & PRs

```bash
git push --no-verify upstream branch    # ALWAYS use --no-verify
```

**Branches:** `master` (base), `release-{2.7,2.10,2.11,2.12}`
**Versions:** 2.7.19 (LTS), 2.10.8, 2.11.7, 2.12.3

**Commit Format:** [Conventional Commits](https://www.conventionalcommits.org/)
```
<type>(<scope>): <subject>

<body>

Fixes #123
```

**Types:** `feat`, `fix`, `chore`, `docs`, `test`, `refactor`, `ci`
**Scopes:** `kuma-cp`, `kuma-dp`, `kumactl`, `xds`, `kds`, `MeshRetry`, `api-server`
**Sign:** Use `-s -S` flags

**PR:** Base: `master` • Template: `.github/PULL_REQUEST_TEMPLATE.md` • Changelog: from title or `> Changelog:` • Labels: `ci/*` at creation • MADR: `docs/madr/decisions/000-template.md` • No Kong Mesh mentions (use "downstream project")

---

## Gotchas & Debugging

### Common Issues

**Build:** `make install` (dependencies) • `make generate` (outdated code) • `make check` (formatting)
**Test:** `UPDATE_GOLDEN_FILES=true make test` (golden mismatch) • `make test/e2e` (cluster setup)
**Multi-zone:** Test K8s + Universal • Verify isolation • Validate cross-zone carefully

### Debugging

**Temp logs:** Prefix "DEBUG" for easy removal

**Envoy config:**
```bash
kubectl exec deploy/<name> -c kuma-sidecar -- \
  wget -qO- localhost:9901/config_dump | jq '.configs[]'
```

### Performance & Security

**Performance:** Minimize allocations (many dataplanes) • Optimize DB queries • XDS efficiency critical
**Security:** Validate external inputs • Careful with mTLS/certs • Follow RBAC implications • Secure secrets handling

### Tool Management

**mise:** Config: `mise.toml` • Pinned versions • `make install` or `mise install`
**Tools:** buf, ginkgo, helm, kind, kubectl, protoc, yq, golangci-lint, skaffold

---

## Resources

- **DEVELOPER.md** - Setup, testing details
- **CONTRIBUTING.md** - PR workflow, commit format
- **docs/madr/decisions/** - Architecture decisions
- **Kuma Docs** - https://kuma.io/docs
- **Slack** - https://kuma-mesh.slack.com
