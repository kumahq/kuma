# MADR-081 Formalizing security update policy for dependencies

* Status: accepted

## Context and problem statement

There is currently no formal policy for patching insecure dependencies in Kuma. Our goal has been to patch all vulnerabilities classified as **High** or **Critical** based on CVSS 3.0 for all supported versions. Supported versions are defined by the release branches listed in [`active-branches.json`](../../../active-branches.json).

A consistent and scoped policy is needed to ensure we handle security updates effectively, especially across multiple release branches, without introducing unnecessary noise or instability.

## Scope of coverage

This policy applies to three categories of code:

* Kuma's own source code
* Third-party code directly linked by Kuma (e.g. Envoy, CoreDNS)
* Third-party code bundled in the convenience Docker image (e.g. Python, Perl, cURL)

For third-party code directly linked by Kuma, we only act on vulnerabilities that have a confirmed CVE.

For third-party code bundled in the convenience Docker image, vulnerabilities are only addressed during the regular release process. These components are not part of Kuma itself, and vulnerabilities in them are not considered exploitable during normal operation. Convenience images are rebuilt with all available patches during each release, but may accumulate vulnerabilities over time. Users are encouraged to build their own images using secure, trusted base images. Users relying on the provided images should always upgrade to the latest patch release of their Kuma version to receive the most recently rebuilt and patched images. We do not patch convenience images outside the release cycle, except in rare cases reviewed by the team, such as a CVE with clearly critical impact in a base image. These are exceptions, and there are no guarantees or formal promises.

### Out of scope

For this document, we refer to a **Kuma vulnerability** as a vulnerability in Kuma's own source code. Whether a vulnerability in a third-party component should also be classified as a Kuma vulnerability (and whether it should be assigned a Kuma CVE) is out of scope for this MADR and will be handled separately if needed.

This MADR does not cover whether we are using the slimmest possible base images for our convenience images. That investigation, including the possibility of switching to slimmer base images to reduce exposure, will be handled in a separate technical story and addressed by a different MADR.

## CVSS 3.0 severity definitions used in this policy

This policy uses the [Common Vulnerability Scoring System v3.0](https://www.first.org/cvss/v3-0/specification-document) to assess the severity of vulnerabilities. The following thresholds apply:

| Severity | CVSS 3.0 Score Range | Notes                                                         |
|----------|----------------------|---------------------------------------------------------------|
| Critical | 9.0–10.0             | Severe impact, often easily exploitable, requires urgent fix  |
| High     | 7.0–8.9              | Major impact, exploitable under certain conditions            |
| Medium   | 4.0–6.9              | Limited impact or harder to exploit                           |
| Low      | 0.1–3.9              | Minimal impact, often with difficult or unlikely exploitation |
| None     | 0.0                  | Informational only, not considered a vulnerability            |

CVEs classified as **None** severity are excluded from this policy.

## Driving factors

* Some vulnerabilities are only fixed in newer minor or major versions, requiring additional changes such as updating the Go version or refactoring due to changed APIs or transitive dependencies
* Updates targeting **Medium** or **Low** severity CVEs sometimes require broad dependency changes, which are not suitable for patch releases
* Disruptive updates can introduce incompatibilities or regressions that are difficult to evaluate or justify when not addressing **Critical** or **High** severity vulnerabilities
* CVEs (regardless of severity) that cannot be exploited in the context of Kuma might still be patched if the update is clean and non-disruptive, but will not trigger a patch release on their own

These issues lead to inconsistent and ad hoc decision-making around update PRs, and a tendency to either over-accept or completely ignore such changes. A clear policy is needed.

## Decision

We adopt the following best-effort policy for patching security vulnerabilities.

### What we patch

For all supported release branches, we aim to patch:

* CVEs classified as **Critical** or **High**
* Confirmed CVEs only (for third-party code linked directly by Kuma)
* Vulnerabilities in the convenience Docker image components, but only during scheduled patch releases

### Confirmed CVEs

In this policy, a **Confirmed CVE** is a vulnerability that has an assigned CVE ID and meets at least one of the following conditions:

* It is published in the [GitHub Advisory Database](https://github.com/advisories), in which case it must have been reviewed by GitHub
* It is published in the [OSV Vulnerability Database](https://osv.dev/list)

### Evaluating impact on Kuma

When no patch is available or cannot be adopted, we evaluate the impact of the vulnerability on Kuma using the following steps:

* Investigate whether the vulnerability is relevant to Kuma.
* Analyze if the issue is realistically exploitable in the context of a typical Kuma deployment.
* Identify the required conditions, configuration, or paths that must exist for the exploit to be possible.
* Recommend workarounds, mitigations, or usage guidance where applicable.
* Summarize findings in a [GitHub security advisory](https://github.com/kumahq/kuma/security/advisories) if applicable.

The structure and format of advisories will be defined as part of a separate effort tracked in [kumahq/kuma#13917](https://github.com/kumahq/kuma/issues/13917).

### Handling linked third-party components

When we refer to a CVE as **Critical** or **High** in linked third-party components, we mean the severity assigned to the original CVE in that component. This is based on [CVSS 3.0](https://www.first.org/cvss/v3-0/specification-document). The severity reflects the impact of the vulnerability in the context of the third-party component itself, not in Kuma.

We handle these situations on a best-effort basis:

* If an upstream fix is released, we will update the component.
* If no fix is available within a reasonable timeframe, or if the available fix cannot be adopted in timely manner, we follow a process of due diligence (see [Evaluating impact on Kuma](#evaluating-impact-on-kuma)).

Decisions around what qualifies as a vulnerability **in Kuma**, how CVEs are assigned, and how advisories are structured will be covered in separate MADRs.

### Exceptions: patching or forking components

We usually do not attempt to patch or modify vulnerable third-party components ourselves. However, in rare cases and based on a case-by-case team decision, we may choose to do so. This can include:

* Preparing our own fix or patch for the affected dependency
* Forking the dependency to include the fix while making a best-effort attempt to upstream the patch
* Using the forked version in Kuma until the official upstream includes the fix
* Avoiding unnecessary effort to drive the upstream process forward beyond what is reasonable

These exceptional actions will be considered only when the issue is significant, no upstream fix is available in a timely manner, and the patch is feasible with limited impact.

**Each exception must be covered in a separate, simple MADR**. That document should describe the context, options considered, pros, cons, risks, and the final decision. It must also include a plan for removing the fork or reverting the exception once the upstream fix is available or the reason for the deviation no longer applies. This ensures transparency and consistency when deviating from the standard policy.

### Other CVEs

The following types of updates are evaluated and accepted by the team on a case-by-case basis. They will not trigger a patch release on their own and will only be included if other important changes are already planned:

* CVEs classified as **Medium** or **Low**, if the update is clean, does not involve disruptive cascading updates, does not introduce breaking changes, and does not require refactoring
* CVEs determined to be non-exploitable in the context of Kuma, based on the process described in [Evaluating impact on Kuma](#evaluating-impact-on-kuma), if the patch is minimal and non-disruptive

### Update acceptance criteria

We will accept dependency updates that meet one or more of the following conditions:

* Fix a CVE classified as **Critical** or **High**. For such updates, the team is committed to making the necessary effort to adopt the fix, even if it involves breaking changes, disruptive cascading updates, or requires significant refactoring
* Address **Medium** or **Low** severity vulnerabilities, or CVEs non-exploitable in the context of Kuma, **only if** the update is minimal, well-scoped, and does not require broad dependency changes (such as major library or toolchain upgrades)
* Apply cleanly to the target branch without introducing instability, meaning no breaking changes, refactoring, or compatibility issues are required

We will not accept updates that:

* Only address **Medium** or **Low** severity CVEs or non-exploitable issues and introduce broad or disruptive dependency changes
* Introduce incompatibilities or regressions inappropriate for a patch release without justification

## Consequences

* We have a clear, written policy on what security updates we will address
* This policy applies uniformly across all supported release branches
* **Critical** and **High** severity updates are prioritized even if complex or disruptive
* **Medium** and **Low** severity or non-exploitable issues are only accepted when minimal and non-invasive
* Patch releases are not triggered by non-exploitable or lower severity CVEs unless other important updates are included
* Users of the convenience Docker image are responsible for adopting updated images or building their own using secure base layers
