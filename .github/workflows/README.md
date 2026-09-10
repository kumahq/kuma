# Choosing runners

Every job picks its runner through two repository variables, so a fork or a downstream
repository can move CI onto its own runners without editing any workflow. Nothing here is
required: with both variables unset, every job runs on the GitHub-hosted default it has
always used.

## The variables

| name | scope |
| --- | --- |
| `RUNNERS_<BRANCH_SLUG>_<ARCH>` | one branch, one architecture |
| `RUNNERS_<ARCH>` | one architecture, every branch |

`<ARCH>` is `AMD64` or `ARM64`. `<BRANCH_SLUG>` is the branch name uppercased with `-` and
`.` replaced by `_`, so `master` is `MASTER` and `release-2.14` is `RELEASE_2_14`.

The value is a JSON object mapping a size to the labels a runner must carry:

```json
{
  "sm": ["self-hosted", "linux", "x64", "size-sm"],
  "md": ["self-hosted", "linux", "x64", "size-md"],
  "lg": ["self-hosted", "linux", "x64", "size-lg"]
}
```

A job runs only on a runner that has **all** the labels listed for its size. A repository
without a pool for one of the sizes points that key at a pool it does have; nothing
requires three distinct pools.

## Sizes

Jobs declare the size they need, not a pool. The rule:

- **sm** — no checkout, or a checkout plus only `gh`, `jq`, `git` or `curl`. Gates,
  dispatchers, matrix generators, comment posters, and the pollers that spend hours waiting
  on another workflow.
- **md** — needs the mise toolchain and does network, artifact or git work, but compiles no
  Go, builds no container image and starts no cluster. Publishing, SBOM generation,
  backports, doc generation.
- **lg** — compiles Go, runs `go test -race` or golangci-lint, builds images with buildx,
  starts a k3d or kind cluster, or builds a CodeQL database.

When a job's steps change, revisit its size. Overshooting wastes a large slot; undershooting
gets the job OOM-killed on a small one.

## The expression

```yaml
runs-on: ${{ fromJSON(vars.RUNNERS_MASTER_AMD64 || vars.RUNNERS_AMD64 || '{}').lg || 'ubuntu-24.04' }}
```

Reading it: take the per-branch variable, else the global one, else an empty object; look up
the size; if any step yields nothing, fall back to the GitHub-hosted label. Only one
possibly-missing property is ever dereferenced, because `'{}'` guarantees the object exists.

Workflows that a `pull_request` event can reach — directly, or as a reusable workflow called
from one — add a fork guard in front, so code from a fork never runs on a self-hosted
runner:

```yaml
runs-on: ${{ (github.event_name == 'pull_request' && github.event.pull_request.head.repo.full_name != github.repository) && 'ubuntu-24.04' || fromJSON(vars.RUNNERS_MASTER_AMD64 || vars.RUNNERS_AMD64 || '{}').lg || 'ubuntu-24.04' }}
```

The `env` context is not available in `runs-on`, which is why the default label is written
inline at every site rather than defined once.

### The e2e jobs

`_test.yaml` runs the same job across architectures, so it cannot read `vars` per leg.
`build-test-distribute.yaml` splices both variables into its `RUNNERS_BY_ARCH` input, and
each leg indexes the result by `matrix.arch` and then by size. That path additionally sends
`master` to GitHub-hosted runners; the per-job sites above do not.

## Exceptions

- `scorecard.yml` must keep a literal label — `scorecard-action` rejects an
  expression-based `runs-on` during workflow verification.
- `_provenance.yaml` and `lifecycle.yml` have no `runs-on`; they call reusable workflows
  that choose their own runner.

## Cutting a release branch

Variable names cannot contain `-` or `.`, and an inline expression cannot sanitize
`github.ref_name`, so the branch slug is written into the workflows by hand. After cutting
`release-X.Y`, replace `RUNNERS_MASTER_` with `RUNNERS_RELEASE_X_Y_` across
`.github/workflows/` on the new branch, comments included, and set the matching variables.
Nothing verifies this, so a missed rename silently falls back to the global variable.
