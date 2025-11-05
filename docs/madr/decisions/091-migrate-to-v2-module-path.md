# Migrate Go Module Path to `/v2` for Semantic Import Versioning

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/2073

## Context and Problem Statement

Kuma and Kong Mesh have been at `v2.x` since 2021 but do not use the `/v2` suffix in their Go module paths. This violates Go's semantic import versioning requirements.

**Current state:**

```go
// Kuma
module github.com/kumahq/kuma
// Results in: github.com/kumahq/kuma v0.0.0-20251103153646-fc66bbfced92

// Kong Mesh
module github.com/Kong/kong-mesh
// Results in: github.com/Kong/kong-mesh v0.0.0-20251103153646-fc66bbfced92
```

**Problems:**

1. Pseudo-versions instead of semantic versions (`v0.0.0-20251103...` instead of `v2.12.4`)
2. Version resolution failures - Go cannot distinguish between incompatible major versions
3. Cannot have `v1` and `v2` coexist in dependency tree
4. Violates Go module best practices (opened in 2021, still unresolved)

## Critical Decision Required

### Should we include this breaking change in patch releases?

#### Arguments FOR Patch Releases

1. **Fixes Existing Breakage**: Current state already violates Go module semantics (issue opened 2021)
2. **Combined with v-prefix**: We're already changing tag format (`2.12.4` → `v2.12.4`)
3. **One-Time Pain**: Delaying spreads migration across more releases
4. **Clear Migration**: Simple find-replace: `github.com/kumahq/kuma/` → `github.com/kumahq/kuma/v2/`

#### Arguments AGAINST Patch Releases

1. **Semantic Versioning**: Patch releases shouldn't contain breaking changes
2. **User Expectations**: Patches should be drop-in replacements
3. **Communication**: Harder to communicate in patch notes

#### Recommendation

**Proceed with patch releases** - we're fixing existing breakage (4 years old), combined with tag format changes already happening, clear migration path.

**Reviewers: Please approve or reject this approach.**

## Design

### Why `/v2` Path is Required

Simply adding `v` prefix to tags (`v2.12.4`) is **insufficient**. Go requires **both** v-prefixed tags AND `/v2` module path for major versions ≥2.

The `/vN` suffix is how Go distinguishes incompatible major versions. Without it, Go treats all as `v0`/`v1`, causing resolution failures and preventing `v1`/`v2` coexistence.

**Reference**: [Go Modules: v2 and Beyond](https://go.dev/blog/v2-go-modules)

### Considered Alternatives

#### Option 1: Do Nothing

- ❌ Leaves issue unresolved, violates Go best practices

#### Option 2: Only Add v-Prefix to Tags

- ❌ Insufficient - doesn't fix version resolution
- ❌ Pseudo-versions persist

#### Option 3: Use `+incompatible` Suffix

- ❌ Auto-upgrades can break builds
- ❌ Rejected by Go 1.13+ if `go.mod` exists
- ❌ Can cause runtime panics

#### Option 4: Major Subdirectory Approach (v2/ subdirectory)

- ❌ Duplicates codebase
- ❌ Harder to maintain across branches
- ❌ Unnecessary (we don't maintain v1)

#### Option 5: Major Branch Approach with `/v2` Path ⭐ **RECOMMENDED**

Update module path to `/v2` on all active branches in-place.

**Changes**:

```go
// go.mod
module github.com/kumahq/kuma/v2

// All imports
import "github.com/kumahq/kuma/v2/pkg/..."
```

**Result**:

```go
// Tagged releases
github.com/kumahq/kuma/v2 v2.12.4

// Untagged commits
github.com/kumahq/kuma/v2 v2.0.0-20251103153646-fc66bbfced92
```

**Pros**:

- ✅ Properly fixes the issue
- ✅ All 5 active branches get the fix
- ✅ No codebase duplication
- ✅ Follows Go best practices
- ✅ Migration scripts tested

**Cons**:

- ⚠️ Breaking change - external users must update imports
- ⚠️ Requires coordination across 5 branches

**External user migration**:

```bash
find . -name "*.go" -type f -exec sed -i '' \
  's|"github.com/kumahq/kuma/|"github.com/kumahq/kuma/v2/|g' {} +
go mod tidy
```

### Affected Components

**Kuma repository** (~10,700+ occurrences):

- `go.mod`, `**/*.go`, `.golangci.yml`, `mk/*.mk`, `tools/**/*.sh`

**Kong Mesh repository**:

- Similar scope to Kuma (internal codebase)

**Branches** (both repositories):

`master`, `release-2.12`, `release-2.11`, `release-2.10`, `release-2.7`

**External impact**:

Users importing Kuma or Kong Mesh as Go modules must update imports. Users only deploying (not importing as libraries) have no impact.

## Security Implications and Review

Low security risk - purely mechanical path changes, no changes to security mechanisms.

## Reliability Implications

**Positive impact**:

Fixes version resolution issues, prevents silent upgrades.

**Risk mitigation**:

- Full CI matrix, E2E tests, smoke tests
- Staged rollout: merge PRs, monitor CI, then tag releases
- **Before tagging**: Easy rollback (revert commits)
- **After tagging**: Cannot revert - must fix forward

## Implications for Kong Mesh

Kong Mesh must migrate to `/v2` module path as part of this effort.

**Dependencies:**

Kong Mesh depends on Kuma, so migration must happen in phases:

1. **Phase 1** (before Kuma migration): Update `mk/upgrade.mk` to use `github.com/kumahq/kuma/v2@`
2. **Phase 2** (after Kuma migration): Migrate Kong Mesh module path to `github.com/Kong/kong-mesh/v2`
3. **Phase 3** (ongoing): Handle automatic Kuma upgrade PRs with import fixes

**Coordination:**

Phase 1 must complete before Kuma migration begins. Phase 2 happens after Kuma releases are tagged.

## Decision

Adopt Major Branch Approach with `/v2` path across all 5 active branches for both Kuma and Kong Mesh.

**What we're doing**:

1. Migrate Kuma and Kong Mesh to `/v2` module paths
2. Apply to all branches in patch releases
3. Provide migration guide and support

**Why**:

- Only viable solution that fixes Go module version resolution
- Aligns with Go best practices
- Fixes 4-year-old issue
- Simple migration for external users (find-replace)
- Kong Mesh must follow to maintain consistency

**Communication**:

- Breaking change prominently featured in release notes
- Migration guide published
- Announced in community channels
- Create and pin GitHub issue with migration instructions for visibility

## Implementation Plan

### Phase 1: Kong Mesh Preparation

Update `mk/upgrade.mk` in Kong Mesh to use `github.com/kumahq/kuma/v2@`. Must complete before Phase 2.

### Phase 2: Kuma Migration

For each branch (`master`, `release-2.12`, `release-2.11`, `release-2.10`, `release-2.7`):

1. Run migration script (or manual steps below)
2. Verify: `make check && make test && make build`
3. Create PR with `ci/run-full-matrix` label
4. Merge (coordinate all branches same day)

<details>
<summary>Migration steps</summary>

**Step 1: Update go.mod module declaration**

```sh
sed -i.bak 's|^module github.com/kumahq/kuma$|module github.com/kumahq/kuma/v2|' go.mod
```

**Step 2: Update all Go imports**

```sh
find . -name '*.go' -type f -not -path './vendor/*' -exec sed -i.bak 's|"github.com/kumahq/kuma/|"github.com/kumahq/kuma/v2/|g' {} +
```

**Step 3: Update .golangci.yml**

```sh
sed -i.bak 's|github.com/kumahq/kuma/|github.com/kumahq/kuma/v2/|g' .golangci.yml
```

**Step 4: Update makefiles**

```sh
find mk -name '*.mk' -type f -exec sed -i.bak 's|github.com/kumahq/kuma|github.com/kumahq/kuma/v2|g' {} +
```

**Step 5: Update shell scripts**

```sh
find tools -name '*.sh' -type f -exec sed -i.bak 's|github.com/kumahq/kuma|github.com/kumahq/kuma/v2|g' {} +
```

**Step 6: Update vulnerable dependencies script (use prefix check instead of exact match)**

```sh
sed -i.bak 's#select(.name != "github.com/kumahq/kuma")#select(.name | startswith("github.com/kumahq/kuma") | not)#g' tools/ci/update-vulnerable-dependencies/update-vulnerable-dependencies.sh
```

**Step 7: Update .proto files go_package option**

```sh
find api pkg/config pkg/plugins test/server/grpc/api -name '*.proto' -type f | xargs sed -i.bak 's|go_package = "github.com/kumahq/kuma/|go_package = "github.com/kumahq/kuma/v2/|g'
```

**Step 8: Clean up backup files**

```sh
find . -name '*.bak' -type f -delete
```

**Step 9: Tidy dependencies**

```sh
go mod tidy
```

**Step 10: Clean tools and regenerate all files**

**IMPORTANT:** Use `make clean/tools generate` to ensure Go tools are rebuilt with new imports. The `clean/tools` target removes cached tool binaries, then `generate` rebuilds them fresh and regenerates all files (including `.pb.go` from `.proto` files with updated `go_package` paths).

```sh
make clean/tools generate
```

**Step 11: Verify module path**

```sh
grep '^module github.com/kumahq/kuma/v2$' go.mod
```

**Step 12: Verify no old imports remain**

```sh
grep -r '"github.com/kumahq/kuma/' --include="*.go" --exclude-dir=vendor | grep -v '"/v2/' || echo "✓ All imports updated"
```

**Step 13: Run checks**

```sh
make check && make test && make build
```

</details>

### Phase 3: Staged Release

1. Merge all PRs (same day)
2. Monitor CI/CD
3. Tag `release-2.12` first
4. Tag remaining branches (`release-2.11`, `release-2.10`, `release-2.7`, `master`)
5. Verify each release, run smoke tests
6. Update docs, publish migration guide
7. Create and pin GitHub issue with migration instructions
8. Announce in release notes and community channels

### Phase 4: Kong Mesh `/v2` Migration

For each Kong Mesh branch (same branches as Kuma):

1. Run migration script (same process as Kuma)
2. Update module path and all imports
3. Verify with tests and checks
4. Create PR, merge, and tag releases

### Phase 5: Handle Kuma Upgrade PRs

Update imports in automatic Kong Mesh upgrade PRs triggered by Kuma releases.

### Rollback Procedures

**Before tagging**: Revert merge commits

```bash
for BRANCH in release-2.12 release-2.11 release-2.10 release-2.7 master; do
  git fetch upstream "$BRANCH"
  MERGE_SHA=$(git log upstream/"$BRANCH" --grep="migrate-to-v2-module-path" --format="%H" -n 1)
  git checkout "$BRANCH" && git pull upstream "$BRANCH"
  git revert -m 1 "$MERGE_SHA" && git push --no-verify upstream "$BRANCH"
done
```

**After tagging**: Cannot revert (breaks early adopters). Fix forward with patch release.
