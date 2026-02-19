# Release process

## Branch/tag format

- Branch: `release-X.Y`
- Tag: `vX.Y.Z` (WITH `v` prefix)
- Default base: `master`
- Latest versions: see `versions.yml`

## Steps

1. Verify CI green on `release-X.Y`
2. Tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"` then `git push --no-verify upstream vX.Y.Z`
3. Monitor build-test-distribute workflow
4. Smoke tests (Universal + K8s)
5. `gh workflow run release.yaml --repo kumahq/kuma --ref master -f version=X.Y.Z -f check=true`
6. Publish release notes
7. Kong Mesh sync (PRs, tag, build, release)
8. Post-release: docs metadata, merge changelog/docs PRs
9. Verify: `curl -L https://kuma.io/installer.sh | VERSION=vX.Y.Z sh -`
10. Announce (GTM Jira, notify EM/PM on #team-mesh)

## Utilities

- Check dep version: `git show upstream/<branch>:go.mod | grep <dep>`
- Image scan: `trivy image <image>:<tag>`
- Go dep scan: `osv-scanner --lockfile=go.mod`
- Update deps: `gh workflow run update-insecure-dependencies.yaml --repo kumahq/kuma` (runs daily 03:00 UTC)
