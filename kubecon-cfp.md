# KubeCon CFP: Hooks, Skills, Speed: AI Agents for Open Source Maintainers

**Track:** AI + ML
**Format:** Session (30 min)
**Title:** Hooks, Skills, Speed: AI Agents for Open Source Maintainers (59 chars)

---

## Version 1 — Patterns-focused

Reproducing a community-reported Service Mesh issue on Kubernetes used to take 1–2 days. With agentic coding, it takes under 30 minutes. This talk covers how a CNCF Service Mesh maintainer team integrated AI coding agents into their daily workflow: agents spin up local clusters, reproduce service mesh failures, and write test suites from problem descriptions. The catch is non-determinism. AI agents speed things up, but their output is unpredictable. Hooks fix this by enforcing commit standards and linting before anything reaches a PR. The session walks through the three patterns behind this workflow: hooks as quality gates, custom skills for repetitive maintenance, and context files that encode codebase conventions for the agent. Both tools are open source: klaudiu.sh for hooks management and the test harness built for the Service Mesh.

---

## Version 2 — Mindset shift-focused

What if your team had multiple engineers who never got bored of spinning up Kubernetes clusters? A community-reported issue that took 1–2 days to reproduce now takes under 30 minutes. Those engineers are agents. This case study follows a CNCF Service Mesh maintainer team through the adjustments needed to work with one effectively. The work shifted from writing code to writing instructions. Code review gave way to hooks that enforce standards automatically. Debugging stopped being a manual exercise and became a description the agent executes. None of this happened without friction. The urge to micromanage every output is real. Agents write code that doesn't compile, commit messages that fail linting, tests that assert the wrong thing. Hooks are what made it work: shell commands that run at fixed points in the agent's workflow, turning best practices into enforced rules rather than suggestions the LLM can ignore. Each shift gets concrete treatment: what changed, what broke, and what the team learned. Two open source tools sit at the center of this workflow: a hooks layer that enforces standards on agent output, and a test harness built for the Service Mesh.

---

## Benefits to the Ecosystem

These patterns work with any AI coding agent—the hooks-and-instructions model transfers regardless of which tool generates the code. Open source projects share a real bottleneck: issue reproduction is expensive and contributor bandwidth cannot keep up with the mental load of maintaining a production-grade cloud-native project. Agentic coding helps, but most tooling assumes greenfield solo work. This case study covers the harder case: multi-maintainer repos with existing CI pipelines, linting rules, and codebase conventions an agent must respect rather than override. CNCF maintainers leave with open sourced configurations: instruction templates, hook definitions, and a Kubernetes test harness ready to adapt.
