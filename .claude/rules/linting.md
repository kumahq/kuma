# Linting rules and import conventions

## Required import aliases

The `importas` linter enforces these aliases. Using wrong names fails `make check`:

- `core_mesh` → `github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh`
- `mesh_proto` → `github.com/kumahq/kuma/v2/api/mesh/v1alpha1`
- `system_proto` → `github.com/kumahq/kuma/v2/api/system/v1alpha1`
- `util_proto` → `github.com/kumahq/kuma/v2/pkg/util/proto`
- `util_rsa` → `github.com/kumahq/kuma/v2/pkg/util/rsa`
- `kuma_cmd` → `github.com/kumahq/kuma/v2/pkg/cmd`
- `bootstrap_k8s` → `github.com/kumahq/kuma/v2/pkg/plugins/bootstrap/k8s`
- `config_core` → `github.com/kumahq/kuma/v2/pkg/config/core`
- `core_model` → `github.com/kumahq/kuma/v2/pkg/core/resources/model`
- `common_api` → `github.com/kumahq/kuma/v2/api/common/v1alpha1`
- `api_types` → `github.com/kumahq/kuma/v2/api/openapi/types`

## Import ordering

Enforced by `gci`: standard library → third-party → `github.com/kumahq/kuma/v2`

## Forbidden patterns

Use `tracing.SafeSpanEnd(span)` instead of `span.End()`. The `forbidigo` linter blocks direct `span.End()` calls to prevent panics during OTel init/shutdown.

## Blocked packages (depguard)

- `github.com/golang/protobuf` → use `google.golang.org/protobuf` (except for JSON, see next line)
- `google.golang.org/protobuf/encoding/protojson` → use `github.com/golang/protobuf/jsonpb` (compatibility issues)
- `sigs.k8s.io/controller-runtime/pkg/log` → use `sigs.k8s.io/controller-runtime` (data race in init containers, see #13299)
- `io/ioutil` → use `io` and `os`
- `github.com/kumahq/kuma/v2/app` from `pkg/`. Architectural boundary (`pkg/` cannot import `app/`)

## RBAC validation

`make check` runs `check/rbac`: if any RBAC manifests in `deployments/` change (Role, RoleBinding, ClusterRole, ClusterRoleBinding), `UPGRADE.md` must also be updated.
