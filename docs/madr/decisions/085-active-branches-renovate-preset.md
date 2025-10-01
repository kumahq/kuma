# Change `active-branches.json` Into a Renovate Preset

* Status: Accepted

## Context and Problem Statement

`active-branches.json` is a raw JSON array of branch patterns:

```json
["release-2.7","release-2.9","release-2.10","release-2.11","release-2.12","master"]
```

It is used by scripts and workflows, but it is not a valid Renovate preset.

Goal: change it to an object with `baseBranchPatterns` so Renovate can extend it:

```json
{ "baseBranchPatterns": ["release-2.7","release-2.9","release-2.10","release-2.11","release-2.12","master"] }
```

Then in `renovate.json`:

```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": ["kumahq/kuma:active-branches"]
}
```

This enables OSV alerts on release branches and lets Renovate patch Envoy on those branches.

## Decision Drivers

* One source of truth for active branches
* Reduce custom security scripts for non default branches
* Enable Renovate scope control on release branches
* Keep the change low risk

## Options Considered

1. Keep array and add a separate preset file
2. Switch to object with `baseBranchPatterns` and update all consumers
3. Create a new file for Renovate and keep the old one during migration

## Decision Outcome

Chosen option: **2. Switch to object with `baseBranchPatterns` and update all consumers.**
Single file, no duplication, immediately useful to Renovate.

### Positive Consequences

* One file powers scripts and Renovate
* Consistent branch list across repos
* Easier enablement of OSV alerts and controlled updates

### Negative Consequences

* All consumers expecting an array must be adjusted

## Scope of Change

Update all consumers that rely on `active-branches.json`, including:

* `kumahq/kuma` workflows and scripts
* `kumahq/ci-tools` release-tool
* Relevant private repositories

## Backwards Compatibility and Migration

Support both shapes during rollout with tolerant `jq`:

```bash
# Iterate branches as lines
jq -r 'if type=="array" then .[] else .baseBranchPatterns[] end' active-branches.json

# Get a compact JSON array
jq -c 'if type=="array" then . else .baseBranchPatterns end' active-branches.json
```

After migration, callers can read `.baseBranchPatterns` directly.

## Renovate Specifics

* The file becomes a preset named `active-branches`
* Extending it sets `baseBranchPatterns` in Renovate
* OSV vulnerability alerts are enabled by the `Kong/public-shared-renovate` default preset we use

## Example Renovate Configuration for Release Branches

To limit regular updates on release branches to only OSV fixes and Envoy patch bumps:

```json
{
  "extends": [
    "kumahq/kuma:active-branches"
  ],
  "packageRules": [
    {
      "matchBaseBranches": ["/release-.*/"],
      "matchPackageNames": ["*"],
      "enabled": false
    },
    {
      "matchBaseBranches": ["/release-.*/"],
      "matchDepNames": ["envoy"],
      "matchUpdateTypes": ["patch"],
      "enabled": true
    }
  ]
}
```

OSV vulnerability alerts will still open PRs on all branches listed by `baseBranchPatterns`.

## Testing

* Test on forks that use the GitHub App installation of Renovate
* Confirm that `baseBranchPatterns` from the preset is applied
* Verify that OSV PRs target release branches and that only Envoy patch updates are proposed on those branches

## Rollout and Rollback

* Rollout: Update the file and switch consumers to tolerant `jq` in one PR, then enable the Renovate preset in dependent repos
* Rollback: Revert the file to the array shape; tolerant consumers continue to work

## Security and Reliability

* No secrets involved
* Replace custom security alert logic on non default branches with built-in Renovate behavior

## Notes

* Once all consumers are updated, simplify to reading `.baseBranchPatterns` only
