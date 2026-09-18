# Choosing runners

Almost every job picks its runner through repository variables, so a fork or a downstream repository can move CI onto its own runners without editing any workflow. Three of them apply to one architecture: a pull-request tier, a per-branch one and a global one, each set independently. Nothing here is required: with all three unset, every job keeps the runner it had, which is a GitHub-hosted label for most of them and the shared `ubuntu-latest-kong` pool for the e2e legs.

## The variables

| name | scope | wins over |
| --- | --- | --- |
| `RUNNERS_PR_<ARCH>` | `pull_request` runs only | everything below |
| `RUNNERS_<BRANCH_SLUG>_<ARCH>` | one branch, one architecture | the global variable |
| `RUNNERS_<ARCH>` | one architecture, every branch | the built-in default |

`RUNNERS_PR_<ARCH>` exists so pull requests can run somewhere other than pushes to the same branch, on a smaller or cheaper pool, without giving up the per-branch override for pushes. Leave it unset and pull requests follow the branch.

`<ARCH>` is `AMD64` or `ARM64`. `<BRANCH_SLUG>` is the branch name uppercased with `-` and `.` replaced by `_`, so `master` is `MASTER` and `release-2.14` is `RELEASE_2_14`.

The value is a JSON object mapping a size to the labels a runner must carry:

```json
{
  "sm": ["self-hosted", "linux", "x64", "size-sm"],
  "md": ["self-hosted", "linux", "x64", "size-md"],
  "lg": ["self-hosted", "linux", "x64", "size-lg"]
}
```

A job runs only on a runner that has **all** the labels listed for its size. A repository without a pool for one of the sizes points that key at a pool it does have; nothing requires three distinct pools.

## Sizes

Jobs declare the size they need, not a pool. The rule:

- **sm** for a job with no checkout, or a checkout plus only `gh`, `jq`, `git` or `curl`. Gates, dispatchers, matrix generators, comment posters, and the pollers that spend hours waiting on another workflow.
- **md** for a job that needs the mise toolchain and does network, artifact or git work, but compiles no Go, builds no container image and starts no cluster. Publishing, SBOM generation, backports, doc generation.
- **lg** for a job that compiles Go, runs `go test -race` or golangci-lint, builds images with buildx, starts a k3d or kind cluster, or builds a CodeQL database.

When a job's steps change, revisit its size. Overshooting wastes a large slot. Undershooting gets the job OOM-killed on a small one.

## The expression

```yaml
runs-on: >-
  ${{ fromJSON((github.event_name == 'pull_request' && vars.RUNNERS_PR_AMD64)
  || vars.RUNNERS_MASTER_AMD64
  || vars.RUNNERS_AMD64
  || '{}').lg
  || 'ubuntu-24.04' }}
```

`>-` folds the newlines into single spaces, so the expression GitHub evaluates is the same one-line string. Every continuation line has to sit at the same indentation, or YAML keeps the newline instead of folding it.

Reading it: on a pull request take the PR variable, otherwise the per-branch one, otherwise the global one, otherwise an empty object. Then look up the size. If any step yields nothing, fall back to the GitHub-hosted label. Only one possibly-missing property is ever dereferenced, because `'{}'` guarantees the object exists.

Workflows that a `pull_request` event can reach, directly or as a reusable workflow called from one, add a fork guard in front, so code from a fork never runs on a self-hosted runner:

```yaml
runs-on: >-
  ${{ (github.event_name == 'pull_request'
  && github.event.pull_request.head.repo.full_name != github.repository)
  && 'ubuntu-24.04'
  || fromJSON((github.event_name == 'pull_request' && vars.RUNNERS_PR_AMD64)
  || vars.RUNNERS_MASTER_AMD64
  || vars.RUNNERS_AMD64
  || '{}').lg
  || 'ubuntu-24.04' }}
```

The `env` context is not available in `runs-on`, which is why the default label is written inline at every site rather than defined once.

### The e2e jobs

`_test.yaml` runs the same job across architectures, so it cannot read `vars` per leg. `build-test-distribute.yaml` resolves the tiers for both architectures into one `RUNNERS_BY_ARCH` map, and each leg indexes the result by `matrix.arch` and then by size. That path additionally sends `master` to GitHub-hosted runners; the per-job sites above do not.

## Adding an architecture

Set `RUNNERS_ARM64` (or the `PR` or per-branch form) to the same size-to-labels object, with labels that name an arm64 pool. Nothing else changes: `build-test-distribute.yaml` resolves both architectures into one `RUNNERS_BY_ARCH` map, hands it to every reusable workflow it calls, and the jobs with an architecture matrix index it by `matrix.arch`. An architecture nobody has set resolves to an empty object, so those jobs keep falling back to the runner they use now.

Today that means the arm64 e2e legs in `_test.yaml` and the `linux/arm64` leg of `build-binaries`, which cross-compiles on an amd64 host while `RUNNERS_ARM64` is unset and builds natively once it is. Binaries are unaffected either way, since `CGO_ENABLED=0` makes the two byte-identical. A `darwin` leg cross-compiles wherever it lands, so it always asks for amd64.

`build-images` still builds every architecture on one host through qemu, so an arm64 pool does not speed it up without splitting that job per architecture first.

## Exceptions

- `scorecard.yml` must keep a literal label, because `scorecard-action` rejects an expression-based `runs-on` during workflow verification.
- `pr-comments.yaml` must stay GitHub-hosted. It checks out the head of the PR a maintainer commented on, which is a fork on most pull requests, and runs `make` against it. `runs-on` cannot read the step that resolves `isCrossRepository`, so the runner is chosen before the workflow knows whose code it is about to run.
- `_provenance.yaml` and `lifecycle.yml` have no `runs-on`. They call reusable workflows that choose their own runner.

## Cutting a release branch

Variable names cannot contain `-` or `.`, and an inline expression cannot sanitize `github.ref_name`, so the branch slug is written into the workflows by hand. After cutting `release-X.Y`, replace `RUNNERS_MASTER_` with `RUNNERS_RELEASE_X_Y_` across `.github/workflows/` on the new branch, comments included, and set the matching variables. A missed rename would silently fall back to the global variable, so `validate-workflows-and-scripts.yaml` fails when a slug in the workflows does not match the branch. `PR` is exempt, since it names a tier rather than a branch.

# What decides how much CI a pull request runs

`build-test-distribute` decides in its `meta` job, and every job that reads a label reads that decision. Draft state is the exception: `build_check` and `check` read it straight from the event, because it is correct there and waiting on `meta` would put an API call in front of a forty minute job. Two things feed it.

## Draft state

A draft runs nothing expensive: `build_check`, `check`, `test_unit`, the whole e2e matrix, and `build_publish` with the container-structure test inside it are all skipped. Press **Ready for review** to run them - `ready_for_review` starts a fresh run - and converting back to a draft cancels the run in flight and replaces it with one that skips. `validate-workflows-and-scripts` follows the same rule. `check.yaml` - the "PR health" workflow, not the `check` job above - deliberately does not: it runs commitlint over the pull request title, which becomes the commit message on a squash merge, and it takes seconds on the smallest runner. A wrong title is worth catching on the first push.

Two consequences worth knowing. A pull request a bot opens as a draft gets the same treatment - a backport whose cherry-pick conflicted is opened as a draft, so it runs nothing until whoever resolves the conflict marks it ready. And GitHub disables auto-merge when a pull request becomes a draft, so `auto-merge.yaml` listens for `ready_for_review` to arm it again.

Anything automated that marks a pull request ready has to do it with an app token, never `github.token`. GitHub does not raise a workflow run for an event its own token caused, so `ready_for_review` would not fire: no new run would replace the draft run's skips, and the pull request would sit at ready carrying a row of green that nothing earned. `backport.yaml` opens a conflicted cherry-pick as a draft and already falls back to `github.token` when `secrets.APP_ID` is unset, so the trap is one automation away rather than hypothetical. Nothing marks a pull request ready today.

The gate is always a job-level condition, never a narrowed trigger. A job skipped by a condition reports Success and satisfies a required status check, while a workflow that never fires leaves that check waiting for a report and blocks the pull request for good. Whoever owns branch protection should therefore read a green box on a draft as *not applicable* rather than as passed - GitHub refuses to merge a draft either way, and marking it ready starts the run that produces the real answer. `distributions` asserts that the other way round, and it asks the question the merge box needs answered: do this run's results still describe this pull request? The gates decide from the event payload, which is right - draft state is correct there and waiting on `meta` would put an API call in front of a forty minute job - but `distributions` compares what was skipped against the draft state `meta` read live. A run that skipped its gates and belongs to a pull request that is no longer a draft fails, which catches both a gate that quietly stopped working and the one way a replayed event can lie: re-run an old draft run after marking the pull request ready and GitHub writes that run's skips over the real results, so the commit would otherwise show a row of green that nothing earned. One exception is worth knowing: a *matrix* job skipped this way reports under its unexpanded name - `test / e2e (default, ${{ matrix.k8sVersion }}, ${{ matrix.arch }})` - so the per-leg e2e names do not report whenever the matrix comes out empty, which `ci/skip-test` has always done and a draft now does too. They come back when the matrix has entries.

Two rules follow from that, and a change here has to keep both. `distributions` is the fan-in: it lists every job in its `needs`, because a job it does not need is a job whose failure it cannot see, and nothing else catches it - it runs on `!cancelled()` rather than on its needs having succeeded. And each gate compares a decision as a string against the value that turns work off, so a decision that never arrived falls on the safe side: the tests run and publishing does not.

### What the gate costs

Every ordinary case runs less CI than it did, and a draft push runs about 97% less - four jobs, each under a minute, against twenty jobs and 180 runner-minutes. Two cases cost more, both bounded and both deliberate.

Marking a draft ready runs the suite, and it runs it even when the head commit already had a green run - open ready, convert to draft, mark ready again without pushing, and the suite runs twice. It has to: the draft run wrote its skips over the first run's results, so the answer is no longer on the commit. Short-circuiting that would mean letting a row of skips satisfy a pull request that is no longer a draft, which is the thing the fan-in check exists to refuse.

`ci/force-publish` on a pull request from a fork used to abort `check` at its first step. The refusal is its own job now, so it reports in seconds while `check` runs on for its usual thirteen minutes. That is thirteen minutes of GitHub-hosted runner, not of the Kong pool, which every fork pull request uses anyway.

## Labels

`meta` reads the labels the pull request carries at the moment it runs, through the API, rather than the set the webhook carried - a pull request cannot be created with labels, so they always arrive in a second call and the webhook's copy is routinely empty. A label added before `meta` runs counts; one added after it takes effect on the next run, or on a re-run of this one - the read goes to the API rather than replaying the webhook's copy, so **Re-run all jobs** picks up a label the original run never saw. `meta` reaches that step twenty to forty seconds into a run, which is worlds more than a tool needs - `gh pr create --label`, Renovate, `release.yaml` and `backport.yaml` all attach within a second of opening - and not long enough for someone clicking a label by hand. For them the label lands on the next push or a re-run. The exception is `ci/auto-merge`, which `auto-merge.yaml` reads from the event payload - it triggers on `labeled`, so applying it is itself the event.

| label | what it does |
| --- | --- |
| `ci/skip-test` | skips the unit tests, the e2e matrix and the container-structure test |
| `ci/skip-e2e-test` | skips the e2e matrix only |
| `ci/skip-container-structure-test` | skips the container-structure test inside `build_publish` |
| `ci/run-full-matrix` | runs the full matrix instead of the reduced pull-request one |
| `ci/run-build` | builds the artifacts a pull request does not build by default |
| `ci/force-publish` | builds and publishes them; refused on a pull request from a fork |
| `ci/auto-merge` | approves and enables auto-merge, and re-arms when a draft is marked ready |
| `ci/verify-stability` | reruns CI to find flakes, removed after several consecutive green runs; drafts are ignored |
| `ci/verify-stability-merge-master` | the same, merging master before each rerun |

Every one of them is declared in `meta_repo.yml`.
