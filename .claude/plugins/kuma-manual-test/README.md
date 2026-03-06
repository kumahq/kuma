# kuma-manual-test

Reproducible manual test harness for Kuma service mesh on local k3d clusters. Tracks manifests, commands, and artifacts for full run reproducibility.

## Installation

### Quick install

```bash
claude plugin marketplace add smykla-skalski/sai
claude plugin install kuma-manual-test@smykla-skalski-sai
```

### Manual

```bash
claude --plugin-dir /path/to/sai/claude/kuma-manual-test
```

## Usage

```
/kuma-manual-test [suite-path] [--profile single-zone|multi-zone] [--repo /path/to/kuma] [--run-id ID] [--resume RUN_ID]
```

| Flag | Default | Purpose |
|:-----|:--------|:--------|
| (positional) | - | Suite path or bare name (looked up in `~/.local/share/sai/kuma-manual-test/suites/`) |
| `--profile` | `single-zone` | Cluster profile |
| `--repo` | auto-detect cwd | Path to Kuma repo checkout |
| `--run-id` | timestamp-based | Override run identifier |
| `--resume` | - | Resume a partial run by its run ID |

## Documentation

See [SKILL.md](./skills/kuma-manual-test/SKILL.md) for detailed workflow and configuration.

## License

MIT - See [../../LICENSE](../../LICENSE)
