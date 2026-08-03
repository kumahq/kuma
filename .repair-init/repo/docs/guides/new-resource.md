# Creating New Resources with policy-gen

## Quick Start

1. Bootstrap Resource

For resources (non-policies):
```bash
go run ./tools/policy-gen/bootstrap \
  --name YourResource \
  --path pkg/core/resources/apis \
  --version v1alpha1 \
  --is-policy=false \
  --has-status \
  --force
```

2. Edit Generated Files

Add markers above your struct in `api/v1alpha1/<resource>.go`:

```go
// YourResource
// +kuma:policy:is_policy=false
// +kuma:policy:has_status=true
// +kuma:policy:short_name=yr
// +kuma:policy:is_referenceable_in_to=true
type YourResource struct {
    // your fields
}
```

Edit implementation:
- `api/v1alpha1/<resource>.proto` - struct definition + markers
- `api/v1alpha1/validator.go` - validation logic
- `plugin/v1alpha1/plugin.go` - Apply() method (policies only)

3. Generate Code

```bash
make generate
```

4. Update tests.

```shell
make test UPDATE_GOLDEN_FILES=true
```


## Common Markers

### Resource Configuration
- `+kuma:policy:is_policy=false` - resource vs policy
- `+kuma:policy:has_status=true` - adds status field
- `+kuma:policy:short_name=msvc` - kubectl shortname
- `+kuma:policy:scope=Mesh|Global` - resource scope
- `+kuma:policy:skip_registration=true` - skip auto-registration

### KDS & References
- `+kuma:policy:kds_flags=model.ZoneToGlobalFlag` - KDS sync direction
- `+kuma:policy:is_destination=true` - destination resource
- `+kuma:policy:is_referenceable_in_to=true` - can be in targetRef.to

### Kubernetes
- `+kubebuilder:printcolumn:JSONPath=".spec.field",name=Field,type=string` - kubectl columns
- `+kuma:policy:allowed_on_system_namespace_only=true` - system namespace only

## Bootstrap Flags

- `--name` - resource name (UpperCamelCase, required)
- `--path` - base path for generated code
- `--version` - API version (default: v1alpha1)
- `--is-policy` - mark as policy vs resource
- `--has-status` - add status field
- `--generate-target-ref` - add top-level targetRef
- `--generate-to` - add 'to' array
- `--generate-from` - add 'from' array
- `--force` - overwrite existing code
- `--skip-validator` - skip validator generation

## Generated Files

**Manual files (edit these):**
- `api/v1alpha1/<resource>.proto` - struct with markers
- `api/v1alpha1/validator.go` - validation
- `plugin/v1alpha1/plugin.go` - implementation (policies)

**Generated files (don't edit):**
- `api/v1alpha1/zz_generated.resource.go` - core resource
- `api/v1alpha1/zz_generated.deepcopy.go` - deep copy
- `api/v1alpha1/zz_generated.helpers.go` - helpers (policies)
- `api/v1alpha1/schema.yaml` - OpenAPI schema
- `k8s/v1alpha1/zz_generated.types.go` - K8s CRD types
- `k8s/crd/kuma.io_<resource>.yaml` - CRD definition
- `zz_generated.plugin.go` - plugin registration

## Tool Locations

- Bootstrap: `tools/policy-gen/bootstrap/main.go`
- Generator: `tools/policy-gen/generator/`
- Helper scripts: `tools/policy-gen/*.sh`

## Make Targets

- `make generate` - regenerate everything
- `make generate/resources` - all resources
- `make generate/policy/<name>` - specific policy/resource

## Tips

1. Use bootstrap to create skeleton, not manual files
2. Markers control generation - don't edit `zz_generated.*` files
3. Run `make generate` after changing markers/fields
4. For policies: implement Apply() in plugin.go
5. Check existing resources (meshservice, meshexternalservice) as examples
6. Use `--force` carefully - deletes existing code
