# Adopt MADR 

* Status: accepted 

Technical Story: N/A 

## Context and Problem Statement

With the growth of Kuma it has now become necessary to establish a better process for designing important changes.
So far we've been using [proposals](/docs/proposals/README.md) but they lack of a good format.

## Decision Drivers

* Keep an historical record of decisions, what motivated them and what alternatives were taken in consideration 
* Have a standard way to describe new designs and facilitate reviews and discussion. 
* Help users to provide input to a change and how it might impact their existing setup. 

## Considered Options

* [MADR](https://github.com/adr/madr) 
* [KEPs](https://github.com/kubernetes/enhancements/tree/master/keps) 
* Github issues

## Decision Outcome

Chosen option: MADR, because of familiarity and simplicity. 

### Process

#### When to open a MADR

- if a ticket is marked as `kind/design` during triage it means that it was decided a design document was required.
- if anything touches the public api it should have a document as well. This includes:
  - Kuma api
  - Kubernetes CRDs
  - Metrics
  - Upgrade path
  - kumactl args and options

> You should be opening a MADR if you are not thinking of shortly working on implementing it.
> If you have an idea but no time to implement it open an issue instead.

#### Open a new MADR

- Always use the [template](/docs/madr/decisions/000-template.md) and pick the next number.
- Update the [README](/docs/madr/README.md) with a link to the MADR and a 1 line description of the subject.
- Socialize your MADR in Kuma Slack.

> Prior to open a MADR it might be useful to have an informal chat with committers as they may have suggestions/ideas.

#### Process to reach consensus on a MADR

The goal of a MADR is to get to a consensus to either accept or reject a design.
The person that opens the MADR has the responsibility to socialize and gather consensus.
They may need to call a meeting once it's ready to iron out the final points.

#### When is a MADR accepted

MADR must receive at least 3 approvals and no veto to be accepted.

Only votes from committers do count.
Pull Request reviews are used to submit approvals or vetos.

Once a MADR is accepted it is merged in the repo.

The MADR has one of the following status: `rejected | accepted | deprecated | superseded by [ADR-0005](0005-example.md) | implemented in ....`

It's recommended to open the PR with the status `accepted` as it's very likely what you are trying to get to.
The status can be updated once it's merged depending on the evolution of Kuma.
Here's a description of when to use each status

- rejected: If after a discussion we decided to go against a design or a decision it's still good to keep a trace, we'll therefore merge it .
- accepted: This has been reviewed and accepted, but it's not implemented yet, or it's just a process change.
- deprecated: We did it but the feature was removed. It's important to add information about release that will either remove the feature or point to the PR and version that's removing it.
- superseded: There's another MADR that seems to be a better approach to this problem at stake.
- implemented in: The PR/Issue and Kuma version this feature was added. If this is a new process some information on the first times this is done.

#### What happens when a MADR is accepted

When a MADR is merged if there was a design issue, implementation issues should be created and mentioned in the design issue.

### Positive Consequences

* We'll track rejected options 
* There will be a single place to have an overview of what's being developed.

### Negative Consequences

* Submitting changes will be slower.

## Pros and Cons of the Options

### KEP 

While the KEP process is very complete it's pretty heavy.
It seems that for the moment it's simpler to use a less complete process.

### Github issues 

Github issues are very lightweight. 

- However, it's harder to find what is a design issue and what is an implementation issue.
- Long exchanges are not threaded.
- It's hard to decide what's the state of a design from issues.

While it's possible to use labels for a lot of these.
It seems that using files in a VCS is a little more straightforward for discovery and persistence.
