### Checklist prior to review

<!--
Each of these sections need to be filled by the author when opening the PR.

If something doesn't apply please check the box and add a justification after the `--`
-->

- [ ] [Link to relevant issue][1] as well as docs and UI issues --
- [ ] This will not break child repos: it doesn't hardcode values (.e.g "kumahq" as a image registry) and it will work on Windows, system specific functions like `syscall.Mkfifo` have equivalent implementation on the other OS --
- [ ] Tests (Unit test, E2E tests, manual test on universal and k8s) --
  - Don't forget `ci/` labels to run additional/fewer tests
- [ ] Do you need to update [`UPGRADE.md`](../blob/master/UPGRADE.md)? --
- [ ] Does it need to be backported according to the [backporting policy](../blob/master/CONTRIBUTING.md#backporting)? ([this](https://github.com/kumahq/kuma/actions/workflows/auto-backport.yaml) GH action will add "backport" label based on these [file globs](https://github.com/kumahq/kuma/blob/master/.github/workflows/auto-backport.yaml#L6), if you want to prevent it from adding the "backport" label use [no-backport-autolabel](https://github.com/kumahq/kuma/blob/master/.github/workflows/auto-backport.yaml#L8) label) --

<!--
> Changelog: skip
-->
<!--
Uncomment the above section to explicitly set a [`> Changelog:` entry here](https://github.com/kumahq/kuma/blob/master/CONTRIBUTING.md#submitting-a-patch)?
-->

[1]: https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue#linking-a-pull-request-to-an-issue-using-a-keyword
