# Modern Go Error Handling Strategy

* Status: accepted
* Supersedes: [MADR 073 - Error Wrapping](073-error-wrapping.md)

## Context and Problem Statement

Inconsistent error handling in Kuma:
- [MADR 073](073-error-wrapping.md) recommends `fmt.Errorf` + `%w` (stdlib)
- Codebase uses `github.com/pkg/errors` (30+ usages in `pkg/`)
- `pkg/errors` archived 2024 - no maintenance/security updates

Requirements: stack traces, best performance, `errors.Is/As` support, no timeline pressure.

## Design

### Option 1: Standard Library (`fmt.Errorf` + `%w`)

Current MADR 073 approach.

```go
var ErrNotFound = errors.New("not found")
return fmt.Errorf("resource %w: type=%q", ErrNotFound, rt)
```

**Pros:** No deps, Go team maintained, stdlib compatible
**Cons:** No stack traces, 5x slower with `errors.Is()`

### Option 2: Keep `github.com/pkg/errors`

Status quo.

**Pros:** No migration, stack traces
**Cons:** Archived (security risk), 5x perf degradation, author abandoned

### Option 3: Migrate to `github.com/cockroachdb/errors`

Drop-in replacement, actively maintained.

```go
import "github.com/cockroachdb/errors"
return errors.Wrapf(err, "sync failed: zone=%s", zone)
```

**Pros:** Stack traces, better perf, drop-in replacement, maintained, stdlib compatible
**Cons:** External dep, migration effort

### Option 4: Custom Hybrid

Build custom error types with manual stack capture.

**Pros:** Full control
**Cons:** High effort, maintenance burden, reinventing wheel

## Security implications and review

- `pkg/errors`: No security updates, CVE risk
- `cockroachdb/errors`: Active maintenance, production-hardened

## Reliability implications

- CP handles many DPs, error handling in xDS hot paths
- `pkg/errors` + `errors.Is()` = 5x slowdown
- Stack traces essential for multi-zone debugging
- KDS error propagation benefits from stack context

## Implications for Kong Mesh

Inherits Kuma error handling. Migration synchronizable.

## Decision

**Migrate to `github.com/cockroachdb/errors`**

Rationale: Stack traces + performance + maintained + drop-in replacement + stdlib compatible.

### Migration

Gradual, no timeline:
1. `go get github.com/cockroachdb/errors`
2. Update imports: `github.com/pkg/errors` → `github.com/cockroachdb/errors`
3. Convert `fmt.Errorf` → `errors.Wrap` for stack traces
4. Configure golangci-lint to enforce error handling patterns (errorlint, wrapcheck)

### Usage

```go
// Sentinel errors (from MADR 073)
var ErrNotFound = errors.New("not found")

// Construction
func ErrorResourceNotFound(rt, name, mesh string) error {
    return errors.Wrapf(ErrNotFound, "resource: type=%q name=%q mesh=%q", rt, name, mesh)
}

// Checking
if errors.Is(err, ErrNotFound) { /* handle */ }
```

## Notes

Alternative: `gitlab.com/tozd/go/errors` (less adoption)
Stack overhead: negligible, only on error creation

## References

### Go Error Handling

- [Best Practices - JetBrains](https://www.jetbrains.com/guide/go/tutorials/handle_errors_in_go/best_practices/)
- [Working with Errors - Go Blog](https://go.dev/blog/go1.13-errors)
- [Error Wrapping - Bitfield](https://bitfieldconsulting.com/posts/wrapping-errors)

### Sentinel vs Custom Types

- [Go Error Handling Techniques](https://arashtaher.wordpress.com/2024/09/05/go-error-handling-techniques-exploring-sentinel-errors-custom-types-and-client-facing-errors/)
- [Sentinel vs Custom - alesr](https://alesr.github.io/posts/go-errors/)

### fmt.Errorf vs pkg/errors

- [Wrapf vs Errorf - Stack Overflow](https://stackoverflow.com/questions/61933650/whats-the-difference-between-errors-wrapf-errors-errorf-and-fmt-errorf)
- [Can stdlib replace pkg/errors?](https://blog.dharnitski.com/2019/09/09/go-errors-are-not-pkg-errors/)

### Performance & Alternatives

- [errors.Is() 500% slowdown](https://www.dolthub.com/blog/2024-05-31-benchmarking-go-error-handling/)
- [CockroachDB errors library](https://dr-knz.net/cockroachdb-errors-everyday.html)
- [cockroachdb/errors docs](https://pkg.go.dev/github.com/cockroachdb/errors)
