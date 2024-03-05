<!-- Synced from kumahq/.github update lifecycle action (and remove this comment) to stop syncing -->
# Kuma Governance

This document defines governance policies for the Kuma project.

Anyone can become a Kuma contributor simply by contributing to the project, with code, documentation or other means.
As with all Kuma community members, contributors are expected to follow
the [Kuma Code of Conduct](./CODE_OF_CONDUCT.md).

## Voting

All changes are always initiated either by issue creation (organization members) or pull requests (governance, maintainership).
Votes take place using the https://github.com/cncf/gitvote application (Already installed in the kumahq organization).

## Steering committee

Steering committee members demonstrate a strong commitment to the project with views in the interest of the broader Kuma
community.
They are the stewards of the entire Kuma organization and are expected to dedicate thoughtful and serious effort towards
the goal of general success in the ecosystem.

Responsibilities include:

- Own the overall vision of the Kuma project
- Provide guidance to maintainers
- Review and approve core architecture and design changes
- Add/Remove members and approve deleting or archiving repositories in the Kumahq organization
- Regularly attend community meetings
- Facilitate community process (votes, process changes...)

The list of members of the steering committee are in [OWNERS.md](./OWNERS.md). All steering committee members are owners
of the Kumahq github organization.

### Bootstrapping the committee

The bootstrapped committee will consist of the 3 most active committers namely:

- [Jakub Dyszkiewicz](https://github.com/jakubdyszkiewicz)
- [Michael Beaumont](https://github.com/michaelbeaumont)
- [Charly Molter](https://github.com/lahabana)

### Changes to the steering committee

- Members are elected for 2 years.
- Members can step down by submitting an update to [OWNERS.md](./OWNERS.md)
- Reelection and new members are accepted after a 6 weeks voting period and super majority of the steering committee is
  required (at least 2/3 of votes) to make changes.


## Maintainership

The Kuma project consists of multiple repositories.
Each repository is subject to the same governance model, but different maintainers and reviewers.

### Maintainers

Maintainers have write access to the repo.
The current maintainers of a repo can be found in [OWNERS.md](./OWNERS.md) file of the repo.

Maintainers have the most experience with the given repo and are expected to lead its growth and improvement.
Adding and removing maintainers for a given repo is the responsibility of the existing maintainer team for that repo and
therefore does not require approval from the steering committee

This privilege is granted with some expectation of responsibility: maintainers are people who care about the Kuma
project and want to help it grow and improve.
A maintainer is not just someone who can make changes, but someone who has demonstrated his or her ability to
collaborate with the team, get the most knowledgeable people to review code, contribute high-quality code, and follow
through to fix issues (in code or tests).

#### Becoming a Maintainer

To become a maintainer you need to demonstrate the following:

* commitment to the project
    * participate in discussions, contributions, code reviews for substantial time
    * perform code reviews on non-trivial pull requests,
    * contribute to non-trivial pull requests and have them merged into master,
* ability to write good code,
* ability to collaborate with the team,
* understanding of how the team works (policies, processes for testing and code review, etc),
* understanding of the project's code base and coding style.

### Reviewers

Each repository can have a list of reviewers.
Reviewers help maintainers review new contributions.
They're typically newer to the project and interested in working toward becoming a maintainer.
Reviewers may approve but not merge PRs - all PRs must be approved by a maintainer.

The process for adding/removing reviewers is the same as maintainers

The current list of reviewers for each repository (if any) is published and updated in each repo’s OWNERS.md file.

> Note about auto assignment of PR reviewers:
> For simplicity we use [CODEOWNERS](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners) in github
> teams backing codeowners should be name `<repo>-maintainers` and the matching repo should have an OWNERS.md file with a maintainers section.
> This avoids keeping maintainers lists in all repos that are closely related (for example all maintainers of kumahq/kuma are also maintainers of kumahq/kuma-tools).

### Emeritus Maintainers/Reviewers

Any maintainer can become Emeritus maintainer in two ways:

- Asking explicitly by opening a PR to [OWNERS.md](./OWNERS.md) (no vote).
- Someone calling a vote which is open for 5 days with at least one +1 vote for the steering committee and no -1 vote
  from the steering committee.

### Changes in Maintainership
A new maintainer can be proposed by opening a PR (with title `Maintainer Nomination`) to the repository containing the following information:

* nominee's first and last name,
* nominee's email address and GitHub user name,
* an explanation of why the nominee should be a maintainer/reviewer (adding links to significant contributions)

At least two maintainers need to agree with the nomination (or all maintainers if there's a single maintainer).
If no one objects in 5 days, the nomination is accepted and PR is merged.
If anyone objects or wants more information, the maintainers discuss and usually come to a consensus.
If issues can't be resolved, there's a simple majority vote among steering committee members.

Maintainers and reviewers can be removed if 2 maintainers approve and none disapprove. 
Maintainers and reviewers can leave by just submitting a PR to the repository's OWNERS.md (no vote required in this case).

## Organization members

Organization members are people who have `triage` access to the organization.
Community members who wish to become members of the organization should meet the following requirements, which are open to the discretion of the steering committee:

- Have enabled 2FA on their GitHub account.
- Have joined the Kuma slack.
- Are actively contributing to the project. Examples include:
   - opening issues
   - providing feedback on the project
   - engaging in discussions on issues, pull requests, Slack, etc.
   - attending community meetings
   - Have reached out to two current organization members who have agreed to sponsor their membership request.

To do so they need to open an issue in kumahq/kuma showing that they fill the above requirements. Sponsors express their support by adding `+1` as a comment.

## Changes in Governance

All changes in Governance require a 2/3 majority vote by the steering committee.

## Other Changes

Unless specified above, all other changes to the project require a 2/3 majority vote by the steering committee.
Additionally, any maintainer may request that any change require a 2/3 majority vote by the steering committee.
