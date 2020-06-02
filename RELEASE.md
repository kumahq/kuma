# Release Checklist

Below you can find the checklist of all the work that has to be performed when cutting a new release for both major and minor versions of Kuma.

To release a new version of Kuma follow these steps:

- [ ] Update the `CHANGELOG.md`. Use `make changelog` as a helper.
- [ ] Double-check that all the new changes have been documented on the [kuma.io website](https://github.com/Kong/kuma-website) by opening up a documentation PR with the new content for the new version of Kuma.
- [ ] Create a new Git tag for the release.
- [ ] Push the Git tag. This will trigger the release job on CI.
- [ ] Make sure that new binaries are available in [Bintray](https://bintray.com/kong/kuma).
- [ ] Download the new Kuma version and double-check that it works with the demo app. Check that is works both in `universal` and `kubernetes` modes.
- [ ] Merge PR to website repository.
- [ ] Create a new [Github release](https://github.com/Kong/kuma/releases) and create a link to both the changelog and to the assets download.
- [ ] Create a blog post that describes the most important features of the release, linking to the `CHANGELOG.md`, and including the download links.
- [ ] Review and approve the blog post.
- [ ] Post the content on the Kong blog.
- [ ] In preparation for the email newsletter announcment, include any reminder for upcoming community events like meetups or community calls to the content, and send the email to the Kuma newsletter.
- [ ] Announce the release on the community channels like Slack and Twitter.
