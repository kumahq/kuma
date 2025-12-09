---
applyTo:
  - "pkg/plugins/policies/**"
---

# Policy Development Guidelines

## Policy Structure

Every policy follows this structure:

```
pkg/plugins/policies/<policy-name>/
  api/v1alpha1/           # <policy>_types.go, validator.go
  plugin/v1alpha1/        # plugin.go (application logic)
  xds/                    # configurer.go (Envoy config)
```

## Development Checklist

**API (`api/v1alpha1/`):**
- Complete validator implementation
- All TargetRef kinds supported (`Mesh`, `MeshService`, `MeshSubset`)
- Inbound (`from`) and outbound (`to`) rules if applicable
- Test with table-driven tests for validation

**Plugin (`plugin/v1alpha1/`):**
- Application logic for policy matching
- Follows existing plugin patterns
- Minimal, focused implementation

**XDS (`xds/`):**
- Configurer for Envoy config generation
- Protocol handling (HTTP, TCP, gRPC)
- Efficient ResourceSet usage
- Metadata for debugging
- Golden file tests

## Before Implementation

1. Search for similar policies in `pkg/plugins/policies/`
2. Study existing matcher/configurer patterns
3. Verify K8s + Universal compatibility
4. Test inbound + outbound scenarios

## Review Focus

- Complete validator with edge cases
- api/plugin/xds structure correct
- All TargetRef kinds supported
- Protocol-specific handling
- K8s + Universal tested
