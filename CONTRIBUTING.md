# Contributing to Kuma

Hello, and welcome! Whether you are looking for help, trying to report a bug,
thinking about getting involved in the project or about to submit a patch, this
document is for you! Its intent is to be both an entry point for newcomers to
the community (with various technical backgrounds), and a guide/reference for
contributors and maintainers.

## Where to seek help?

[Slack](https://kuma-mesh.slack.com) is the main chat channel used by the
community and the maintainers of this project. If you do not have an
existing account, please follow this [link](https://join.slack.com/t/kuma-mesh/shared_invite/zt-1tu98qcw4-pJNW9lRdBqx1I9kF9rvImA) to sign
up for free.

**Please avoid opening GitHub issues for general questions or help**, as those
should be reserved for actual bug reports. The Kuma community is welcoming and
more than willing to assist you on Slack!

## Where to report bugs?

Feel free to [submit an issue](https://github.com/kumahq/kuma/issues/new) on
the GitHub repository, we would be grateful to hear about it! Please make sure
to respect the GitHub issue template, and include:

1. A summary of the issue
2. A list of steps to reproduce the issue
3. The version of Kuma you encountered the issue with
4. Your Kuma configuration, or the parts that are relevant to your issue

If you wish, you are more than welcome to propose a patch to fix the issue!
See the [Submit a patch](#submitting-a-patch) section for more information
on how to best do so.

## Where to submit feature requests?

You can [submit an issue](https://github.com/kumahq/kuma/issues/new) for feature
requests. Please add as much detail as you can when doing so.

You are also welcome to propose patches adding new features. See the section
on [Submitting a patch](#submitting-a-patch) for details.

## Contributing

We welcome contributions of all kinds, you do not need to code to be helpful!
All of the following tasks are noble and worthy contributions that you can
make without coding:

- Reporting a bug (see the [report bugs](#where-to-report-bugs) section)
- Helping other members of the community on Slack
- Fixing a typo in the code
- Fixing a typo in the documentation at https://kuma.io/docs (see
  the [documentation contribution](#improving-the-documentation) section)
- Providing your feedback on the proposed features and designs
- Reviewing Pull Requests
- Writing a blog post on how you're using Kuma

If you wish to contribute code (features or bug fixes), see the [Submitting a
patch](#submitting-a-patch) section.

### Improving the documentation

The documentation hosted at https://kuma.io/docs is open source and built
with [Netlify](https://www.netlify.com/). You are welcome to propose changes to it
(correct typos, add examples or clarifications...)!

The repository is also hosted on GitHub at:
[kumahq/kuma-website](https://github.com/kumahq/kuma-website)

### Improving the GUI

The GUI code is in the [kumahq/kuma-gui](https://github.com/kumahq/kuma-gui) repository.

### Submitting a patch

Feel free to contribute fixes or minor features, we love to receive Pull
Requests! If you are planning to develop a larger feature, come talk to us
first on [Slack](#where-to-seek-for-help)!

When contributing, please follow the guidelines provided in this document. They
will cover topics such as the different Git branches we use, the commit message
format to use or the appropriate code style.

Once you have read them, and you are ready to submit your Pull Request, be sure
to verify a few things:

- Your work was based on the appropriate branch (`master` vs. `feature/latest`),
  and you are opening your Pull Request against the appropriate one
- Your commit history is clean: changes are atomic and the git message format
  was respected
- Rebase your work on top of the base branch (seek help online on how to use
  `git rebase`; this is important to ensure your commit history is clean and
  linear)
- The code formatting and the files are generated is good: `make check` will help.
- The tests are passing: `make test`, `make test/kuma-cp`,
  `make test/kuma-dp`, or whichever make target under `test/` is appropriate
  for your change
- If your PR are open and some tests are failing due to outdated golden files
  or formatted and generated files are incorrect a maintainer can fix it by adding a
  comment `/format` or `/golden_files`.
- If you are introducing a change that might break on ipv6 or old k8s kubernetes (v1.19.16-k3s1)
  consider creating PR with a label `ci/run-full-matrix` that will trigger the full test matrix
- If your PR doesn't need to run e2e tests or tests at all use: `ci/skip-test` or `ci/skip-e2e-test` labels on the PR.
- If you are introducing a change which requires specific attention when
  upgrading update UPGRADE.md
- Do not update CHANGELOG.md yourself. Your change will be included there when we release, no worries!
  You can however help us maintain a good changelog and release notes by following this process:
    - By default, the changelog generation will use the PR title. If you want to specify something different as a Changelog entry add a line that starts with `> Changelog:` in the PR description e.g.:`> Changelog: feat: my new feature`. Reusing the same content across multiple PRs will rollup all the PRs in a single changelog entry (this is useful for features that span multiple PRs).
    - By default, all commit messages that start with build, ci, test, refactor will be excluded from the Changelog. If you still want a change to be included add a `> Changelog:` entry in the PR description.
    - If you want a change to not be included in the changelog you can add `> Changelog: skip` to explicitly ignore it.
  This is in the PR description because it enables you to add/modify it even after merging a PR. curious about the automation? You can find it in: [`tools/releases/changelog`](tools/releases/changelog).

If the above guidelines are respected, your Pull Request will be reviewed by
a maintainer.

If you are asked to update your patch by a reviewer, please do so! Remember:
**you are responsible for pushing your patch forward**. If you contributed it,
you are probably the one in need of it. You must be prepared to apply changes
to it if necessary.

If your Pull Request was accepted and fixes a bug, adds functionality, or
makes it significantly easier to use or understand Kuma, congratulations!
You are now an official contributor to Kuma. Get in touch with us to receive
your very own [Contributor T-shirt](#contributor-t-shirt)!

Your change will be included in the subsequent release Changelog, and we will
not forget to include your name if you are an external contributor. :wink:

#### Writing tests

We use [Ginkgo](https://github.com/onsi/ginkgo) to write our tests. Your patch
should include the related test updates or additions, in the appropriate test
suite.
Checkout [DEVELOPER.md](DEVELOPER.md) for some extra info related to tests.

#### Sign Your Work

The sign-off is a simple line at the end of the explanation for a commit. All
commits needs to be signed. Your signature certifies that you wrote the patch or
otherwise have the right to contribute the material. The rules are pretty simple,
if you can certify the below (from [developercertificate.org](https://developercertificate.org/)):

To signify that you agree to the DCO for a commit, you add a line to the git
commit message:

```txt
Signed-off-by: Jane Smith <jane.smith@example.com>
```

In most cases, you can add this signoff to your commit automatically with the
`-s` flag to `git commit`. You must use your real name and a reachable email
address (sorry, no pseudonyms or anonymous contributions).

#### Commit atomicity

When submitting patches, it is important that you organize your commits in
logical units of work. You are free to propose a patch with one or many
commits, as long as their atomicity is respected. This means that no unrelated
changes should be included in a commit.

For example: you are writing a patch to fix a bug, but in your endeavour, you
spot another bug. **Do not fix both bugs in the same commit!** Finish your
work on the initial bug, propose your patch, and come back to the second bug
later on. This is also valid for unrelated style fixes, refactors, etc...

You should use your best judgment when facing such decisions. A good approach
for this is to put yourself in the shoes of the person who will review your
patch: will they understand your changes and reasoning just by reading your
commit history? Will they find unrelated changes in a particular commit? They
shouldn't!

Writing meaningful commit messages that follow our commit message format will
also help you respect this mantra (see the below section).

#### Commit message format

We follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
for commit messages and PR titles.

A quick overview on commit messages:

- The tense of your message must be **present**
- Your message must be prefixed by a type, and a scope
- The header of your message should not be longer than 50 characters
- A blank line should be included between the header and the body
- The body of your message should not contain lines longer than 72 characters

Here is a template of what your commit message should look like:

```
<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

The exact details of the format can be found in the [conventional commits
specification](https://www.conventionalcommits.org/en/v1.0.0/#specification).

##### Type

Allowed types are enforced via `commitlint`. The list can be found at the
[`config-conventional` repository](https://github.com/conventional-changelog/commitlint/tree/master/%40commitlint/config-conventional#type-enum).
`fix` and `feat` should only be used for end user visibile contributions. A change to CI, build, a MADR or docs contribution should use the respective `ci`, `build`, `test` types.

##### Scope

The scope is the part of the codebase that is affected by your change. Choosing
it is at your discretion, but here are some of the most frequent ones:

- **kuma-cp**: A change that affects the control-plane
- **kuma-dp**: A change that affects the data-plane
- **kumactl**: A change to the kumactl
- **deps**: When updating dependencies (to be used with the `chore` prefix)
- **api-server**: When changing to admin api
- **xds**: When changing something related to xds
- **kds**: When change something related to kds
- **MeshRetry**: When changing a specific policy
- **policy**: When changing something that affects all policies
- **conf**: Configuration-related changes (new values, improvements...)

The usual casing for scope is to use singular kebab-case except if it's a public api in which
case use whatever the public name is (which usually is PascalCase).
This is why you can have `feat(api-server)` but `fix(MeshRetry)`.

Missing or `*` scope is not allowed. There's usually always something you can put

##### Subject

Your subject should contain a succinct description of the change. It should be
written so that:

- It uses the present, imperative tense: "fix typo", and not "fixed" or "fixes"
- It is **not** capitalized: "fix typo", and not "Fix typo"
- It does **not** include a period. :smile:

##### Body

The body of your commit message should contain a detailed description of your
changes. Ideally, if the change is significant, you should explain its
motivation, the chosen implementation, and justify it.

As previously mentioned, lines in the commit messages should not exceed 72
characters.

##### Footer

The footer is the ideal place to link to related material about the change:
related GitHub issues, Pull Requests, fixed bug reports, etc...

##### Examples

Here are a few examples of good commit messages to take inspiration from:

```
fix(admin): send HTTP 405 on unsupported method

The appropriate status code when the request method is not supported
on an endpoint it 405. We previously used to send HTTP 404, which
is not appropriate. This updates the Admin API helpers to properly
return 405 on such user errors.

* return 405 when the method is not supported in the Admin API helpers
* add a new test case in the Admin API test suite

Fix #678
```

#### Backporting

We have a strict policy on backporting changes to stable branches.
We apply this policy to simplify upgrades between patch versions by keeping the delta between patch versions as small as possible.

Changes that are to be backported should only be critical bug fixes of one of these types:

- Loss of data
- Memory corruption
- Panic, crash, hang
- Security
- CI/CD (anything related to the release process)

If you think your PR applies and should be backported please add the label: `backport`.
Once the PR is approved and merged the action `backport.yaml` will open a new PR with the backport for each of the maintained branches. If you backport a change it's your responsibility to make sure the backports goes through.

#### Reviewing

You can indicate that you are reviewing a PR by using the `eyes` emoji on the PR description.
If you give up on doing so please remove the emoji.

### Contributor T-shirt

If your Pull Request to [kumahq/kuma](https://github.com/kumahq/kuma) was
accepted, and it fixes a bug, adds functionality, or makes it significantly
easier to use or understand Kuma, congratulations! You are eligible to
receive the very special Contributor T-shirt! Go ahead and fill out the
[Contributors Submissions form](https://goo.gl/forms/5w6mxLaE4tz2YM0L2).
