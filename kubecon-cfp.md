# KubeCon CFP: Hooks, Skills, Speed: AI Agents for Open Source Maintainers

**Track:** AI + ML
**Format:** Session (30 min)
**Title:** Hooks, Skills, Speed: AI Agents for Open Source Maintainers (59 chars)

---

## Version 1 — Patterns-focused

Reproducing a community-reported Service Mesh issue on Kubernetes used to take 1–2 days. With agentic coding, it takes under 30 minutes. This talk covers how a CNCF Service Mesh maintainer team integrated AI coding agents into their daily workflow: agents spin up local clusters, reproduce service mesh failures, and write test suites from problem descriptions. The catch is non-determinism. AI agents speed things up, but their output is unpredictable. Hooks fix this by enforcing commit standards and linting before anything reaches a PR. The session walks through the three patterns behind this workflow: hooks as quality gates, custom skills for repetitive maintenance, and context files that encode codebase conventions for the agent. Both tools are open source: klaudiu.sh for hooks management and the test harness built for the Service Mesh.

---

## Version 2 — Mindset shift-focused

What if your team had multiple engineers who never got bored of spinning up Kubernetes clusters? A community-reported issue that took 1–2 days to reproduce now takes under 30 minutes. That engineer is an agent. This talk is about the adjustments a CNCF Service Mesh maintainer team made to work with one effectively. The work shifted from writing code to writing instructions. Code review gave way to hooks that enforce standards automatically. Debugging stopped being a manual exercise and became a description the agent executes. None of this happened without friction—AI agents are non-deterministic by nature, and the urge to micromanage every output is real. But once the team stopped fighting that and started channeling it through hooks and skills, the agent became a reliable team member. Each shift gets concrete treatment: what changed, what broke, and what the team learned. Two open source tools sit at the center of this workflow: klaudiu.sh, which adds predictability to LLM output through hooks, and the test harness built for the Service Mesh.

---

## Benefits to the Ecosystem

The patterns in this talk work with any AI coding agent. That matters for CNCF maintainers who may be using different tools or switching between them—the architecture is the point, not the specific tool. Open source projects share a real bottleneck: issue reproduction is expensive and contributor bandwidth cannot keep pace with the cognitive overhead of maintaining a production-grade cloud-native project. Agentic coding helps, but most published guides target solo developers on greenfield projects, not multi-maintainer repos with existing CI pipelines, linting rules, and codebase conventions accumulated over years. This talk fills that gap. CNCF maintainers leave with context file templates, skill definitions, and access to klaudiu.sh and the Service Mesh test harness, both open source. For new contributors, the barrier to getting started—understanding a complex codebase, writing tests, navigating CI—drops when an agent handles the orientation work. All configurations are published before the conference.
