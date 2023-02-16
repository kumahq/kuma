### Checklist prior to review

<!--
Each of these sections need to be filled by the author when opening the PR.

If something doesn't apply please check the box and add a justification after the `--`
-->

- [ ] Link to docs PR or issue --
- [ ] Link to UI issue --
- [ ] Is the [issue worked on linked][1] (Add `fix #0000` in the PR description) ? --
- [ ] The PR will not break child repos it doesn't hardcode values (.e.g "kumahq" as a image registry) and it will work for both Linux and Windows, system specific functions like `syscall.Mkfifo` have equivalent implementation on the other OS --
- [ ] Tests (Unit test, E2E tests, manual test on universal and k8s) --
- [ ] Do you need to update [`UPGRADE.md`](../blob/master/UPGRADE.md)? --
- [ ] Does it need to be backported according to the [backporting policy](../blob/master/CONTRIBUTING.md#backporting)? --
- [ ] Do you need to explicitly set a [`> Changelog:` entry here](../blob/master/CONTRIBUTING.md#submitting-a-patch) or add a `ci/` label to run fewer/more tests?

[1]: https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue#linking-a-pull-request-to-an-issue-using-a-keyword
