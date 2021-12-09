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

To release a new version of Kuma follow these steps:

- [ ] Update the `CHANGELOG.md`. Use `make changelog` as a helper.
- [ ] Double-check that `UPGRADE.md` is updated with any task that maybe required.
- [ ] In case of Envoy version change update the compatibility matrix.
- [ ] Double-check that all the new changes have been documented on the [kuma.io website](https://github.com/kumahq/kuma-website) by opening up a documentation PR with the new content for the new version of Kuma.
- [ ] Manually bump the `version` and `appVersion` of the Helm chart
- [ ] Create a new Git tag for the release using the [make-release-tag.sh](./tools/release/make-release-tag.sh) script. This script takes 2 arguments, the name of the previous release tag and the name of the new release tag to create and create an annotated tag containing a summary of the commits in the release.
- [ ] Push the Git tag. This will trigger the release job on CI.
- [ ] Make sure the new binaries are available in [Bintray](https://bintray.com/kong/kuma).
- [ ] Download the new Kuma version and double-check that it works with the demo app. Check that is works both in `universal` and `kubernetes` modes.
- [ ] Merge PR to website repository.
- [ ] Create a new [Github release](https://github.com/kumahq/kuma/releases) and create a link to both the changelog and to the assets download.
- [ ] Make sure the `kumactl` formula is updated at [Homebrew](https://github.com/Homebrew/homebrew-core/blob/master/Formula/kumactl.rb)
- [ ] Make sure the Helm Charts are updated at [Helm Hub](https://hub.helm.sh/charts/kuma/kuma)
- [ ] Update the README.md badge if a new release branch has been created and declared stable
- [ ] Create a blog post that describes the most important features of the release, linking to the `CHANGELOG.md`, and including the download links.
- [ ] Review and approve the blog post.
- [ ] Post the content on the Kong blog.
- [ ] In preparation for the email newsletter announcement, include any reminder for upcoming community events like meetups or community calls to the content, and send the email to the Kuma newsletter.
- [ ] Announce the release on the community channels like Slack and Twitter.
- [ ] Add issues to upgrade to the latest version of each non dependabot dependency (most notably dependencies in the helm chart like grafana, loki, jaeger...)
