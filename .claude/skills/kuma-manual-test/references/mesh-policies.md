# Contents

1. [Authoring model](#1-authoring-model)
2. [Policy role checks](#2-policy-role-checks)
3. [targetRef guardrails](#3-targetref-guardrails)
4. [Safe apply flow](#4-safe-apply-flow)
5. [Debug flow](#5-debug-flow-when-policy-is-not-effective)
6. [Edge-case matrix](#6-edge-case-matrix-for-every-mesh-feature-test-plan)
7. [Common commands](#7-common-command-set-for-policy-focused-runs)

---

# Mesh\* policies: authoring, apply, debug

Use this reference when a test run includes any `Mesh*` policy.

This file is based on Kuma docs `2.13.x` pages (check kuma.io/docs/ for newer versions):

- https://kuma.io/docs/2.13.x/policies/introduction/
- https://kuma.io/docs/2.13.x/explore/inspect-api/
- https://kuma.io/docs/2.13.x/guides/consumer-producer-policies/
- https://kuma.io/docs/2.13.x/guides/targeting-meshhttproutes-in-supported-policies/
- https://kuma.io/docs/2.13.x/policies/meshaccesslog/
- https://kuma.io/docs/2.13.x/policies/meshmetric/
- https://kuma.io/docs/2.13.x/policies/meshtrace/

## 1) Authoring model

- Use only new `Mesh*` policies for a given feature area.
- Do not combine old and new policy families for the same feature in one test.
  - Example: avoid `MeshTrace` + `TrafficTrace` in one scenario.
- On Kubernetes, use `apiVersion: kuma.io/v1alpha1`.
- Namespace is part of policy behavior.
- Always set `metadata.labels["kuma.io/mesh"]` explicitly.

## 2) Policy role checks

Kuma policy role affects priority and multi-zone sync.

| Role           | Typical shape                                                       | Namespace and intent                               |
| -------------- | ------------------------------------------------------------------- | -------------------------------------------------- |
| producer       | has `spec.to` and targets service in same namespace                 | backend owner sets defaults for callers            |
| consumer       | has `spec.to` and targets service in other namespace (or label set) | caller owner overrides how caller reaches upstream |
| workload-owner | has `spec.rules`, or only `spec.targetRef/default`                  | owner configures own dataplane proxy behavior      |
| system         | policy created in system namespace (`kuma-system`)                  | operator-managed mesh-wide behavior                |

Verify labels after apply:

```bash
kubectl get <policy-kind> <name> -n <namespace> -o jsonpath='{.metadata.labels.kuma\.io/policy-role}'
kubectl get <policy-kind> <name> -n <namespace> -o jsonpath='{.metadata.labels.kuma\.io/origin}'
```

## 3) `targetRef` guardrails

- `targetRef.kind` must be valid for that policy and field.
- Use `name` + `namespace` for one target.
- Use `labels` only when intentionally targeting a set.
- If `namespace` is omitted in `targetRef`, Kuma uses policy namespace.
- `sectionName` can target a named section or a numeric port.
- For system policies, use label-based targeting.
- `MeshHTTPRoute` targeting is only valid for policies that support it.

Before writing test manifests, confirm supported kinds on the policy page in docs.

## 4) Safe apply flow

```bash
"${CLAUDE_SKILL_DIR}/scripts/validate-manifest.sh" \
  --kubeconfig "${KUBECONFIG}" \
  --manifest "<manifest-file>"

"${CLAUDE_SKILL_DIR}/scripts/apply-tracked-manifest.sh" \
  --run-dir "${RUN_DIR}" \
  --kubeconfig "${KUBECONFIG}" \
  --manifest "<manifest-file>" \
  --step "<step-name>"
```

Do not run raw `kubectl apply` for test manifests.

## 5) Debug flow when policy is not effective

1. Check object acceptance and stored shape.

```bash
kubectl get <policy-kind> <name> -n <namespace> -o yaml
```

2. Check impacted dataplanes from policy side.

```bash
KUMACTL="$("${CLAUDE_SKILL_DIR}/scripts/find-local-kumactl.sh")"
"${KUMACTL}" inspect <policy-resource-name> <name> --mesh <mesh>
```

`<policy-resource-name>` is the CLI resource form, for example `meshretry`.

3. Check matched policies from dataplane side.

```bash
"${KUMACTL}" inspect dataplane <dataplane-name> --mesh <mesh>
```

Look at all four attachment points:

- `Dataplane`
- `Inbound`
- `Outbound`
- `Service`

4. Check generated Envoy config.

```bash
"${KUMACTL}" inspect dataplane <dataplane-name> --mesh <mesh> --type=config-dump
```

In multi-zone, run `inspect --type=config-dump` against a zone control plane,
not global.

5. Check runtime logs.

```bash
kubectl logs -n <ns> <pod-name> -c kuma-sidecar --tail=300
kubectl logs -n kuma-system deploy/kuma-control-plane --tail=400
```

6. Check protocol assumptions.

- HTTP policies need HTTP listeners, not TCP proxy listeners.
- Confirm service protocol markers and `appProtocol` where used.
- For tracing, confirm protocol support on targeted traffic paths.

7. In multi-zone, check sync.

```bash
kubectl --kubeconfig "${KUBECONFIG_GLOBAL}" get zones
kubectl --kubeconfig "${KUBECONFIG_GLOBAL}" get zoneinsights -o yaml
```

## 6) Edge-case matrix for every Mesh\* feature test plan

Add these groups to the suite unless out of scope:

| Case                       | What to verify                                         | Artifacts                                |
| -------------------------- | ------------------------------------------------------ | ---------------------------------------- |
| baseline mesh-level policy | expected default behavior                              | apply log + runtime output               |
| specificity overrides      | mesh vs dataplane labels vs dataplane name/section     | `inspect dataplane` + config-dump        |
| producer vs consumer       | consumer override precedence in caller namespace       | policy labels + request behavior         |
| route-level targeting      | `MeshHTTPRoute` override on one route only             | route manifest + route-specific behavior |
| sectionName targeting      | named section and numeric section behavior             | config-dump listener/cluster match       |
| selector fan-out           | one label selector applying to many dataplanes         | `inspect <policy>` affected list         |
| invalid schema             | admission rejects wrong enum or shape                  | validation log                           |
| dangling reference         | accept/reject behavior and runtime effect              | apply + control-plane logs               |
| update and rollback        | config changes and restore behavior                    | before/after artifacts                   |
| delete semantics           | effective cleanup after delete                         | post-delete inspect/config-dump          |
| multi-zone propagation     | origin/sync and zone runtime behavior                  | `zones`, `zoneinsights`, zone logs       |
| protocol mismatch          | expected non-application when listener type mismatches | config-dump + negative result            |

## 7) Common command set for policy-focused runs

```bash
KUMACTL="$("${CLAUDE_SKILL_DIR}/scripts/find-local-kumactl.sh")"

kubectl api-resources --api-group kuma.io
kubectl get mesh -A
kubectl get dataplanes -A

"${KUMACTL}" inspect dataplane <dp-name> --mesh <mesh>
"${KUMACTL}" inspect dataplane <dp-name> --mesh <mesh> --type=config-dump
"${KUMACTL}" inspect <policy-resource-name> <policy-name> --mesh <mesh>
```

Record each command with `"${CLAUDE_SKILL_DIR}/scripts/record-command.sh"` when it is used as test artifacts.
