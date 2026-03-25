# KubeCon CFP: Hooks, Skills, Speed: AI Agents for Open Source Maintainers

**Track:** AI + ML
**Format:** Session (30 min)
**Title:** Hooks, Skills, Speed: AI Agents for Open Source Maintainers (59 chars)

---

## Version 1 — Patterns-focused

Reproducing a community-reported Service Mesh issue on Kubernetes used to take 1–2 days; agentic coding cuts it to under 30 minutes. This talk covers how a CNCF Service Mesh maintainer team integrated AI coding agents into their daily workflow—spinning up local Kubernetes clusters, reproducing service mesh failures, and generating test suites from problem descriptions. AI agents accelerate the workflow but introduce non-determinism; hooks restore predictability by enforcing standards before agentic output reaches a PR. The session walks through three concrete patterns: hooks as quality gates enforcing linting and commit standards on agentic output, custom skills that encapsulate repetitive maintenance work, and context files that encode codebase conventions for an AI agent. Attendees leave with access to two open source tools central to the workflow: klaudiu.sh, which adds predictability to LLM-driven workflows through hooks, and the test harness built for the Service Mesh—ready to adapt for any CNCF project.

---

## Version 2 — Mindset shift-focused

The first instinct when adopting agentic coding is wrong: more output, less control, and a codebase you no longer trust. This talk examines the mindset shifts a CNCF Service Mesh maintainer team made when moving to an AI-native workflow: from writing code to authoring intent, from manually reviewing output to enforcing standards through hooks, and from reproducing failures by hand to describing them for an agent to execute. The shift is not without friction; AI agents introduce non-determinism, and the instinct to micromanage every output must give way to trusting automated quality gates. Reproducing a community-reported Service Mesh issue on Kubernetes used to take 1–2 days—once the team stopped fighting the non-determinism and started channeling it through hooks and skills, it dropped to under 30 minutes. Attendees leave with a framework for AI-native open source contribution and two open source tools to back it: klaudiu.sh for hooks management and the test harness built for the Service Mesh.

---

## Benefits to the Ecosystem

The context-hooks-skills model works with any AI coding agent, making the patterns immediately applicable across the CNCF ecosystem regardless of which tool a team chooses. Open source maintainership faces a common bottleneck: issue reproduction is expensive, contributor bandwidth cannot keep pace, and onboarding takes too long. Agentic coding addresses all three, but the community has little shared knowledge on integrating these tools safely into multi-maintainer workflows. CNCF maintainers leave with context file templates, skill definitions, and direct access to two open source tools: klaudiu.sh for hooks management and the test harness built for the Service Mesh. New contributors will see the barrier to meaningful contribution—understanding a complex codebase, writing tests, navigating CI pipelines—drop significantly with well-configured agents. Every configuration is documented, reusable, and published before the conference, giving maintainers across the CNCF ecosystem working code rather than slides.
