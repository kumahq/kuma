# Rethink the releasing pipeline to release faster

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/16973

## Context and Problem Statement

Pushing a `vX.Y.Z` tag runs the full `build-test-distribute.yaml`, including the
entire e2e suite, before any artifact ships. Recent tag runs: v2.13.8 218m,
v2.14.0 124m, v2.12.12 119m, v2.11.16 102m (~2–3.5h), on top of the manual steps in
`.claude/rules/release.md`.

On a tag push (`FULL_MATRIX=true`):

- `check` ~13m (lint, `make check`, SBOM/CVE).
- `test`: `test_unit` ~23m plus the full e2e matrix, ~17 jobs (k8s / universal /
  multizone / gatewayapi × versions × amd64/arm64 × CNI × IPv6) at `max-parallel: 5`,
  ~25–35m each. This is the long pole.
- `build_publish` is `needs: [check, test]`, so it waits for the whole matrix;
  the build itself is cheap (binaries ~13m, images ~6–8m parallel).
- `provenance` (tags only) plus `distributions`.

The core inefficiency: the tag is HEAD of `release-X.Y`, which already ran the full
matrix (release prereq: "CI green on `release-X.Y`"). The tree is byte-identical, yet
we re-run ~90–150m of e2e and gate all publishing behind it. Secondary problem: e2e is
the flakiest CI stage, so a flake can fail an otherwise-good release.

One constraint shapes the options. The version is ldflags-injected from the git tag
(`version.sh`, `mk/build.mk`), so release-branch artifacts (`2.14.1-<hash>`) cannot be
byte-promoted to GA (`2.14.0`). Any fix either rebuilds on the tag or changes how the
version is derived.

Goals: tag-to-artifacts well under an hour; no weaker supply-chain
(SBOM/provenance/signing/scan); fewer release failures from flakes or operator error.
Non-goals: changing cadence/branching/tag format; adopting external infra (Prow/GCB).

## Design

### How peer projects release

Axis = does the full test suite re-run on the tag?

| Project | Re-tests on tag? | Mechanism |
|:--------|:-----------------|:----------|
| Kubernetes | No | `krel stage`→`release`; images promoted by **digest** |
| Istio | No | `release-builder` build/publish split; copy staging→prod |
| Envoy | No | tag job builds+signs+publishes only; SHA idempotency guard |
| Cilium | No | tag job builds+signs only; gate = branch CI + RC + protected env |
| cert-manager | No | `cmrel stage`/`publish`; e2e gated on the branch |
| Knative | No (promote) | `--from-latest-nightly` retags nightly, `SKIP_TESTS=1` |
| Argo CD | No | release workflow runs zero tests; 7-week RC freeze |
| **Linkerd** | **Yes** | rebuilds + runs full integration suite on tag |

Only Linkerd re-tests on the tag, which is Kuma's current model. Everyone else gates on
the branch/RC and makes the tag build-and-publish only. Envoy and Cilium do it natively
on GitHub Actions, the closest templates for us.

### Option 1 — Skip e2e on the tag; gate on the branch (recommended)

- No e2e (and no e2e smoke) on the tag. Keep `test_unit` and container-structure
  tests, which are fast and validate the artifact. Drop the ~90–150m matrix.
- Build and publish without the e2e gate, so the build starts right after the gate.
- Hard-enforce a green tagged SHA. The first job asserts the commit has a passing
  `build-test-distribute` run on its branch (`gh api` check-runs), else it fails.

After this, green-SHA gate ~1m, `build_publish` ~15m, plus provenance/distributions,
lands at roughly 20–30m versus the current 100–220m.

- (+) Answers the issue; no new infra, staging, or version change; removes the flakiest stage.
- (−) Loses the exact-tag re-test (mitigated by the hard green-SHA gate, RC soak, and
  rehearsal); still rebuilds on the tag (~15m, fine).

### Option 2 — RC + promotion by digest (build once, ship the tested bits)

`vX.Y.Z-rc.N` builds and tests once; GA promotes the RC artifacts by digest with no
rebuild or retest. This is the k8s/Istio gold standard.

- Blocker: the version is baked into the binary, so RC images carry `-rc.N`. Promotion
  needs the version decoupled from the build (runtime or retag), a larger change to
  `version.sh`, the build, and the registry.
- (+) The tested bits are the shipped bits, with RC soak. (−) Needs a staging registry,
  promotion tooling, version rework, and a new RC cadence.

### Option 3 — Stage/publish split via one resumable release tool

`stage` (build once, pinned) and `publish` (promote), as numbered idempotent steps in
one tool (extend `release-tool`), like Envoy and `cmrel`.

- (+) Fixes the *error-prone* half: resumable, decoupled publishing, one RC/GA path.
- (−) Medium effort, and orthogonal to the e2e waste, so it layers on top of Option 1.

### Option 4 — Status quo (rejected)

Keep re-testing on the tag. This is the problem itself, and the lone outlier among peers.

### Implementation: dedicated release workflow vs. conditionals (orthogonal to Options 1–4)

Today the release path lives in `build-test-distribute.yaml` via scattered
`ref_type=='tag'` / `FULL_MATRIX` / `ALLOW_PUSH` conditionals.

- A, more conditionals in the shared workflow. (+) smallest diff. (−) more
  tag-vs-PR branching in an already-conditional workflow; release stays hard to read;
  shared concurrency/permissions.
- B, a dedicated `release-publish.yaml` (`tags: ["v*"]`) reusing
  `_build_publish.yaml` and `_provenance.yaml`, never `_test.yaml`; drop the tag trigger
  from `build-test-distribute.yaml`; `release.yaml` keeps changelog/version
  bookkeeping. (+) linear readable release flow; own concurrency/permissions; no
  duplicated build logic; matches the Envoy/Cilium/Argo pattern; natural host for the
  rehearsal and a future RC. (−) new file plus a one-time trigger migration.

Picked B.

### Release rehearsal (scheduled dry-run)

Moving the gate off the tag introduces a risk: the release machinery breaks and we find
out only at release time. A scheduled rehearsal de-risks it. Checking the
Makefiles, `version.sh`, `helm.sh`, the workflows, and the tool docs shows this is mostly
config, not new tooling:

- Already runs on every master and `release-X.Y` push: `version.sh` stamps a
  `-preview.v<hash>` version that auto-routes to `notary-internal` and preview
  Cloudsmith, and `ALLOW_PUSH=true` triggers a real build, push, sign, and pulp plus a
  helm *package* via `_build_publish.yaml`. The rehearsal just decouples this from e2e
  and schedules it.
- Provenance: SLSA generators v2.1.0 support `schedule`, `push`, and `workflow_dispatch`
  (not `pull_request`); the generic one still emits `.intoto.jsonl` when
  `upload-assets=false` (already wired as `upload-assets: ref_type=='tag'`); containers
  push to an alt repo. So we loosen the job `if`, no new mode.
- Sign, SBOM, and scan (Kong actions): `signature_registry` is free-form (already
  `notary-internal`), OIDC signing is keyless (no tag needed), and
  `upload-sbom-release-assets` defaults to false. Already off-tag ready.
- Helm release (`cr` 1.8.1): `cr upload` and `cr index` target any owner/repo via flags
  (`helm.sh` routes `GH_OWNER`/`GH_REPO`). Point them at a throwaway test charts repo
  (plus a token).

Excluded by design: a real GitHub Release on `kumahq/kuma`, and prod
registry/notary/Cloudsmith. Caveats: run on `schedule` or `workflow_dispatch`, and
provision a test charts repo. This follows the nightly-build pattern (Istio, Knative,
k8s `krel stage`, `cmrel --mock`). It complements the existing `release.yaml` daily
`--check`, which verifies the *last* release's prod artifacts.

### Evolution

This started from the issue's "skip the full suite on the tag." Peer research confirmed
it is the industry norm; only Linkerd re-tests. The bigger wins, digest promotion and
the stage/publish split, depend on a version rework and an RC process Kuma lacks. So
Option 1 ships now, with Options 2 and 3 as follow-up.

## Security implications and review

- e2e is not a security control; SBOM, provenance, signing, and scan live in
  `build_publish` and `provenance`, which Option 1 leaves unchanged.
- New risk: tagging a never-green commit, mitigated by the hard green-SHA gate.
- Container-structure tests stay, so artifact regressions are still caught on the tag.
- Option 2 *strengthens* security, since the scanned and signed bits are the shipped bits.

## Reliability implications

- Faster releases mean faster security-patch turnaround.
- Dropping the redundant e2e removes the flakiest release gate.
- Residual risk: the branch run goes stale against the tag. The exact green-SHA gate,
  RC soak, or a one-off branch matrix before tagging mitigate this.
- `test_unit` and container-structure tests act as a fast tag tripwire.
- The rehearsal continuously validates build, publish, and sign (plus provenance and
  helm once the gates loosen). Only the real GitHub Release and prod routing stay
  tag-only.

## Implications for Kong Mesh

- The downstream fork mirrors this pipeline; port "skip e2e plus green-SHA gate" directly.
- For Option 2 later, the downstream needs a matching staging registry and a version
  change, and syncs by promoting rather than rebuilding.
- Drop "wait for full suite on tag" from the downstream runbooks.

## Decision

Adopt Option 1, as a dedicated release workflow plus a scheduled rehearsal:

- A dedicated `release-publish.yaml` reusing `_build_publish.yaml` and
  `_provenance.yaml`, never `_test.yaml`; drop the tag trigger from
  `build-test-distribute.yaml`.
- No e2e or smoke on the tag; keep `test_unit` and container-structure tests.
- Hard-enforce a green tagged SHA.
- A scheduled rehearsal of build, publish, sign, SBOM, scan, and provenance (loosen
  the tag-only `if`s, keep the preview/internal targets); route the helm release to a
  test charts repo; run on `schedule` or `workflow_dispatch`.

This cuts ~100–220m down to ~20–30m with no new infra. Options 2 (RC plus digest
promotion) and 3 (stage/publish split) are follow-up; RC is deferred; version
decoupling is undecided.

## Notes

Resolved: no smoke e2e on the tag; the green-SHA gate is hard-enforced; RC tags deferred.

Open: version decoupling (undecided; rebuild-on-tag is fine for Option 1); the exact
green-SHA mechanism and rehearsal cadence/scope (impl detail); provisioning a test
charts repo and token for the helm rehearsal leg.

Sources. Peer projects: Kubernetes (sig-release handbook, promo-tools), Istio (release-builder,
daily-release), Envoy (RELEASES.md, `_publish_build.yml`), Cilium
(`build-images-releases.yaml`), cert-manager (`cmrel`), Knative (`hack/release.sh`),
Argo CD (`release.yaml`), Linkerd (`release.yml`). Tools: slsa-github-generator
generic+container READMEs; helm/chart-releaser `cr` flags; Kong/public-shared-actions
security-actions (`upload-sbom-release-assets` default false, free-form
`signature_registry`).
