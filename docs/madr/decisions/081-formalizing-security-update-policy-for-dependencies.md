# MADR-081 Formalizing security update policy for dependencies

* Status: accepted

## Context and problem statement

There is currently no formal policy for patching insecure dependencies in Kuma. Our goal has been to patch all vulnerabilities with CVSS 3.0 score ≥ 7.0 (High or Critical) across all supported versions. Supported versions are defined by the release branches listed in [`active-branches.json`](../../../active-branches.json).

A consistent and scoped policy is needed to ensure we handle security updates effectively, especially across multiple release branches, without introducing unnecessary noise or instability.

## Scope of coverage

This policy applies to three categories of code:

* Kuma's own source code
* Third-party code directly linked by Kuma (e.g. Envoy, CoreDNS)
* Third-party code bundled in the convenience Docker image (e.g. Python, Perl, cURL)

For third-party code directly linked by Kuma, we only act on vulnerabilities that have a confirmed CVE.

For third-party code bundled in the convenience Docker image, vulnerabilities are only addressed during the regular release process. These components are not part of Kuma itself, and vulnerabilities in them are not considered exploitable during normal operation. Convenience images are rebuilt with all available patches during each release, but may accumulate vulnerabilities over time. Users are encouraged to build their own images using secure, trusted base images. Users relying on the provided images should always upgrade to the latest patch release of their Kuma version to receive the most recently rebuilt and patched images. We do not patch convenience images outside of the release cycle.

## Driving factors

* Some vulnerabilities are only fixed in newer minor or major versions, requiring additional changes such as updating the Go version or refactoring due to changed APIs or transitive dependencies
* Updates targeting low or medium severity CVEs sometimes require broad dependency changes, which are not suitable for patch releases
* Disruptive updates can introduce incompatibilities or regressions that are difficult to evaluate or justify when not addressing critical security issues

These issues lead to inconsistent and ad hoc decision-making around update PRs, and a tendency to either over-accept or completely ignore such changes. A clear policy is needed.

## Decision

We adopt the following best-effort policy for patching security vulnerabilities.

### What we patch

For all supported release branches, we aim to patch:

* CVEs with CVSS 3.0 score ≥ 7.0 (High or Critical)
* Confirmed CVEs only (for third-party code linked directly by Kuma)
* Vulnerabilities in the convenience Docker image components only during scheduled releases

CVEs with CVSS 3.0 score < 7.0 may also be patched if the update is clean, does not involve disruptive cascading updates, does not introduce breaking changes, and does not require refactoring. There is no formal policy or guarantee for such cases. These are evaluated and accepted on a case-by-case basis by the team.

### Update acceptance criteria

We will accept dependency updates that meet one or more of the following conditions:

* Fix a CVE with CVSS 3.0 score ≥ 7.0. For such updates, the team is committed to making the necessary effort to adopt the fix, even if it involves breaking changes, disruptive cascading updates, or requires significant refactoring.
* Address lower severity vulnerabilities (CVSS < 7.0) **only if** the update is minimal, well-scoped, and does not require broad dependency changes (such as major library or toolchain upgrades).
* Apply cleanly to the target branch without introducing instability, meaning no breaking changes, refactoring, or compatibility issues are required.

We will not accept updates that:

* Only address CVEs below High severity and introduce broad or disruptive dependency changes
* Introduce incompatibilities or regressions inappropriate for a patch release without justification

## Consequences

* We have a clear, written policy on what security updates we will address
* This policy applies uniformly across all supported release branches
* Critical updates are prioritized even if complex or disruptive
* Lower severity updates are only accepted when they are minimal and safe
* Users of the convenience Docker image are responsible for adopting updated images or building their own using secure base layers
