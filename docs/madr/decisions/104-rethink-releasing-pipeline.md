# Rethink the releasing pipeline to release faster

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/16973

## Context and Problem Statement

Pushing a `vX.Y.Z` tag runs the full `build-test-distribute.yaml`, including the
entire e2e suite, before any artifact ships. Recent tag runs: v2.13.8 218m,
v2.14.0 124m, v2.12.12 119m, v2.11.16 102m (~2–3.5h), on top of the manual steps in
the team-mesh release issues.

On a tag push (`FULL_MATRIX=true`):

- `check` ~13m (lint, `make check`, SBOM/CVE).
- `test`: `test_unit` ~23m plus the full e2e matrix, ~17 jobs (k8s / universal /
  multizone / gatewayapi × versions × amd64/arm64 × CNI × IPv6) at `max-parallel: 5`,
  ~25–35m each. This is the long pole.
- `build_publish` is `needs: [check, test]`, so it waits for the whole matrix;
  the build itself is cheap (binaries ~13m, independently reducible via build
  caching; images ~6–8m parallel).
- `provenance` (tags only) plus `distributions`.

The core inefficiency: the tag is HEAD of `release-X.Y`, which already ran the full
matrix (release prereq: "CI green on `release-X.Y`"). The tree is byte-identical, yet
we re-run ~90–150m of e2e and gate all publishing behind it. Secondary problem: e2e is
the flakiest CI stage, so a flake can fail an otherwise-good release.

One constraint shapes the options. The version is ldflags-injected from the git tag
(`version.sh`, `mk/build.mk`): a build on `release-2.14` is stamped as the next-patch
preview (`X.Y.Z-preview.<hash>`), not the version its GA tag carries (`X.Y.Z`), so
branch artifacts cannot be byte-promoted to GA. Any fix either rebuilds on the tag or
changes how the version is derived.

Goals: the tag re-runs no step already validated on the green release-branch SHA
(build and publish only, not re-test); no weaker supply-chain
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

- No re-test on the tag: drop unit and e2e. The source is identical to the green
  branch SHA, so re-running them adds nothing. Verify only what the tag build *changes*
  versus that SHA, the ldflags version, via the container-structure tests on the
  freshly-built image (already in `_build_publish.yaml`) plus an assert that the built
  binary reports the tag version.
- `check`'s lint/`make check`/SBOM is redundant for the same reason; only its metadata
  outputs (`IMAGES`/`REGISTRY`/`VERSION_NAME`) are needed downstream, so keep those.
- With e2e skipped, `test` completes immediately and `build_publish` publishes right
  away instead of waiting ~90m.
- Hard-enforce a green tagged SHA: the tagged commit must have a passing
  `build-test-distribute` run (full matrix) on its branch, else the release fails. This
  reuses the existing run, not new machinery.

After this, `build_publish` ~15m plus provenance/distributions lands at roughly 20–30m
versus the current 100–220m.

- (+) Answers the issue; no new infra, staging, or version change; removes the flakiest stage.
- (−) Loses the exact-tag re-test (mitigated by the hard green-SHA gate and, for risky
  releases, RC soak); still rebuilds on the tag (~15m, fine).

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

### Implementation: reuse the existing skip mechanism vs. a dedicated workflow

Orthogonal to Options 1–4: how to wire the tag pipeline. Today the release path lives in
`build-test-distribute.yaml`, which runs on every PR, master, and release-branch push,
so it builds and publishes a preview on every push.

- A, stay in the monolith and reuse the existing `ci/skip-test` mechanism. Tags don't
  skip today: `_test.yaml` keys the skip off PR labels
  (`github.event.pull_request.labels`), which are empty on a tag push, so the full matrix
  runs. Add `github.ref_type == 'tag'` next to that condition and e2e/unit skip on the
  tag; the existing `needs:` chain then lets `build_publish` proceed. (+) roughly a
  one-condition change, no new workflow; the publish path stays exercised on every push,
  so it cannot rot between releases. (−) the monolith keeps its tag-vs-PR conditionals.
- B, a dedicated `release-publish.yaml` on tags, reusing `_build_publish.yaml` and
  `_provenance.yaml`, never `_test.yaml`. (+) a linear, readable release flow. (−) it
  runs only on tags, so the publish path is no longer exercised between releases and can
  break unnoticed (the predictability the monolith gives for free); plus a new file and
  trigger migration.

Picked A. The monolith already publishes on every push, the strongest guard that the
release machinery works, and reusing `ci/skip-test` keeps the change tiny. (An earlier
draft picked B with a scheduled rehearsal to re-create that exercise; maintainer review
pointed out the monolith already provides it, making both the dedicated workflow and the
rehearsal unnecessary.)

### The publish path is already exercised continuously (no separate rehearsal)

Moving the gate off the tag raises a question: does the release machinery still get
exercised? It does, for free. `build-test-distribute.yaml` runs on every master and
`release-X.Y` push with `ALLOW_PUSH=true`, so every push already builds, pushes, signs,
and pulps a `-preview.<hash>` artifact through the same `_build_publish.yaml` the release
uses (`version.sh` auto-routes preview versions to `notary-internal` and the preview
Cloudsmith repo). That continuous run is the rehearsal; a separate scheduled one would be
redundant machinery.

The only steps not exercised between releases are the tag-gated ones: SLSA provenance
upload, the helm chart release to `kumahq/charts`, and the GitHub Release. They are first
exercised at release time. If that risk ever bites, revisit it with Option 2 rather than
standing up separate rehearsal machinery.

### Evolution

This started from the issue's "skip the full suite on the tag." Peer research confirmed
it is the industry norm; only Linkerd re-tests. An early draft added a dedicated
`release-publish.yaml` plus a scheduled rehearsal; maintainer review showed the monolith
already publishes a preview on every push, so reusing the existing `ci/skip-test`
mechanism is enough and the extra machinery was dropped. The bigger wins, digest
promotion and the stage/publish split, depend on a version rework and an RC process Kuma
lacks. So Option 1 ships now, with Options 2 and 3 as follow-up.

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
- Container-structure on the freshly-built image plus a binary-version assert are the
  only tag-specific checks; they verify exactly what the tag build changes (the
  ldflags), nothing the green branch run already covered.
- Every push already builds and publishes a preview through the release's own
  `_build_publish.yaml`, so the publish/sign path is validated continuously; only the
  tag-gated steps (provenance upload, helm release, GitHub Release) are first exercised
  at release time.

## Implications for Kong Mesh

- The downstream fork mirrors this pipeline; port "skip e2e plus green-SHA gate" directly.
- For Option 2 later, the downstream needs a matching staging registry and a version
  change, and syncs by promoting rather than rebuilding.
- Drop "wait for full suite on tag" from the downstream runbooks.

## Decision

Adopt Option 1, implemented in the existing `build-test-distribute.yaml`. No new
workflow, no scheduled rehearsal:

- Skip e2e and unit on the tag by extending the existing `ci/skip-test` condition with
  `github.ref_type == 'tag'`; the existing `needs:` chain then lets `build_publish`
  publish immediately.
- Verify only the ldflags delta: container-structure on the freshly-built image plus a
  binary-version assert. Keep `check`'s metadata computation, skip its redundant
  validation.
- Hard-enforce a green tagged SHA, reusing the commit's existing branch run.
- No separate rehearsal: every push already builds and publishes a preview through the
  same `_build_publish.yaml`.

This cuts ~100–220m to ~20–30m with roughly a one-condition change. Options 2 (RC plus
digest promotion, the KEG/k8s model) and 3 (stage/publish split) are follow-up; RC is
deferred; version decoupling is undecided.

## Notes

Resolved: no re-test on the tag (drop unit and e2e; keep only container-structure plus a
binary-version assert, the ldflags delta); implement via the existing `ci/skip-test`
mechanism in the monolith, not a new workflow; no separate rehearsal (per-push preview
builds already exercise publish); the green-SHA gate is hard-enforced; RC tags deferred.

Open: version decoupling (undecided; rebuild-on-tag is fine for Option 1); the exact
green-SHA enforcement mechanism (impl detail).

Sources. Peer projects: Kubernetes (sig-release handbook, promo-tools), Istio (release-builder,
daily-release), Envoy (RELEASES.md, `_publish_build.yml`), Cilium
(`build-images-releases.yaml`), cert-manager (`cmrel`), Knative (`hack/release.sh`),
Argo CD (`release.yaml`), Linkerd (`release.yml`).
