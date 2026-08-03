---
excludeAgent: "code-review"
---

# Development Guidelines

## Development Workflow

**Before Coding:**
1. ASK questions until 95% confident
2. Search for similar implementations (Grep/Glob)
3. Read existing code in relevant `pkg/` directory
4. Create plan, get approval (use Plan Mode for complex tasks)
5. Work incrementally (20-50 lines per step)

**Process:**
1. `make check` - clean state (ONLY for code changes, skip for config/docs/CI)
2. Read existing code
3. Write tests first (`*_test.go`, Ginkgo/Gomega)
4. Implement minimal changes
5. `make check && make test` - validate

---

## Go Conventions

### Error Handling

```go
// Custom error types with Is() - NO reflect.TypeOf
func (e *NotFound) Is(err error) bool {
    _, ok := err.(*NotFound)  // Use type assertion
    return ok
}

// Wrap with context
return errors.Errorf("failed to create dataplane: %w", err)

// Type assertions with clear messages
if !ok {
    return errors.Errorf("invalid type: expected=%T, got=%T", expectedType, actualType)
}

// Multi-error (transactions)
import "go.uber.org/multierr"
return multierr.Append(errors.Wrap(rollbackErr, "rollback failed"), err)
```

**Rules:** Always wrap errors • Use custom types with `Is()` • NO `reflect.TypeOf()` • Check `context.Canceled` before rollback • Use `multierr` for multiple errors

### Naming & Organization

**Naming:** Unexported types (`dataplaneManager`) • Exported interfaces (`ResourceManager`) • Descriptive constants

**Patterns:**
- Embedded interfaces for composition: `struct { ResourceManager; store Store }`
- Strategy pattern: `type TimeoutConfigurer struct {...}; func (t) Configure(cluster) error`
- Context propagation: `type ctxKey struct{}`; `context.WithValue(ctx, ctxKey{}, val)`
- Builder for tests: `Dataplane().WithName("x").Build()`

**Comments:** Public APIs only • Explain "why" not "what" • Test markers: `// setup`, `// given`, `// when`, `// then`

### Linter Errors

**Formatter:** `gofmt` (auto via `make check`)
**Linter:** `golangci-lint` (config: `.golangci.yml`)

**ALWAYS:**
- Attempt to fix linter errors properly
- Research solutions online if unclear how to fix
- Fix root cause, not symptoms

**NEVER:**
- Use skip/disable directives (`//nolint`, `// revive:disable`)
- Ignore linter warnings
- Work around linter errors

**If stuck:**
1. Try fixing the error
2. Research online for proper solution
3. If still unclear after research, ASK what to do (don't skip/disable)

---

## Simplicity Principles

### Anti-Patterns to AVOID

❌ **NEVER:**
- Long functions (>50 lines per function)
- Over-engineering (unnecessary abstractions/configurers)
- `reflect.TypeOf()` for type checking
- Generic `interface{}` instead of typed maps
- Panics in production code
- Entire files at once
- Placeholders `// ... rest of code ...`

✅ **ALWAYS:**
- Functions <50 lines, single responsibility
- Search existing patterns FIRST (Grep/Glob)
- Three similar lines > premature abstraction
- Minimal, surgical changes
- Reuse existing components

**Check before implementing:**
1. Does similar code exist? What pattern does it use?
2. Can this be simpler/shorter?
3. Am I following existing plugin/configurer/manager pattern?
4. Is this minimal change to achieve goal?

---

## Code Generation Rules

### NEVER

- Generate long functions (>50 lines per function)
- Generate entire files at once (>100 lines in single response)
- Make big changes in single step
- Use placeholders like `// ... rest of code ...`
- Modify code unrelated to the task
- Assume requirements without asking
- Add features beyond what's requested

### ALWAYS

- Show complete code (no placeholders)
- Incremental changes (20-50 lines per step)
- Surgical, minimal changes only
- Test-driven development (write tests first)
- Follow existing patterns found in codebase
- Ask questions before assuming requirements

### Incremental Development Process

**Break changes into steps:**
1. Define interfaces/types
2. Write tests (input/output pairs)
3. Implement core logic (minimal)
4. Add error handling
5. Run tests, iterate
6. Update golden files if needed

**Each step:** Review, test, approve, then proceed to next.

---

## Complexity Hotspots

| Area | Location | Challenge |
|------|----------|-----------|
| **Policy Matching** | `pkg/plugins/policies/core/matchers/` | Selector evaluation, wildcards, cross-zone |
| **XDS Generation** | `pkg/xds/generator/`, `pkg/xds/envoy/` | Merging policies, precedence, protocol-specific |
| **Multi-zone Sync** | `pkg/kds/` | Eventual consistency, failures, versioning |
| **Rule Resolution** | `pkg/plugins/policies/core/rules/` | Inbound/outbound, subsets, precedence |

**Tips:**
- Study existing matchers/configurers before adding new logic
- Use strategy pattern, keep configurers focused (single concern)
- Test K8s + Universal modes
- Validate mesh isolation and policy boundaries
