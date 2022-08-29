### Checklist prior to review

<!--
Each of these sections need to be filled by the author when opening the PR.

If something doesn't apply please check the box and add a justification after the `--`
-->

- [ ] Link to docs PR or issue --
- [ ] Link to UI issue or PR --
- [ ] Is the [issue worked on linked][1]? --
- [ ] The PR does not hardcode values that might break projects that depend on kuma (e.g. "kumahq" as a image registry) --
- [ ] The PR will work for both Linux and Windows, system specific functions like `syscall.Mkfifo` have equivalent implementation on the other OS --
- [ ] Unit Tests --
- [ ] E2E Tests --
- [ ] Manual Universal Tests --
- [ ] Manual Kubernetes Tests --
- [ ] Do you need to update [`UPGRADE.md`](../blob/master/UPGRADE.md)? --
- [ ] Does it need to be backported according to the [backporting policy](../blob/master/CONTRIBUTING.md#backporting)? --
- [ ] Do you need to explicitly set a [`> Changelog:` entry here](../blob/master/CONTRIBUTING.md#submitting-a-patch)?

[1]: https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue#linking-a-pull-request-to-an-issue-using-a-keyword
