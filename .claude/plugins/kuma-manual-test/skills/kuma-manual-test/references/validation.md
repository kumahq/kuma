# Manifest validation

Pre-apply checks for every resource or policy manifest.

For `Mesh*` policy specifics (roles, `targetRef` rules, inspect flow), read
`references/mesh-policies.md`.

## Pre-apply checklist

- Confirm all `kind` values exist in the live API (`kubectl explain`)
- Confirm namespace and labels match test intent
- Confirm required fields and enum values are valid for the cluster's CRDs
- If the manifest references other resources, confirm those names exist in the same mesh and expected namespace
- Run server-side dry-run and block on any failure

## Validate

```bash
"$SKILL_DIR/scripts/validate-manifest.sh" \
  --kubeconfig "${KUBECONFIG}" \
  --manifest "<manifest-file>"
```

Validation is not optional. Never pass `--validate=false` to kubectl - validation failures indicate a bug in the manifest or missing CRDs. Fix the root cause instead.

## Safe apply flow

1. Write manifest to `${RUN_DIR}/manifests/` - never to `/tmp` or anywhere else. Copy from suite baseline/groups or write inline YAML directly to the run manifests dir. When a suite group provides inline YAML, use it verbatim without modifications.
2. Validate with `"$SKILL_DIR/scripts/validate-manifest.sh"`.
3. Apply with `"$SKILL_DIR/scripts/apply-tracked-manifest.sh"`.
4. Run verification commands and record each one via `record-command.sh`, saving output to `artifacts/`.
5. Record artifacts in report. Every artifact path must resolve to a real file.

## Tracked apply example

```bash
"$SKILL_DIR/scripts/apply-tracked-manifest.sh" \
  --run-dir "${RUN_DIR}" \
  --kubeconfig "${KUBECONFIG}" \
  --manifest "<manifest-file>" \
  --step "<step-name>"
```

This creates a versioned manifest copy, logs validation and apply output, and updates manifest and command indexes.
