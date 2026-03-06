# kuma-suite-author

Generate test suites for kuma-manual-test by reading Kuma source code. Produces ready-to-run suites with manifests, validation steps, and expected outcomes.

## Installation

### Quick install

```bash
claude plugin marketplace add smykla-skalski/sai
claude plugin install kuma-suite-author@smykla-skalski-sai
```

### Manual

```bash
claude --plugin-dir /path/to/sai/claude/kuma-suite-author
```

## Usage

```
/kuma-suite-author <feature-name> [--repo /path/to/kuma] [--mode generate|wizard] [--from-pr PR_URL] [--from-branch BRANCH]
```

| Flag            | Default         | Purpose                                                 |
| :-------------- | :-------------- | :------------------------------------------------------ |
| (positional)    | -               | Feature or policy name (e.g., `meshretry`, `meshtrace`) |
| `--repo`        | auto-detect cwd | Path to Kuma repo checkout                              |
| `--mode`        | `generate`      | `generate` (full AI) or `wizard` (interactive)          |
| `--from-pr`     | -               | GitHub PR URL to scope the feature from                 |
| `--from-branch` | -               | Git branch to diff against master for scope             |
| `--suite-name`  | derived         | Override output filename                                |

## Features

- **Variant detection** - automatically scans source code for deployment topologies, feature modes, backend types, protocol branching, and other signals that expand a suite beyond the standard G1-G7 groups
- **Worktree/branch verification** - confirms you're in the right repo location and branch before reading code
- **Confirmation wizard** - presents the full suite summary for review before saving, with options to add/remove/edit groups
- **Two modes** - `generate` auto-detects variants then confirms, `wizard` walks through each decision step-by-step

## Documentation

See [SKILL.md](./skills/kuma-suite-author/SKILL.md) for detailed workflow and configuration.

## License

MIT - See [../../LICENSE](../../LICENSE)
