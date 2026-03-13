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
"${CLAUDE_SKILL_DIR}/scripts/validate-manifest.sh" \
  --kubeconfig "${KUBECONFIG}" \
  --manifest "<manifest-file>"
```

Validation is not optional. Never pass `--validate=false` to kubectl - validation failures indicate a bug in the manifest or missing CRDs. Fix the root cause instead.

## Safe apply flow

1. Write manifest to `${RUN_DIR}/manifests/` - never to `/tmp` or anywhere else. Copy from suite baseline/groups or write inline YAML directly to the run manifests dir. When a suite group provides inline YAML, use it verbatim without modifications.
2. Validate with `"${CLAUDE_SKILL_DIR}/scripts/validate-manifest.sh"`.
3. If validation fails, follow the manifest error handling flow below.
4. Apply with `"${CLAUDE_SKILL_DIR}/scripts/apply-tracked-manifest.sh"`.
5. If apply fails, follow the manifest error handling flow below.
6. Run verification commands THROUGH `"${CLAUDE_SKILL_DIR}/scripts/record-command.sh"`, saving output to `artifacts/`.
7. Record artifacts in report. Every artifact path must resolve to a real file.

## Manifest error handling

When a suite manifest fails validation or apply, use AskUserQuestion with these options:

1. **"Fix for this run only"** - Write the corrected manifest to `${RUN_DIR}/manifests/`, record a deviation in the run report (what changed, why, user-approved). The suite files stay unchanged. Future runs will hit the same error.

2. **"Fix in suite and this run"** - Fix the manifest in both places:
   - Update the suite file (group file in `${SUITE_DIR}/groups/` or baseline in `${SUITE_DIR}/baseline/`) with the corrected YAML.
   - Append an entry to `${SUITE_DIR}/amendments.md` (create the file if it doesn't exist). Format:
     ```
     ---

     **<date>** | Run: `<run-id>` | <group>

     - **File**: `<relative path within suite>`
     - **Change**: <what was fixed>
     - **Reason**: <why it was wrong>
     ```
   - Write the corrected manifest to `${RUN_DIR}/manifests/` and proceed with apply.
   - Record a deviation in the run report noting both the run fix and the suite amendment.

3. **"Skip this step"** - Skip the step, record it as skipped in the report with the error details.

Always include the validation/apply error message in the AskUserQuestion description so the user can make an informed decision. Do not attempt to fix manifests without asking first - the error might reveal a real product bug rather than a suite authoring mistake.

## Tracked apply example

```bash
"${CLAUDE_SKILL_DIR}/scripts/apply-tracked-manifest.sh" \
  --run-dir "${RUN_DIR}" \
  --kubeconfig "${KUBECONFIG}" \
  --manifest "<manifest-file>" \
  --step "<step-name>"
```

This creates a versioned manifest copy, logs validation and apply output, and updates manifest and command indexes.
