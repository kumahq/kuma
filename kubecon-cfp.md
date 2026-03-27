# KubeCon CFP: Hooks, Skills, Context: AI Agents for Open Source Maintainers

**Track:** AI + ML
**Format:** Session (30 min)
**Title:** Hooks, Skills, Context: AI Agents for Open Source Maintainers (59 chars)

---

## Version 1 — Patterns-focused

Reproducing a community-reported Service Mesh issue on Kubernetes used to take 1–2 days. With agentic coding, it takes under 30 minutes. This talk covers how a CNCF Service Mesh maintainer team integrated AI coding agents into their daily workflow: agents spin up local clusters, reproduce service mesh failures, and write test suites from problem descriptions. The catch is non-determinism. AI agents speed things up, but their output is unpredictable. Hooks fix this by enforcing commit standards and linting before anything reaches a PR. The session walks through the three patterns behind this workflow: hooks as quality gates, custom skills for repetitive maintenance, and context files that encode codebase conventions for the agent. Both tools are open source: klaudiu.sh for hooks management and the test harness built for the Service Mesh.

---

## Version 2 — Mindset shift-focused

What if your team had multiple engineers who never got bored of spinning up Kubernetes clusters? Those engineers are agents. A community-reported Kuma issue that used to take 1–2 days to reproduce now takes under 30 minutes. This case study follows a CNCF Kuma Service Mesh maintainer team's journey of embracing agentic AI for day-to-day work—shifting from tedious, manual work to fast-paced, impact-focused delivery. But that shift cannot happen without proper tooling. Agents produce non-deterministic results: they write code that doesn't compile, commit messages that fail linting, tests that assert the wrong thing. Hooks are what made it work—shell commands that run at fixed points in the workflow, turning best practices into enforced rules rather than suggestions the LLM can ignore. This talk shows how the team uses context files, skills, and hooks to increase productivity and make their work more impactful for the product.

---

## Benefits to the Ecosystem

These patterns work with any AI coding agent—the hooks-and-instructions model transfers regardless of which tool generates the code. Open source projects share a real bottleneck: issue reproduction is expensive and contributor bandwidth cannot keep up with the mental load of maintaining a production-grade cloud-native project. Agentic coding helps, but most tooling assumes greenfield solo work. This case study covers the harder case: multi-maintainer repos with existing CI pipelines, linting rules, and codebase conventions an agent must respect rather than override. CNCF maintainers leave with open sourced configurations: instruction templates, hook definitions, and a Kubernetes test harness ready to adapt.
