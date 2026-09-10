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
