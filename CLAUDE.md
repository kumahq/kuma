# CLAUDE.md

Kuma - CNCF service mesh (Envoy-based) for K8s/VMs. L4-L7 connectivity, security, observability, multi-zone/multi-mesh.

## Commands

```bash
make install              # Install dev tools (runs: mise install + buf dep update)
make build                # Build all components
make format               # Format and generate code (auto-fix)
make check                # Run format + lint, then verify no files changed
make generate             # Generate code (protobuf, policies, resources)
make test                 # Run tests
make test/e2e             # Run E2E tests (slow, requires cluster setup)

make build/kuma-cp        # Build control plane
make build/kuma-dp        # Build data plane
make build/kumactl        # Build CLI

make test TEST_PKG_LIST=./pkg/xds/...    # Test specific package
UPDATE_GOLDEN_FILES=true make test       # Update golden files
make k3d/start && skaffold dev           # Dev environment
```

## Workflow

**Code quality principles** (CRITICAL):

1. **Minimal, surgical changes only**
   - Modify ONLY what's necessary; keep PRs small (<10 files)
   - Don't refactor unrelated code "while you're there"

2. **Follow existing patterns**
   - Study surrounding code before making changes
   - Match naming conventions, error handling, structure
   - When in doubt, ask for clarification

3. **No bloated changes**
   - No unnecessary abstractions or `utils.go`/`helpers.go` files
   - No new dependencies without explicit discussion
   - Remove unused code you encounter, but don't hunt for it

4. **Quality gates** (code changes only, skip for docs/config/CI)
   - `make format` to auto-fix, then `make check` to validate (fails if files changed)
   - Fix ALL linting issues (config: `.golangci.yml`, see `.claude/rules/linting.md`)
   - `git diff` to verify changes are focused

### Standard workflow

1. `make check` to ensure clean starting state
2. Read existing code in relevant `pkg/` directory
3. Write tests first (`*_test.go`, Ginkgo/Gomega)
4. Implement minimal, focused changes
5. `make format && make check && make test` to validate

## Git & PRs

```bash
git push --no-verify <remote> branch-name    # ALWAYS use --no-verify
```

- **Commit format**: `type(scope): description`, e.g., `feat(xds):`, `fix(meshtrace):`, `chore(deps):`
- **Commit signing**: use `git commit -s` (DCO `Signed-off-by:` required by CI)
- **Active branches**: see `active-branches.json`
- **Base branch**: `master`
- **PR template**: `.github/PULL_REQUEST_TEMPLATE.md`
- **Changelog**: from PR title or `> Changelog: {<desc>,skip}`
- **CI labels**: MUST use `gh pr create --label "ci/..."`. Labels added after creation are ignored by CI
- **MADR**: `docs/madr/decisions/000-template.md` for features/architecture decisions
- **Downstream refs**: say "downstream project" or "enterprise fork". Never mention Kong Mesh in PRs/commits (private repo)

## Architecture

### Components

- `kuma-cp`: control plane, serves xDS configs to data planes, manages mTLS certs, coordinates zones
- `kuma-dp`: data plane proxy (wraps Envoy), connects to `kuma-cp` for bootstrap config
- `kumactl`: CLI for `kuma-cp` REST API
- `kuma-cni`: CNI plugin for transparent proxy setup on K8s

### Key directories

- `pkg/core/`: core resource types and managers
- `pkg/xds/`: Envoy xDS config generation (listeners, clusters, routes, endpoints)
- `pkg/kds/`: Kuma Discovery Service, syncs resources between Global CP and Zone CPs
- `pkg/api-server/`: REST API server (validate external inputs here)
- `pkg/dp-server/`: data plane server (SDS, health checks)
- `pkg/plugins/policies/`: policy plugins; see `.claude/rules/policies.md`
- `pkg/plugins/`: plugin architecture (bootstrap, runtime, CA, resources, policies)
- `pkg/transparentproxy/`: transparent proxy (iptables, eBPF)
- `pkg/config/`: configuration management
- `pkg/test/`: test utilities, matchers, and resource builders
- `tools/`: code generation tools (policy-gen, resource-gen, openapi)

### Multi-zone

- Global CP coordinates Zone CPs via KDS (`pkg/kds/`)
- Zone Ingress/Egress handle cross-zone traffic
- Supports Kubernetes and Universal (VM/bare metal)

### Code generation

Run `make generate` after changes to `.proto` files, `pkg/plugins/policies/*/api/` dirs, or resource definitions. Generated files: `zz_generated.*`, `*.pb.go`, `*.pb.validate.go`, `rest.yaml`

### Tool management

Uses `mise` (config: `mise.toml`). Install via `make install`. Includes: buf, ginkgo, helm, kind, kubectl, protoc, golangci-lint, skaffold.

## Testing

- **Framework**: Ginkgo (BDD) + Gomega assertions (`*_test.go`); see `.claude/rules/testing.md`
- **Golden matchers**: `MatchGoldenYAML`, `MatchGoldenJSON`, `MatchGoldenEqual`, `MatchProto`
- **Golden files**: in `testdata/` dirs. Update with `UPDATE_GOLDEN_FILES=true make test`
- **Mocks**: hand-written stubs (no mockgen/counterfeiter). Implement the interface in the test file
- **E2E**: `make test/e2e`. Slow, requires cluster setup

## Gotchas

- **RBAC gate**: if `deployments/` RBAC manifests change, `UPGRADE.md` must also be updated or `make check` fails
- **Import aliases required**: `.golangci.yml` enforces `core_mesh`, `mesh_proto`, `core_model`, `common_api`, etc. See `.claude/rules/linting.md`
- **Cached resources are read-only**: `pkg/core/resources/manager/cache.go` returns shared instances. Never modify them
- **`pkg/` cannot import `app/`**: architectural boundary enforced by depguard linter

## Common issues

- Missing dependencies → `make install`
- Outdated generated code → `make generate`
- Formatting errors → `make format` (auto-fixes), then `make check` (validates)
- Golden file mismatches → `UPDATE_GOLDEN_FILES=true make test`

## Notes

- Versions: `active-branches.json` and `versions.yml` in the repo root
- Docs: `DEVELOPER.md` (setup), `CONTRIBUTING.md` (PR workflow)
- Release process: see `.claude/rules/release.md`
- Local dev and debugging: see `.claude/rules/debug.md`
