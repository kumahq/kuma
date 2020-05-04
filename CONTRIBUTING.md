# Contributing to Kuma

Hello, and welcome! Whether you are looking for help, trying to report a bug,
thinking about getting involved in the project or about to submit a patch, this
document is for you! Its intent is to be both an entry point for newcomers to
the community (with various technical backgrounds), and a guide/reference for
contributors and maintainers.

Consult the Table of Contents below, and jump to the desired section.

## Table of Contents

- [Contributing to Kuma](#contributing-to-kuma)
  - [Table of Contents](#table-of-contents)
  - [Where to seek help?](#where-to-seek-help)
  - [Where to report bugs?](#where-to-report-bugs)
  - [Where to submit feature requests?](#where-to-submit-feature-requests)
  - [Contributing](#contributing)
    - [Improving the documentation](#improving-the-documentation)
    - [Submitting a patch](#submitting-a-patch)
      - [Git branches](#git-branches)
      - [Commit atomicity](#commit-atomicity)
      - [Commit message format](#commit-message-format)
        - [Type](#type)
        - [Scope](#scope)
        - [Subject](#subject)
        - [Body](#body)
        - [Footer](#footer)
        - [Examples](#examples)
      - [Writing tests](#writing-tests)
    - [Contributor T-shirt](#contributor-t-shirt)

## Where to seek help?

[Slack](https://kuma-mesh.slack.com) is the main chat channel used by the
community and the maintainers of this project. If you do not have an
existing account, please follow this [link](https://chat.kuma.io) to sign
up for free.

**Please avoid opening GitHub issues for general questions or help**, as those
should be reserved for actual bug reports. The Kuma community is welcoming and
more than willing to assist you on Slack!

[Back to TOC](#table-of-contents)

## Where to report bugs?

Feel free to [submit an issue](https://github.com/Kong/kuma/issues/new) on
the GitHub repository, we would be grateful to hear about it! Please make sure
to respect the GitHub issue template, and include:

1. A summary of the issue
2. A list of steps to reproduce the issue
3. The version of Kuma you encountered the issue with
4. Your Kuma configuration, or the parts that are relevant to your issue

If you wish, you are more than welcome to propose a patch to fix the issue!
See the [Submit a patch](#submitting-a-patch) section for more information
on how to best do so.

[Back to TOC](#table-of-contents)

## Where to submit feature requests?

You can [submit an issue](https://github.com/Kong/kuma/issues/new) for feature
requests. Please add as much detail as you can when doing so.

You are also welcome to propose patches adding new features. See the section
on [Submitting a patch](#submitting-a-patch) for details.

[Back to TOC](#table-of-contents)

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

[Back to TOC](#table-of-contents)

### Improving the documentation

The documentation hosted at https://kuma.io/docs is open source and built
with [Netlify](https://www.netlify.com/). You are welcome to propose changes to it
(correct typos, add examples or clarifications...)!

The repository is also hosted on GitHub at:
https://github.com/Kong/kuma-website

[Back to TOC](#table-of-contents)

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
- The tests are passing: run `make test`, `make test/kuma-cp`,
  `make test/kuma-dp`, or whichever make target under `test/` is appropriate
  for your change
- Do not update CHANGELOG.md yourself. Your change will be included there in
  due time if it is accepted, no worries!

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

[Back to TOC](#table-of-contents)

#### Git branches

As a best practice to keep your development environment as organized as possible, create local branches to work within. These should also be created directly off of the master branch. If you have write access to the GitHub repository, please follow the following
naming scheme when pushing your branch(es):

- `feat/foo-bar` for new features
- `fix/foo-bar` for bug fixes
- `tests/foo-bar` when the change concerns only the test suite
- `refactor/foo-bar` when refactoring code without any behavior change
- `style/foo-bar` when addressing some style issue
- `docs/foo-bar` for updates to the README.md, this file, or similar documents

[Back to TOC](#table-of-contents)

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

[Back to TOC](#table-of-contents)

#### Commit message format

To maintain a healthy Git history, we ask of you that you write your commit
messages as follows:

- The tense of your message must be **present**
- Your message must be prefixed by a type, and a scope
- The header of your message should not be longer than 50 characters
- A blank line should be included between the header and the body
- The body of your message should not contain lines longer than 72 characters

Here is a template of what your commit message should look like:

```
<type>(<scope>) <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

##### Type

The type of your commit indicates what type of change this commit is about. The
accepted types are:

- **feat**: A new feature
- **fix**: A bug fix
- **hotfix**: An urgent bug fix during a release process
- **tests**: A change that is purely related to the test suite only (fixing
  a test, adding a test, improving its reliability, etc...)
- **docs**: Changes to the README.md, this file, or other such documents
- **style**: Changes that do not affect the meaning of the code (white-space
  trimming, formatting, etc...)
- **perf**: A code change that significantly improves performance
- **refactor**: A code change that neither fixes a bug nor adds a feature, and
  is too big to be considered just `perf`
- **chore**: Maintenance changes related to code cleaning that isn't
  considered part of a refactor, build process updates, or dependency bumps

##### Scope

The scope is the part of the codebase that is affected by your change. Choosing
it is at your discretion, but here are some of the most frequent ones:

- **kuma-cp**: A change that affects the control-plane
- **kuma-dp**: A change that affects the data-plane
- **kumactl**: A change to the kumactl
- **deps**: When updating dependencies (to be used with the `chore` prefix)
- **conf**: Configuration-related changes (new values, improvements...)
- `*`: When the change affects too many parts of the codebase at once (this
  should be rare and avoided)

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
fix(admin) send HTTP 405 on unsupported method

The appropriate status code when the request method is not supported
on an endpoint it 405. We previously used to send HTTP 404, which
is not appropriate. This updates the Admin API helpers to properly
return 405 on such user errors.

* return 405 when the method is not supported in the Admin API helpers
* add a new test case in the Admin API test suite

Fix #678
```

[Back to TOC](#table-of-contents)

#### Writing tests

We use [Ginkgo](https://github.com/onsi/ginkgo) to write our tests. Your patch
should include the related test updates or additions, in the appropriate test
suite.

[Back to TOC](#table-of-contents)

### Contributor T-shirt

If your Pull Request to [Kong/kuma](https://github.com/Kong/kuma) was
accepted, and it fixes a bug, adds functionality, or makes it significantly
easier to use or understand Kuma, congratulations! You are eligible to
receive the very special Contributor T-shirt! Go ahead and fill out the
[Contributors Submissions form](https://goo.gl/forms/5w6mxLaE4tz2YM0L2).

[Back to TOC](#table-of-contents)