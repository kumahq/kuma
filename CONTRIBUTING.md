<!-- Synced from kumahq/.github update lifecycle action (and remove this comment) to stop syncing -->
# Contributing Guide

* [New Contributor Guide](#contributing-guide)
  * [Ways to Contribute](#ways-to-contribute)
  * [Find an Issue](#find-an-issue)
  * [Ask for Help](#ask-for-help)
  * [Pull Request Lifecycle](#pull-request-lifecycle)
  * [Development Environment Setup](#development-environment-setup)
  * [Sign Your Commits](#sign-your-commits)
  * [Pull Request Checklist](#pull-request-checklist)

Welcome! We are glad that you want to contribute to Kuma! 💖

As you get started, you are in the best position to give us feedback on areas of
our project that we need help with including:

* Problems found during setting up a new developer environment
* Gaps in our Quickstart Guide or documentation
* Bugs in our automation scripts

If anything doesn't make sense, or doesn't work when you run it, please open a
bug report and let us know!

## Ways to Contribute

We welcome many types of contributions including:

* New features
* Builds, CI/CD
* Bug fixes
* Documentation
* Issue Triage
* Answering questions on Slack/Mailing List
* Web design
* Communications / Social Media / Blog Posts
* Release management

Not everything happens through a GitHub pull request. Please check
the [community page on kuma.io](https://kuma.io/community/). 

### Come to meetings!
Absolutely everyone is welcome to come to any of our meetings. You never need an
invitation to join us. In fact, we want you to join us, even if you don’t have
anything you feel like you want to contribute. Just being there is enough!

You can find out more about our meetings [here](https://kuma.io/community/).
You don’t have to turn on your video.
The first time you come, introducing yourself is more than enough.
Over time, we hope that you feel comfortable voicing your opinions, giving
feedback on others’ ideas, and even sharing your own ideas, and experiences.

## Find an Issue

We have good first issues for new contributors and help wanted issues suitable
for any contributor.
[good first issue](https://github.com/search?q=org%3Akumahq+label%3A%22good+first+issue%22+state%3Aopen&type=Issues) has extra information to
help you make your first contribution. [help wanted](https://github.com/search?q=org%3Akumahq+label%3A%22help+wanted%22+state%3Aopen&type=Issues) are issues
suitable for someone who isn't a core maintainer and is good to move onto after
your first pull request.

Sometimes there won’t be any issues with these labels. That’s ok! There is
likely still something for you to work on. If you want to contribute but you
don’t know where to start or can't find a suitable issue, you can ask in the #development channel in [the Kuma Slack](https://kuma-mesh.slack.com/).

Once you see an issue that you'd like to work on, please post a comment saying
that you want to work on it. Something like "I want to work on this" is fine.

You might want to familiarize yourself with our [Triage policy](https://github.com/kumahq/.github/blob/main/PROJECT_MANAGEMENT.md#triage).

## Ask for Help

The best way to reach us with a question when contributing is to ask on:

* The original github issue
* The developer mailing list
* [Kuma Slack #developer](https://join.slack.com/t/kuma-mesh/shared_invite/zt-1rcll3y6t-DkV_CAItZUoy0IvCwQ~jlQ) 


## Pull Request Lifecycle

- Use freely Draft PRs but don't except anyone to review or look at it unless you explicitly mention them.
- If you create a regular PR, reviewers will review it so only open PRs that are ready to review.
- If no-one reviews your PR it's ok to ping folks in Slack after 2 business days.
- The reviewer will leave feedback and follow these rules:
  - nits: These are suggestions that you may decide incorporate into your pull request or not without further comment.
  - It can help to put a 👍 on comments that you have implemented so that you can keep track.
  - It is okay to clarify if you are being told to make a change or if it is a suggestion.
- Make changes in new commits (Please don't force push to your PR it's easier to review). Please update your PRs it's ok to have merge commits we'll get rid of them when squashing.
- Ask for a new review by dismissing existing reviews and/or mention the reviewer.
- Please wait 2 business days before pinging the reviewer again.
- When a pull request has been approved, the reviewer will squash and merge your commits. If you prefer to rebase your own commits, at any time leave a comment on the pull request to let them know that.

## Development Environment Setup

See [the dedicated page](./DEVELOPER.md).

## Sign Your Commits

### DCO
Licensing is important to open source projects. It provides some assurances that
the software will continue to be available based under the terms that the
author(s) desired. We require that contributors sign off on commits submitted to
our project's repositories. The [Developer Certificate of Origin
(DCO)](https://developercertificate.org/) is a way to certify that you wrote and
have the right to contribute the code you are submitting to the project.

You sign-off by adding the following to your commit messages. Your sign-off must
match the git user and email associated with the commit.

    This is my commit message

    Signed-off-by: Your Name <your.name@example.com>

Git has a `-s` command line option to do this automatically:

    git commit -s -m 'This is my commit message'

If you forgot to do this and have not yet pushed your changes to the remote
repository, you can amend your commit with the sign-off by running 

    git commit --amend -s 


## Pull Request Checklist

When you submit your pull request, or you push new commits to it, our automated
systems will run some checks on your new code. We require that your pull request
passes these checks, but we also have more criteria than just that before we can
accept and merge it. We recommend that you check the following things locally
before you submit your code:

**TODO**
<!-- list both the automated and any manual checks performed by reviewers, it
is very helpful when the validations are automated in a script for example in a
Makefile target. Below is an example of a checklist:

* It passes tests: run the following command to run all of the tests locally:
  `make build test lint`
* Impacted code has new or updated tests
* Documentation created/updated
* We use [Azure DevOps, GitHub Actions, CircleCI]  to test all pull
  requests. We require that all tests succeed on a pull request before it is merged.

-->
