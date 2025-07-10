# MADR-081 Formalizing security update policy for dependencies

* Status: accepted

## Terminology and criteria used in this policy

### Common Vulnerabilitity and Exposure (CVE)

In this policy, a **CVE** (or **Confirmed CVE**, used interchangeably) refers to a vulnerability with an assigned CVE Identifier that meets at least one of the following criteria:

* It is published in the [GitHub Advisory Database](https://github.com/advisories), in which case it must have been reviewed by GitHub
* It is listed in the [OSV Vulnerability Database](https://osv.dev/list)

### CVSS 4.0 severity levels

This policy uses the [Common Vulnerability Scoring System v4.0](https://www.first.org/cvss/v4-0/specification-document) to assess the severity of vulnerabilities. The following thresholds apply:

| Severity | CVSS 4.0 Score Range | Notes                                                         |
|----------|----------------------|---------------------------------------------------------------|
| Critical | 9.0–10.0             | Severe impact, often easily exploitable, requires urgent fix  |
| High     | 7.0–8.9              | Major impact, exploitable under certain conditions            |
| Medium   | 4.0–6.9              | Limited impact or harder to exploit                           |
| Low      | 0.1–3.9              | Minimal impact, often with difficult or unlikely exploitation |
| None     | 0.0                  | Informational only, not considered a vulnerability            |

CVEs classified as **None** severity are excluded from this policy.

## Context and problem statement

There is currently no formal policy for patching insecure dependencies in Kuma. While the general goal has been to address all vulnerabilities classified as [**High** or **Critical** under CVSS 4.0](#cvss-40-severity-levels), this has been done informally and inconsistently across versions. A consistent and scoped policy is needed to ensure we handle security updates effectively, especially across multiple release branches, without introducing unnecessary noise or instability.

## Scope of coverage

This policy defines how we handle **security updates for dependencies** in response to vulnerabilities classified as [Confirmed CVEs](#common-vulnerabilitity-and-exposure-cve). It applies to the following categories of code:

* Kuma's own source code
* Third-party code directly linked by Kuma (e.g. Envoy, CoreDNS)
* Third-party code bundled in the convenience Docker image (e.g. Bash, cURL, iptables)

For third-party code bundled in the convenience Docker image, vulnerabilities are only addressed during the regular release process. These components are not part of Kuma itself, and vulnerabilities in them are not considered exploitable during normal operation. Convenience images are rebuilt with all available patches during each release, but may accumulate vulnerabilities over time. Users are encouraged to build their own images using secure, trusted base images. Users relying on the provided images should always upgrade to the latest patch release of their Kuma version to receive the most recently rebuilt and patched images.

We do not patch convenience images outside the release cycle, except in rare cases reviewed by the team on case-by-case basis. These are exceptions, and there are no guarantees or formal promises.

### Out of scope

For this document, we refer to a **Kuma vulnerability** as a vulnerability in Kuma's own source code. Whether a vulnerability in a third-party component should also be classified as a Kuma vulnerability (and whether it should be assigned a Kuma CVE) is out of scope for this MADR and will be handled separately.

This MADR does not cover whether we are using the slimmest possible base images for our convenience images. That investigation, including the possibility of switching to slimmer base images to reduce exposure, will be handled in a separate technical story and addressed by a different MADR.

## Driving factors

* Some vulnerabilities are only fixed in newer minor or major versions, requiring additional changes such as updating the Go version or refactoring due to changed APIs or transitive dependencies
* Updates targeting **Medium** or **Low** severity CVEs sometimes require broad dependency changes, which are not suitable for patch releases
* Disruptive updates can introduce incompatibilities or regressions that are difficult to evaluate or justify when not addressing **Critical** or **High** severity vulnerabilities
* CVEs (regardless of severity) that cannot be exploited in the context of Kuma might still be patched if the update is clean and non-disruptive, but will not trigger a patch release on their own

These issues lead to inconsistent and ad hoc decision-making around update PRs, and a tendency to either over-accept or completely ignore such changes. A clear policy is needed.

## Decision

We adopt the following best-effort policy for patching security vulnerabilities.

### What we patch

For all active release branches defined in [`active-branches.json`](../../../active-branches.json), which represent supported Kuma versions, we aim to apply security patches in the following cases:

* Vulnerabilities discovered in Kuma's own source code, whether or not they are assigned a Confirmed CVE

* **Confirmed CVEs** classified as **Critical** or **High**:

  * Affecting dependencies used in Kuma’s own source code
  * Affecting third-party components directly linked by Kuma (such as Envoy or CoreDNS), as described in [Handling linked third-party components](#handling-linked-third-party-components)

* Vulnerabilities in components bundled in the convenience Docker image, addressed only during scheduled patch releases

### Handling linked third-party components

When we refer to a CVE in linked third-party components, we mean the severity assigned to the original CVE in that component. The severity reflects the impact of the vulnerability in the context of the third-party component itself, not in Kuma.

We handle these situations on a best-effort basis:

* If an upstream fix is released, we will update the component
* If no fix is available within a reasonable timeframe, or if the available fix cannot be adopted in timely manner, we follow a process of due diligence (see [Process for assessing vulnerability impact on Kuma](#process-for-assessing-vulnerability-impact-on-kuma))

### Exceptions: patching or forking components

We usually do not attempt to patch or modify vulnerable third-party components ourselves. However, in rare cases and based on a case-by-case team decision, we may choose to do so. This can include:

* Preparing our own fix or patch for the affected dependency
* Forking the dependency to include the fix while making a best-effort attempt to upstream the patch
* Using the forked version in Kuma until the official upstream includes the fix
* Avoiding unnecessary effort to drive the upstream process forward beyond what is reasonable

These exceptional actions will be considered only when the issue is significant, no upstream fix is available in a timely manner, and the patch is feasible with limited impact.

**Each exception must be covered in a separate, simple MADR**. That document should describe the context, options considered, pros, cons, risks, and the final decision. It must also include a plan for removing the fork or reverting the exception once the upstream fix is available or the reason for the deviation no longer applies. This ensures transparency and consistency when deviating from the standard policy.

### Dependency update acceptance criteria

We will accept dependency security updates that meet one or more of the following conditions:

* Fix a CVE classified as **Critical** or **High**. For such updates, the team is committed to adopting the fix, even if it requires breaking changes, disruptive cascading updates, or significant refactoring.

* Fix a **Medium** or **Low** severity CVE, or a CVE that is [non-exploitable in the context of Kuma](#process-for-assessing-vulnerability-impact-on-kuma), **only if** all the following apply:

  * The update is minimal and well-scoped.
  * No major dependency, toolchain, or language runtime upgrades are needed.
  * The patch applies cleanly and does not require breaking changes or significant refactoring.

* Apply cleanly to the target branch without introducing instability, regressions, or compatibility issues.

We will not accept updates that:

* Only address **Medium** or **Low** severity CVEs

* Address vulnerabilities that are non-exploitable in regular Kuma deployments, **and**:

  * Require large or disruptive dependency changes.
  * Introduce any instability, regressions, or breaking changes without justification.

  In these cases, the findings based on the due diligence process described in [Process for assessing vulnerability impact on Kuma](#process-for-assessing-vulnerability-impact-on-kuma) will be documented in a dedicated GitHub issue. This issue can be used to share details, justification, and reference for future discussions.

### Triggering the patch release process

A patch release will be initiated when one or more of the following conditions are met:

* A **High** or **Critical** severity vulnerability in Kuma’s own source code is fixed, based on the process outlined in the [Process for assessing vulnerability impact on Kuma](#process-for-assessing-vulnerability-impact-on-kuma) section
* A **Confirmed CVE** classified as **High** or **Critical** is fixed in a third-party component directly linked by Kuma, and an official upstream fix is publicly available
* A patch release is already planned for other reasons, and eligible security fixes are available and can be included

In rare cases, the team may choose to initiate a patch release for other types of fixes. These decisions are made on a case-by-case basis and must be justified by a strong technical or security rationale.

### Process for assessing vulnerability impact on Kuma

To evaluate whether a vulnerability in a dependency affects Kuma, the team should follow a structured process to assess its relevance and the conditions under which it could be exploited. This includes:

* Reviewing the vulnerability details and identifying if the affected component is used by Kuma
* Assessing whether the vulnerability is reachable or exploitable in typical Kuma deployments
* Outlining the required conditions, configuration, or runtime paths needed for exploitation to be possible in Kuma
* Determining whether the vulnerability could pose a practical risk in production environments
* Identifying and recommending mitigations, configuration changes, or usage guidelines if applicable* Summarizing the findings in a [GitHub security advisory](https://github.com/kumahq/kuma/security/advisories), if applicable, or in a dedicated GitHub issue documenting the due diligence. The format and structure of advisories will be defined separately as part of [kumahq/kuma#13917](https://github.com/kumahq/kuma/issues/13917)

This process applies regardless of whether a patch is available or can be adopted.

## Consequences

* This policy provides a clear and consistent approach to handling security updates for dependencies.
* It applies uniformly across all active and supported release branches listed in `active-branches.json`.
* **Critical** and **High** severity vulnerabilities are prioritized and addressed, even when the fix is complex, disruptive, or requires significant effort.
* **Medium**, **Low**, or non-exploitable vulnerabilities are only accepted when the fix is minimal, non-invasive, and does not require major changes.
* Patch releases are only triggered by high-impact fixes, such as those for vulnerabilities in Kuma’s own code or confirmed issues in linked third-party components.
* Vulnerabilities determined to be non-exploitable or lower severity do not trigger a release on their own.
* Users relying on the provided convenience Docker images are responsible for upgrading to the latest patched version or building their own using secure base images.
