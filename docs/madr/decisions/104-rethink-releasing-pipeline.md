# Rethink the releasing pipeline to release faster

- Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/16973

## Context and Problem Statement

Pushing a `vX.Y.Z` tag runs the full `build-test-distribute.yaml`, including the
entire e2e suite, before any artifact ships. Recent tag runs: v2.13.8 218m,
v2.14.0 124m, v2.12.12 119m, v2.11.16 102m (~1.7–3.6h), on top of the manual steps in
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

- Hard-enforce a green tagged SHA: a guard job on the tag fails the release unless the
  tagged commit already has a trustworthy full-matrix `build-test-distribute` push run on
  its branch that both tested green and published its preview (mechanism in "Enforcing the
  green tagged SHA"). This reads an existing run, not new test machinery.

After this, `build_publish` ~15m plus provenance/distributions lands at roughly 20–30m
versus the current 100–220m.

- (+) Answers the issue; no new infra, staging, or version change; removes the flakiest stage.
- (−) Loses the exact-tag re-test, mitigated by the hard green-SHA gate. (Kuma has no RC
  process today, so RC soak is not a current fallback; it only becomes available under
  Option 2.) Still rebuilds on the tag (~15m, fine).

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

- A, stay in the monolith. Tags carry no PR labels, so `ci/skip-test` can't fire on a tag;
  instead add `|| github.ref_type == 'tag'` to the two guards that label drives —
  `test_unit`'s `if:` (`_test.yaml:21`) and the `SKIP_E2E_TESTS` env (`_test.yaml:82`,
  which empties the e2e matrix). That skips unit and e2e on the tag; the existing `needs:`
  chain then lets `build_publish` proceed. (+) two-line change, no new workflow; the
  publish path stays exercised on every push, so it can't rot. (−) the monolith keeps its
  tag-vs-PR conditionals.

- B, a dedicated `release-publish.yaml` on tags, reusing `_build_publish.yaml` and
  `_provenance.yaml`, never `_test.yaml`. (+) a linear, readable release flow. (−) it
  runs only on tags, so the publish path is no longer exercised between releases and can
  break unnoticed (the predictability the monolith gives for free); plus a new file and
  trigger migration.

Picked A. The monolith already publishes on every push, the strongest guard that the
release machinery works, and extending the existing skip guards keeps the change tiny.
(An earlier draft picked B with a scheduled rehearsal to re-create that exercise; review
pointed out the monolith already provides it, making both the dedicated workflow and the
rehearsal unnecessary.)

### Enforcing the green tagged SHA

The `github.ref_type == 'tag'` skip is unconditional on its own; nothing about it ties the
skip to the tagged commit being green, so by itself it would happily ship an untested tag.
A guard job runs first on the tag and gates the rest of the pipeline on a verified-green
SHA. There are two ways to react to a missing green run:

- Fail-hard (chosen): if the tagged SHA has no trustworthy green *and published* run, the
  tag fails immediately, before anything builds. The releaser re-runs the branch's push CI
  (a re-run keeps `event == push`, so it re-tests *and* re-publishes the preview) and
  re-tags.

- Re-run-on-tag (rejected): fall back to running unit and e2e on the tag when no green run
  is found. Rejected because it re-introduces the flaky ~90–150m e2e this MADR removes,
  precisely at release time; it keeps a second, rarely-exercised code path that rots; and a
  tag-run flake can fail the release, the secondary problem the MADR set out to fix. It is
  also the lone-outlier (Linkerd) model we are moving away from.

We pick fail-hard: the premise is that the branch SHA is already fully tested *and has
already published a preview*, so the gate must *enforce* both invariants, not silently
re-test when they are absent. The failure surfaces in seconds and is actionable, and
because a branch push both runs the full matrix and publishes, recovery is "re-run the
branch's push CI on the SHA, then tag."

Picking the run is the fiddly part (a SHA can have several runs, and only some both tested
*and* published), so the lookup is explicit:

- List `build-test-distribute.yaml` runs for the tagged SHA (`head_sha`), completed, with
  `event == push`. A push forces `FULL_MATRIX=true` (so e2e always ran) and is the only
  event that sets `ALLOW_PUSH=true` (so it also released a `-preview.<hash>` artifact). This
  excludes `pull_request` runs (which can set `SKIP_E2E_TESTS` or reduce the matrix) *and*
  `workflow_dispatch` runs (which run the full matrix but leave `ALLOW_PUSH=false`, so they
  build without publishing anything — passing the old test-only gate while shipping no
  artifact).

- Keep `conclusion == success` and take the most recent.

- Guard against an emptied matrix *and a build-only run*: fetch that run's jobs and require
  the tested jobs (`test_unit`, `test_e2e*`) plus the publish jobs — `digest-images` (the
  `ALLOW_PUSH` sentinel, absent on any non-publishing run), `build-binaries`, `build-images`,
  and `publish-helm` — to be present and successful. A skipped-e2e run reports overall
  success but has no e2e jobs, and a non-publishing run has no `digest-images` job, so
  presence is the real check.

- No match ⇒ fail the tag, with a message pointing at the branch push CI.

### The publish path is already exercised continuously (no separate rehearsal)

Moving the gate off the tag raises a question: does the release machinery still get
exercised? It does, for free. `build-test-distribute.yaml` runs on every master and
`release-X.Y` push with `ALLOW_PUSH=true`, so every push already builds, pushes, signs,
and publishes a `-preview.<hash>` artifact through the same `_build_publish.yaml` the release
uses (`version.sh` auto-routes preview versions to `notary-internal` and the preview
Cloudsmith repo). That continuous run is the rehearsal; a separate scheduled one would be
redundant machinery.

The green-SHA gate turns this from an assumption into an enforced precondition: it accepts
only a `push` run for the tagged SHA — i.e. one that actually published that SHA's preview —
so the tag never ships a tree whose build-and-publish path has not already run green. A
SHA greened only by `workflow_dispatch` (full matrix, but `ALLOW_PUSH=false`) is rejected.

The only steps not exercised between releases are the tag-gated ones: SLSA provenance
upload, the helm chart release to `kumahq/charts`, and the GitHub Release. They are first
exercised at release time. If that risk ever bites, revisit it with Option 2 rather than
standing up separate rehearsal machinery.

### Evolution

This started from the issue's "skip the full suite on the tag." Peer research confirmed
it is the industry norm; only Linkerd re-tests. An early draft added a dedicated
`release-publish.yaml` plus a scheduled rehearsal; maintainer review showed the monolith
already publishes a preview on every push, so extending the existing skip guards with a
`ref_type=='tag'` clause is enough and the extra machinery was dropped. The bigger wins,
digest promotion and the stage/publish split, depend on a version rework and an RC process Kuma
lacks. So Option 1 ships now, with Options 2 and 3 as follow-up.

## Security implications and review

- e2e is not a security control; SBOM, provenance, signing, and scan live in
  `build_publish` and `provenance`, which Option 1 leaves unchanged.
- New risk: tagging a never-green or never-published commit, mitigated by the hard green-SHA
  gate, which requires a green push run that already tested and signed/published a preview
  for the SHA.
- Container-structure tests stay, so artifact regressions are still caught on the tag.
- Option 2 *strengthens* security, since the scanned and signed bits are the shipped bits.

## Reliability implications

- Faster releases mean faster security-patch turnaround.
- Dropping the redundant e2e removes the flakiest release gate.
- Residual risk: the branch run goes stale against the tag. The hard (fail-hard) green-SHA
  gate mitigates this today; RC soak would add to it once Option 2 lands.
- Requiring a *published* push run for the SHA also rejects a tree that tested green but
  never published (e.g. greened only via `workflow_dispatch`), so the tag's build-and-publish
  never runs cold at release time.

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

- Skip e2e and unit on the tag by adding a `github.ref_type == 'tag'` clause to the two
  guards the `ci/skip-test` label drives — `test_unit`'s `if:` (`_test.yaml:21`) and the
  `SKIP_E2E_TESTS` env (`_test.yaml:82`) that empties the e2e matrix; the existing
  `needs:` chain then lets `build_publish` publish immediately.

- Verify only the ldflags delta: container-structure on the freshly-built image plus a
  binary-version assert. Keep `check`'s metadata computation, skip its redundant
  validation.

- Hard-enforce a green, already-published tagged SHA via a guard job that fails the tag
  unless the commit has a green full-matrix `push` run that also released its preview
  (fail-hard; see "Enforcing the green tagged SHA").

- No separate rehearsal: every push already builds and publishes a preview through the
  same `_build_publish.yaml`.

This cuts ~100–220m to ~20–30m with a two-line change. Options 2 (RC plus
digest promotion, the k8s/Istio model) and 3 (stage/publish split) are follow-up; RC is
deferred. Version decoupling is out of scope here: Option 1 rebuilds on the tag and stamps
the correct version, so it needs no decoupling; it is a prerequisite only for Option 2 and
is decided there if that is ever taken up.

## Notes

Resolved: no re-test on the tag (drop unit and e2e; keep only container-structure plus a
binary-version assert, the ldflags delta); implement by extending the two skip guards
(`_test.yaml:21` and `:82`) in the monolith, not a new workflow (the `ci/skip-test` label
never fires on a tag, so this is a `ref_type=='tag'` clause, not the label); no separate
rehearsal (per-push preview builds already exercise publish); the green-SHA gate is
hard-enforced (fail-hard, not re-run-on-tag) and requires a green *push* run that also
published a preview for the SHA, not merely a green test run; RC tags deferred; version
decoupling is out
of scope for Option 1 (rebuild-on-tag stamps the version) and belongs to Option 2.

Open: none for Option 1. Option 2 (digest promotion) still needs version decoupling and an
RC cadence before it can be adopted.

Sources. Peer projects: Kubernetes (sig-release handbook, promo-tools), Istio (release-builder,
daily-release), Envoy (RELEASES.md, `_publish_build.yml`), Cilium
(`build-images-releases.yaml`), cert-manager (`cmrel`), Knative (`hack/release.sh`),
Argo CD (`release.yaml`), Linkerd (`release.yml`).
