# Releases

## The `master` branch

The `master` branch is where the active Kuma development happens. This is where all the new features, critical and minor fixes get accepted. Branch from this one when starting to work on a new PR.

## Release branches

The release branches are named in the format `release-A.B`, where `A` is the major release number and `B` is the minor release number. These branches are considered stable and hold the tagged releases of Kuma. The tags are in the format `A.B.C` where `A` is the major, `B` is the minor and `C` is the patch level of the released version.

### Releases

The process to release a Kuma version is as follows:
 * Create the release branch `release-A.B` and tag `A.B.0-rc1`. This also includes the updated GUI for the release.
 * Next 2 days, test the `rc1` version for the backward compatibility, breaking changes and upgrade instruction. Verify the newly added feature are working as expected. Check the documentation reflects the breaking changes and the new functionality.
 * If a critical problem is found, fix it and release `rc2`. Run the test/verification procedures again
 * When there is confidence tag and release `A.B.0` following the checklist below

### Patch releases

After the initial release there may be a need for a patch release `A.B.1`, `A.B.2` and so on. These typically include only fixes to critical issues, but there might be exceptions where features are backported from master. In most cases the patch releases follow a lighter weight release procedure, without `rc` versions and a testing/verification cycle focused on the changed/fixed functionalities.  

# Release Checklist

Below you can find the checklist of all the work that has to be performed when cutting a new release for both major and minor versions of Kuma.

Cutting a new release branch:

- [ ] Update the `CHANGELOG.md`. See [docs](#generating-changelog) 
- [ ] Update the README.md badge if a new release branch
- [ ] Double-check that `UPGRADE.md` is updated with any task that maybe required.
- [ ] Add the new branch to .mergify.yml
- [ ] Manually bump the `version` and `appVersion` of the Helm chart (They should be the same and be `X.Y.0`).
- [ ] Create the new branch `release-X.Y` from master head
- [ ] Double-check that all the new changes have been documented on the [kuma.io website](https://github.com/kumahq/kuma-website) by opening up issues for missing docs.
- [ ] Run the [Helm Release workflow](https://github.com/kumahq/kuma/actions/workflows/helm-release.yaml).
- [ ] Make sure the new binaries are available in [The repo](https://download.konghq.com/mesh-alpine/).

Releasing a new minor release:

- [ ] Check the Changelog.md is up-to-date (only master branch maters).
- [ ] Create a tag and push it.
- [ ] Wait for the release to be published and manually validate it.
- [ ] Make sure the new binaries are available in [The repo](https://download.konghq.com/mesh-alpine/).
- [ ] Create a new [Github release](https://github.com/kumahq/kuma/releases) and create a link to both the changelog and to the assets download.
- [ ] Make sure the `kumactl` formula is updated at [Homebrew](https://github.com/Homebrew/homebrew-core/blob/master/Formula/kumactl.rb)
- [ ] Run the [Helm Release workflow](https://github.com/kumahq/kuma/actions/workflows/helm-release.yaml) check `helm release` when triggering the action.
- [ ] Make sure the Helm Charts are updated at [Helm Hub](https://hub.helm.sh/charts/kuma/kuma)
- [ ] Create a blog post that describes the most important features of the release, linking to the `CHANGELOG.md`, and including the download links.
- [ ] Review and approve the blog post.
- [ ] Post the content on the Kong blog.
- [ ] In preparation for the email newsletter announcement, include any reminder for upcoming community events like meetups or community calls to the content, and send the email to the Kuma newsletter.
- [ ] Announce the release on the community channels like Slack and Twitter.
- [ ] Add issues to upgrade to the latest version of each non dependabot dependency (most notably dependencies in the helm chart like grafana, loki, jaeger...)

Releasing a new patch release:

- [ ] Check the Changelog.md is up-to-date (only master branch maters).
- [ ] Create a tag and push it.
- [ ] Wait for the release to be published and manually validate it.
- [ ] Make sure the new binaries are available in [The repo](https://download.konghq.com/mesh-alpine/).
- [ ] Create a new [Github release](https://github.com/kumahq/kuma/releases) and create a link to both the changelog and to the assets download.
- [ ] Make sure the `kumactl` formula is updated at [Homebrew](https://github.com/Homebrew/homebrew-core/blob/master/Formula/kumactl.rb)
- [ ] Run the [Helm Release workflow](https://github.com/kumahq/kuma/actions/workflows/helm-release.yaml) check `helm release` when triggering the action.
- [ ] Make sure the Helm Charts are updated at [Helm Hub](https://hub.helm.sh/charts/kuma/kuma)
- [ ] Validate the release

### Generating Changelog

Since Kuma 1.6.0 a tool is provided to help curate the changelog.

It will retrieve all the PRs in your change-set and either use the PR title or whatever value of `> Changelog: ` is in the text description of the PR.
It will roll together any changelog item with the same value and any PR with `> Changelog: skip` will not be in the Changelog.
The code for this tool is in `tools/releases/changelog`.

And you can run it in 2 different ways from a tag that already exists in the branch (useful for patch versions):

```shell
go run ./tools/releases/changelog/... github  --branch release-1.5 --from-tag 1.5.0
chore(k8s): replace cni registry (backport #4070) [4076](https://github.com/kumahq/kuma/pull/4076) @mergify
fix(kuma-cp): default policy creation (backport #4073) [4080](https://github.com/kumahq/kuma/pull/4080) @mergify
fix(kuma-cp): guard the nil version in metadata (backport #3969) [3970](https://github.com/kumahq/kuma/pull/3970) @mergify
```

From a specific commit (useful when cutting a new minor):

```shell
go run ./tools/releases/changelog/... github  --branch master --from-commit ee321e2 # this is the first commit not in release-1.5
chore(deps): bump alpine from 3.15.0 to 3.15.2 in /tools/releases/dockerfiles [4060](https://github.com/kumahq/kuma/pull/4060) [4023](https://github.com/kumahq/kuma/pull/4023) @dependabot
chore(deps): bump github.com/envoyproxy/protoc-gen-validate from 0.6.3 to 0.6.7 [3978](https://github.com/kumahq/kuma/pull/3978) [3976](https://github.com/kumahq/kuma/pull/3976) @dependabot
chore(deps): bump github.com/go-logr/logr from 1.2.2 to 1.2.3 [4040](https://github.com/kumahq/kuma/pull/4040) @dependabot
chore(deps): bump github.com/golang-jwt/jwt/v4 from 4.3.0 to 4.4.1 [4061](https://github.com/kumahq/kuma/pull/4061) [4025](https://github.com/kumahq/kuma/pull/4025) @dependabot
chore(deps): bump github.com/k8s/* from 0.23.4 to 0.23.5 [4043](https://github.com/kumahq/kuma/pull/4043) @lahabana
chore(deps): bump github.com/miekg/dns from 1.1.46 to 1.1.47 [3998](https://github.com/kumahq/kuma/pull/3998) @dependabot
chore(deps): bump github.com/onsi/gomega from 1.18.1 to 1.19.0 [4062](https://github.com/kumahq/kuma/pull/4062) @dependabot
chore(deps): bump github.com/spf13/cobra from 1.3.0 to 1.4.0 [3995](https://github.com/kumahq/kuma/pull/3995) @dependabot
chore(deps): bump go.uber.org/multierr from 1.7.0 to 1.8.0 [3974](https://github.com/kumahq/kuma/pull/3974) @dependabot
chore(deps): bump google.golang.org/grpc from 1.44.0 to 1.45.0 [3993](https://github.com/kumahq/kuma/pull/3993) @dependabot
...
```

Once you have this output you can go back and edit PR details to improve the Changelog.
Once it's ready you can format the changelog for `CHANGELOG.md` to make it easily readable for humans. 
