---
name: kuma-backport
description: Backport a merged kumahq/kuma PR to release branches end-to-end — trigger the backport workflow if needed, find the generated backport PRs, resolve cherry-pick conflicts (module-path rewrites, older APIs), fix DCO, build+test per branch, force-push, clean up PR descriptions/labels, mark ready, and watch CI. Use whenever the user gives a kuma PR link and mentions backporting, backport PRs, cherry-pick conflicts on release-X.Y branches, or asks to fix CI on chore/backport-* PRs.
argument-hint: "<original-pr-url> [release-2.x,release-2.y | all]"
user-invocable: true
---

# Kuma Backport Fixer

Input: the URL or number of the **original merged PR** on kumahq/kuma (master), optionally a list of target branches. Run from a kuma checkout. Below, `$UPSTREAM` is the git remote pointing at `github.com/kumahq/kuma` — backport head branches live there (not on forks), so push access to it is required.

## 1. Gather facts about the original PR

```bash
gh pr view <PR> --repo kumahq/kuma --json state,mergeCommit,title,files
```

Record the merge commit SHA (`MERGE_SHA`) and the changed file list — they are the source of truth for every conflict resolution later. The PR must be MERGED; the workflow rejects unmerged PRs.

## 2. Find or create the backport PRs

Backport PRs are created by workflow_dispatch (there is **no** label-driven backport in kuma):

- Search first: `gh pr list --repo kumahq/kuma --search "backport of #<PR> in:title" --state all --json number,title,baseRefName,isDraft,labels,headRefName`
- If none exist (or the user asked to trigger it):

  ```bash
  gh workflow run backport.yaml --repo kumahq/kuma --ref master -f PR=<num> [-f branches="release-2.x,release-2.y"]
  ```

  Empty `branches` targets all active branches. Then poll `gh run list --workflow=backport.yaml --repo kumahq/kuma --limit 1` until complete and re-run the PR search. Poll in background; the run takes a few minutes.

Facts about what the action produces (this shapes all later steps):

- Head branches: `chore/backport-<release-branch>-<PR>`, pushed to kumahq/kuma directly.
- Committer is `kumahq[bot]`, author is the original PR author. If the author's email doesn't match the `Signed-off-by` trailer, **DCO fails** — this happens even on conflict-free backports, so check DCO on every PR, not just conflicted ones.
- On conflict the action runs `git add . && git cherry-pick --continue` **blindly**: conflict markers get committed into files, and files new on master land with master's import paths. The PR is opened as draft with a `conflict` label and the `git status` dump in the body.

## 3. Resolve conflicts per branch

For each backport PR whose body has the conflict warning (process branches one at a time; verify each before moving on):

```bash
git fetch $UPSTREAM chore/backport-<rel>-<PR> <rel>
git checkout -B backport-<ver> $UPSTREAM/chore/backport-<rel>-<PR>
```

### Know the branch's module path first

Master is `github.com/kumahq/kuma/v3`. Release branches differ and are **not uniform** — always read `git show $UPSTREAM/<rel>:go.mod | head -1` instead of assuming (as of mid-2026: 2.7–2.14 are `/v2` except 2.9 which is unsuffixed; don't trust this list, check).

### Resolution strategy, in order of preference

1. **Whole-file from original commit + path rewrite** — works when the branch's file matches master apart from import paths:

   ```bash
   git show $MERGE_SHA:<file> | sed 's|github.com/kumahq/kuma/v3|<branch-module-path>|g' > <file>
   ```

2. **Patch onto the branch's own file** — required when the branch has older third-party APIs (e.g., controller-runtime signatures changed; a whole-file copy then fails to build). Regenerate the file from the release branch and apply only the fix:

   ```bash
   git show $UPSTREAM/<rel>:<file> > <file>
   SOURCE_MODULE_PATH=$(git show $UPSTREAM/<other-rel>:go.mod | awk 'NR==1 { print $2 }')
   git diff $UPSTREAM/<other-rel> backport-<other-ver> -- <file> \
     | perl -pe 's#\Q'"$SOURCE_MODULE_PATH"'\E#<branch-module-path>#g' \
     | git apply -
   ```

   Reuse a patch produced from an already-fixed sibling branch — the fix hunks are identical across branches even when the surrounding file isn't.

3. Only if both fail: hand-merge, guided by `git show $MERGE_SHA -- <file>` (the intended change) against the branch file.

### Style and hygiene

- Older branches (≤2.13 as of now) use `interface{}` where master uses `any`. Match the branch: `perl -pe 's/\bany\b/interface{}/g'` (BSD sed has no `\b`). Keep the backport diff minimal — don't rewrite branch signatures to master style.
- New test/suite files from the cherry-pick also carry master's module path — rewrite those too.
- Sweep for leftovers before committing:

  ```bash
  grep -Ern '<<<<<<<|kumahq/kuma/v3' <changed-dirs>
  ```

## 4. Validate locally

```bash
eval "$(mise env)"
gofmt -l <changed-dirs>
go build ./<changed-pkg>/... && go test ./<changed-test-pkgs>/...
```

`mise env` gives the branch's toolchain; `gofmt -l` must print nothing. Then verify the branch diff has the same shape as the original PR — same files, similar stat:

```bash
git diff $UPSTREAM/<rel> HEAD --stat
```

Extra files or missing files mean the bot's blind auto-resolution left damage outside the reported conflicts — fix before pushing. Don't run `make check` on a dirty tree (it fails by design); the targeted build/test above is enough locally, CI runs the full check.

## 5. Fix DCO and push

Amend (don't add a fixup commit — keeps the backport a single clean cherry-pick):

```bash
git add -A
git commit --amend --no-edit --reset-author -s
git push --no-verify --force-with-lease $UPSTREAM HEAD:chore/backport-<rel>-<PR>
```

`--reset-author` makes the author match the local identity whose `Signed-off-by` the `-s` adds — that equality is exactly what the DCO probot checks. Add `-S` if you GPG-sign commits. Do this even on conflict-free backport PRs when DCO is red.

A backport commit always needs rewriting (drop conflict markers, fix the author/sign-off), so the push is inherently non-fast-forward — a plain push is rejected and a fixup commit on top can't fix DCO on the underlying bot commit. Force is unavoidable. If a git hook or policy blocks force-pushes in your environment, resolve every branch locally first, then hand the exact `git push --force-with-lease $UPSTREAM <local>:chore/backport-<rel>-<PR>` commands to a human to run rather than trying to work around the block.

## 6. PR hygiene

For each PR:

- **Body**: keep the auto-generated header (cherry-pick line, action link, commit SHA); replace the `:warning:` block + git-status dump with 1–2 sentences stating what conflicted and how it was resolved (e.g., module import path v3→v2; `interface{}` kept per branch style). `gh pr edit <n> --repo kumahq/kuma --body "..."`.
- **Labels**: `--remove-label conflict`. Keep `backport` + `release-X.Y` — that's what merged backports carry.
- **Ready**: `gh pr ready <n> --repo kumahq/kuma`.

## 7. Watch CI

Poll all PRs in one background loop until nothing is `pending`; ignore `skipping` rows:

```bash
gh pr checks <n> --repo kumahq/kuma
```

If `check` fails, read the job log (`gh run view <run-id> --log-failed`) — typical causes are lint (import aliases, gci ordering) or a package that only builds on newer branches. Fix, re-amend, re-push. Report the final per-PR status table to the user.

## Gotchas

- Backport heads live on kumahq/kuma itself — push there, not to a fork, or the PR won't update.
- `distributions`/`check` failing together usually just means the build was broken by committed conflict markers — fix the code, don't debug the jobs.
- The changelog comes from the PR title (`... (backport of #N)`), so body edits are safe.
- Force-push always with `--force-with-lease` right after a fresh fetch — the action or a teammate may have updated the branch.
