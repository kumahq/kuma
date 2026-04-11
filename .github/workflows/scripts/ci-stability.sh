#!/usr/bin/env bash
#
# Process open PRs with the "ci/verify-stability" label:
#   1. Observe check-run status for the current HEAD of each PR.
#   2. Append the observation to a sticky rollup comment (per PR) that
#      tracks run history and surfaces likely-flaky jobs.
#   3. Push an empty commit to re-trigger CI, unless a stop condition
#      is met (fork, pending checks, N consecutive greens).
#
# Inputs:
#   $1                                  Path to JSON file from `gh pr list`
#   env GH_USER, GH_EMAIL               Git identity for trigger commits
#   env GITHUB_TOKEN                    Token with contents+PR write scope
#   env STABILITY_CONSECUTIVE_GREEN_THRESHOLD  (default: 5)
#   env STABILITY_MAX_HISTORY                  (default: 20)

set -uo pipefail

readonly OPEN_PRS_FILE="${1:-open_prs.json}"
readonly OWNER="${GITHUB_REPOSITORY%%/*}"
readonly REPO="${GITHUB_REPOSITORY##*/}"
readonly RUN_URL="${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}"
readonly STICKY_MARKER="<!-- ci-stability-bot -->"
readonly CONFLICT_MARKER="<!-- ci-stability-merge-conflict -->"
readonly STATE_PREFIX="<!-- ci-stability-state: "
readonly STATE_SUFFIX=" -->"
readonly GREEN_THRESHOLD="${STABILITY_CONSECUTIVE_GREEN_THRESHOLD:-5}"
readonly MAX_HISTORY="${STABILITY_MAX_HISTORY:-20}"

log()  { printf '::notice::%s\n' "$*"; }
warn() { printf '::warning::%s\n' "$*"; }
err()  { printf '::error::%s\n' "$*"; }

summary() {
  [[ -n "${GITHUB_STEP_SUMMARY:-}" ]] && printf '%s\n' "$*" >> "$GITHUB_STEP_SUMMARY"
}

# ---- state encoding ---------------------------------------------------------
# State is a JSON blob persisted inside the sticky comment as base64 in an
# HTML comment marker. Base64 sidesteps all markdown/escape concerns.

encode_state() { printf '%s' "$1" | base64 -w0; }
decode_state() { printf '%s' "$1" | base64 -d 2>/dev/null; }

# ---- sticky comment helpers -------------------------------------------------

fetch_sticky_comment() {
  # Prints the sticky comment JSON ({id, body}) or empty.
  local pr=$1
  gh api "repos/${OWNER}/${REPO}/issues/${pr}/comments" --paginate 2>/dev/null |
    jq -c --arg marker "$STICKY_MARKER" '
      [.[] | select(.body | contains($marker))] | first // empty
    '
}

extract_state() {
  local body=$1 encoded decoded
  encoded=$(printf '%s' "$body" |
    sed -n "s|.*${STATE_PREFIX}\([^ ]*\)${STATE_SUFFIX}.*|\1|p" | head -n1)
  if [[ -z "$encoded" ]]; then
    printf '{"runs":[]}'
    return
  fi
  decoded=$(decode_state "$encoded")
  if [[ -z "$decoded" ]] || ! printf '%s' "$decoded" | jq empty 2>/dev/null; then
    printf '{"runs":[]}'
    return
  fi
  printf '%s' "$decoded"
}

render_comment() {
  local state=$1
  local run_count flaky history encoded
  run_count=$(jq '.runs | length' <<<"$state")
  flaky=$(jq -r --argjson total "$run_count" '
    [.runs[] | select(.result == "fail") | .failed_jobs[]?]
    | group_by(.)
    | map({job: .[0], count: length})
    | sort_by(-.count)
    | .[] | select(.count >= 2)
    | "- `\(.job)` — failed \(.count) of \($total) run(s)"
  ' <<<"$state")
  history=$(jq -r --argjson max "$MAX_HISTORY" '
    .runs | .[-$max:] | reverse | .[] |
    "| \(.number) | \(.observed_at) | [`\(.sha[0:7])`](\(.commit_url)) | " +
    (if   .result == "pass" then "✅ pass"
     elif .result == "fail" then "❌ fail"
     else .result end) +
    " | \((.failed_jobs // []) | join(\"<br>\")) | [logs](\(.observed_run_url)) |"
  ' <<<"$state")
  encoded=$(encode_state "$state")

  {
    printf '%s\n' "$STICKY_MARKER"
    printf '## 🔁 CI Stability Monitor\n\n'
    printf 'This PR has `ci/verify-stability` enabled — CI is re-triggered outside business hours to surface flaky tests.\n\n'
    printf -- '- **Stop:** remove the `ci/verify-stability` label\n'
    printf -- '- **Auto-stop:** after %s consecutive green runs\n' "$GREEN_THRESHOLD"
    printf -- '- **Observations recorded:** %s\n\n' "$run_count"
    if [[ -n "$flaky" ]]; then
      printf '### ⚠️ Likely flaky jobs\n%s\n\n' "$flaky"
    fi
    printf '### History\n'
    printf '| # | Observed (UTC) | Commit | Result | Failed jobs | Stability run |\n'
    printf '|---|----------------|--------|--------|-------------|---------------|\n'
    [[ -n "$history" ]] && printf '%s\n' "$history"
    printf '\n%s%s%s\n' "$STATE_PREFIX" "$encoded" "$STATE_SUFFIX"
  }
}

upsert_sticky_comment() {
  local pr=$1 body=$2 existing comment_id
  existing=$(fetch_sticky_comment "$pr")
  if [[ -n "$existing" ]]; then
    comment_id=$(jq -r '.id' <<<"$existing")
    gh api "repos/${OWNER}/${REPO}/issues/comments/${comment_id}" \
      --method PATCH -f body="$body" >/dev/null
  else
    gh pr comment "$pr" --body "$body" >/dev/null
  fi
}

# ---- check-run inspection ---------------------------------------------------

fetch_checks() {
  # Emits "bucket name" lines for the PR's current HEAD checks.
  local pr=$1
  gh pr checks "$pr" --json name,bucket 2>/dev/null |
    jq -r '.[] | "\(.bucket) \(.name)"'
}

# ---- per-PR pipeline --------------------------------------------------------

process_pr() {
  local pr=$1 do_merge=$2

  log "PR #${pr}: processing"

  local pr_json
  if ! pr_json=$(gh pr view "$pr" \
      --json headRefName,headRepositoryOwner,state,headRefOid 2>/dev/null); then
    warn "PR #${pr}: could not fetch PR details"
    summary "- ⚠️ PR #${pr}: details fetch failed"
    return
  fi

  if [[ "$(jq -r '.state' <<<"$pr_json")" != "OPEN" ]]; then
    log "PR #${pr}: not open, skipping"
    return
  fi

  local head_owner
  head_owner=$(jq -r '.headRepositoryOwner.login' <<<"$pr_json")
  if [[ "$head_owner" != "$OWNER" ]]; then
    warn "PR #${pr}: head repo is fork \`${head_owner}\`, cannot push"
    summary "- ⏭️ PR #${pr}: skipped (fork \`${head_owner}\`)"
    return
  fi

  local branch head_sha
  branch=$(jq   -r '.headRefName' <<<"$pr_json")
  head_sha=$(jq -r '.headRefOid'  <<<"$pr_json")

  # --- observe current check status ---
  local checks has_pending=0 has_failed=0
  local -a failed_jobs=()
  checks=$(fetch_checks "$pr")
  if [[ -n "$checks" ]]; then
    while read -r bucket name; do
      [[ -z "$bucket" ]] && continue
      case "$bucket" in
        pending)       has_pending=1 ;;
        fail)          has_failed=1; failed_jobs+=("$name") ;;
        cancel)        has_failed=1; failed_jobs+=("$name (cancelled)") ;;
        pass|skipping) ;;
      esac
    done <<<"$checks"
  fi

  if (( has_pending )); then
    log "PR #${pr}: checks pending on ${head_sha:0:7}, not triggering"
    summary "- ⏳ PR #${pr}: pending on \`${head_sha:0:7}\`"
    return
  fi

  # --- load prior state ---
  local sticky_json state
  sticky_json=$(fetch_sticky_comment "$pr")
  if [[ -n "$sticky_json" ]]; then
    state=$(extract_state "$(jq -r '.body' <<<"$sticky_json")")
  else
    state='{"runs":[]}'
  fi

  # --- record observation if definitive ---
  local result="none"
  if [[ -n "$checks" ]]; then
    if (( has_failed )); then result="fail"; else result="pass"; fi
  fi
  if [[ "$result" != "none" ]]; then
    local now_utc failed_json run_number observation
    now_utc=$(date -u +"%Y-%m-%d %H:%M")
    if (( ${#failed_jobs[@]} > 0 )); then
      failed_json=$(printf '%s\n' "${failed_jobs[@]}" | jq -R . | jq -s .)
    else
      failed_json='[]'
    fi
    run_number=$(( $(jq '.runs | length' <<<"$state") + 1 ))
    observation=$(jq -n \
      --argjson n "$run_number" \
      --arg observed_at "$now_utc" \
      --arg sha "$head_sha" \
      --arg commit_url "${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/commit/${head_sha}" \
      --arg run_url "$RUN_URL" \
      --arg result "$result" \
      --argjson failed "$failed_json" '
      {
        number: $n, observed_at: $observed_at, sha: $sha,
        commit_url: $commit_url, observed_run_url: $run_url,
        result: $result, failed_jobs: $failed
      }')
    state=$(jq --argjson obs "$observation" --argjson max "$MAX_HISTORY" \
      '.runs += [$obs] | .runs = (.runs | .[-$max:])' <<<"$state")
  fi

  # --- auto-stop after N consecutive greens ---
  local consec_green
  consec_green=$(jq -r '
    ([.runs[].result] | reverse) as $r |
    ($r | map(. == "pass") | index(false)) // ($r | length)
  ' <<<"$state")
  if (( consec_green >= GREEN_THRESHOLD )); then
    log "PR #${pr}: ${consec_green} consecutive green runs, removing label"
    gh pr edit "$pr" --remove-label "ci/verify-stability"              >/dev/null 2>&1 || true
    gh pr edit "$pr" --remove-label "ci/verify-stability-merge-master" >/dev/null 2>&1 || true
    local body
    body=$(render_comment "$state")
    body="${body}

---
✅ **Auto-stopped** after ${consec_green} consecutive green stability runs. Removed \`ci/verify-stability\` label."
    upsert_sticky_comment "$pr" "$body"
    summary "- ✅ PR #${pr}: auto-stopped (${consec_green} consecutive greens)"
    return
  fi

  # --- fetch & checkout PR branch (local temp ref) ---
  local local_ref="refs/stability/pr-${pr}"
  if ! git fetch --quiet --force origin "pull/${pr}/head:${local_ref}" 2>/dev/null; then
    warn "PR #${pr}: failed to fetch pull ref"
    summary "- ⚠️ PR #${pr}: fetch failed"
    return
  fi
  if ! git -c advice.detachedHead=false checkout --quiet "${local_ref}"; then
    warn "PR #${pr}: checkout failed"
    summary "- ⚠️ PR #${pr}: checkout failed"
    return
  fi
  git config user.name  "${GH_USER}"
  git config user.email "${GH_EMAIL}"

  # --- optional master merge ---
  local merge_note=""
  if (( do_merge )); then
    git fetch --quiet origin master
    if git merge origin/master --no-ff --no-commit --quiet 2>/dev/null; then
      git commit --quiet --allow-empty -m "Merge master into PR #${pr}"
      merge_note=" (merged master)"
    else
      warn "PR #${pr}: merge conflict with master"
      git merge --abort 2>/dev/null || true
      gh pr edit "$pr" --remove-label "ci/verify-stability-merge-master" \
        >/dev/null 2>&1 || true
      gh pr comment "$pr" --body "${CONFLICT_MARKER}
⚠️ **CI stability monitor**: could not merge \`master\` into this PR — **conflicts detected**.

Removing the \`ci/verify-stability-merge-master\` label. Rebase or merge \`master\` manually, then re-add the label to resume auto-merges. Regular \`ci/verify-stability\` triggers will continue." >/dev/null 2>&1 || true
      merge_note=" (master merge conflict)"
      summary "- ⚠️ PR #${pr}: merge conflict with master"
    fi
  fi

  # --- push empty trigger commit ---
  local new_run_number trigger_msg
  new_run_number=$(( $(jq '.runs | length' <<<"$state") + 1 ))
  trigger_msg="ci(stability): trigger run #${new_run_number} for PR #${pr}

Workflow run: ${RUN_URL}"
  git commit --quiet --allow-empty -m "$trigger_msg"
  if ! git push --quiet origin "HEAD:refs/heads/${branch}"; then
    err "PR #${pr}: push failed"
    summary "- ❌ PR #${pr}: push failed"
    git update-ref -d "$local_ref" 2>/dev/null || true
    return
  fi

  upsert_sticky_comment "$pr" "$(render_comment "$state")"

  log "PR #${pr}: observed=${result}, triggered run #${new_run_number}${merge_note}"
  summary "- 🔁 PR #${pr}: observed \`${result}\` on \`${head_sha:0:7}\`, triggered run #${new_run_number}${merge_note}"

  git update-ref -d "$local_ref" 2>/dev/null || true
}

# ---- main -------------------------------------------------------------------

main() {
  if [[ ! -f "$OPEN_PRS_FILE" ]]; then
    err "PR list file not found: $OPEN_PRS_FILE"
    exit 1
  fi

  summary "# 🔁 CI Stability monitor"
  summary ""
  summary "Workflow run: ${RUN_URL}"
  summary ""
  summary "## Processed PRs"

  local prs_stability prs_merge
  prs_stability=$(jq -r '.[] | select(.labels[]?.name == "ci/verify-stability") | .number' "$OPEN_PRS_FILE")
  prs_merge=$(jq     -r '.[] | select(.labels[]?.name == "ci/verify-stability-merge-master") | .number' "$OPEN_PRS_FILE")

  if [[ -z "$prs_stability" ]]; then
    log "No PRs with ci/verify-stability label"
    summary "_No PRs with \`ci/verify-stability\` label._"
    return 0
  fi

  declare -A merge_set=()
  while read -r p; do
    [[ -n "$p" ]] && merge_set["$p"]=1
  done <<<"$prs_merge"

  while read -r pr; do
    [[ -z "$pr" ]] && continue
    local do_merge=0
    [[ -n "${merge_set[$pr]:-}" ]] && do_merge=1
    process_pr "$pr" "$do_merge" || warn "PR #${pr}: processing error (continuing)"
  done <<<"$prs_stability"
}

main "$@"
