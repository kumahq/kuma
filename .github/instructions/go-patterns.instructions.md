---
excludeAgent: "coding-agent"
---

# Go Code Review Patterns & Best Practices

Source: 100go.co + Real-world OSS patterns

## Two-Phase Review Process (CoVe)

**Phase 1: Initial Analysis (Internal Only)**
- Trace execution paths, identify concerns
- DO NOT output anything yet

**Phase 2: Verification**

Before reporting, verify:
1. Could params/types prevent this issue?
2. Is there defensive code elsewhere?
3. Could this be intentional design?
4. Does broader context explain this?
5. Would this actually fail at runtime?

**If ANY answer is "yes" or "maybe": DISCARD SILENTLY**

**Report ONLY if ALL true:**
1. Traced exact execution path showing issue
2. Verified NO params/guards prevent issue
3. Checked full function context
4. Issue WILL fail at runtime
5. NOT intentional design pattern
6. Verification confirms issue

**No Hedging:**
- NEVER output "actually, upon closer inspection"
- NEVER explain why NOT reporting something
- If uncertain: stay silent
- Only output confirmed high-confidence issues

---

## Critical Issues (95%+ Block)

### Error Handling

**Ignored errors:**
```go
// ❌ Never ignore
_ = cleanup()

// ✅ Handle or document
if err := cleanup(); err != nil {
    log.Error(err)
}
```

**Error wrapping:**
```go
// ❌ Wrong comparison
if err == target { }

// ✅ Use errors.Is/As
errors.Is(err, sentinel)
errors.As(err, &targetType)
```

**Error handling twice:**
```go
// ❌ Log AND return
log.Error(err)
return err

// ✅ Handle once (wrap allows propagation)
return fmt.Errorf("context: %w", err)
```

### Concurrency

**Data races:**
```go
// ❌ Shared state without sync
type Counter struct { count int }
func (c *Counter) Inc() { c.count++ }

// ✅ Use mutex
type Counter struct {
    mu    sync.Mutex
    count int
}
func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}
```

**Goroutine leaks:**
```go
// ❌ No stop mechanism
go func() {
    for {
        work()
    }
}()

// ✅ Context cancellation
ctx, cancel := context.WithCancel(...)
defer cancel()
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            work()
        }
    }
}()
```

**Copying sync types:**
```go
// ❌ Value copy
func process(m sync.Mutex) { }

// ✅ Pointer
func process(m *sync.Mutex) { }
```

### Resource Leaks

**Unclosed resources:**
```go
// ❌ No close
resp, err := http.Get(url)

// ✅ Defer close
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()
```

**Slice memory leaks:**
```go
// ❌ Retains full backing array
func getFirst(huge []Item) Item {
    return huge[0]
}

// ✅ Copy to prevent leak
func getFirst(huge []Item) Item {
    item := huge[0]
    return item
}
```

**Map memory leaks:**
```go
// ❌ Maps grow, never shrink
cache := make(map[K]V)
// ... grows to millions
// ... delete most entries
// Still uses memory!

// ✅ Recreate or use pointers
cache = make(map[K]V)
```

---

## Major Issues (80%+ Change)

### Interface Design

**Interface pollution:**
```go
// ❌ Premature interface
type UserStore interface {
    Get(id int) User
    Save(u User)
}
type userDB struct { }
func NewUserStore() UserStore { return &userDB{} }

// ✅ Concrete type, interface at consumer
type UserDB struct { }
// Consumer defines needed interface
type UserGetter interface { Get(id int) User }
```

**Returning interfaces:**
```go
// ❌ Restricts flexibility
func New() io.Reader { return &myReader{} }

// ✅ Return concrete
func New() *myReader { return &myReader{} }
```

### Testing

**No race detection:**
```bash
# ❌ Missing
go test ./...

# ✅ Always run
go test -race ./...
```

**Sleep in tests:**
```go
// ❌ Flaky
time.Sleep(100 * time.Millisecond)
assert(condition)

// ✅ Use channels/sync
done := make(chan struct{})
go func() {
    work()
    close(done)
}()
<-done
```

**Missing test coverage:**
- Every feature path needs test
- Edge cases explicitly tested
- Negative scenarios included
- Feature combinations validated

### Code Organization

**Duplicate code (3+ occurrences):**
```go
// ❌ Repeated
if err := validate(x); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}
// ... later ...
if err := validate(y); err != nil {
    return fmt.Errorf("validation failed: %w", err)
}

// ✅ Extract
func validateAndWrap(v Validator) error {
    if err := v.Validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}
```

**Nested code:**
```go
// ❌ Deep nesting
if condition {
    if anotherCondition {
        if yetAnother {
            // work
        }
    }
}

// ✅ Early returns
if !condition {
    return
}
if !anotherCondition {
    return
}
if !yetAnother {
    return
}
// work
```

### Naming

**Generic names:**
```go
// ❌ Generic
func process(data interface{}) error
func handle(item *Thing) error

// ✅ Specific
func validatePodSpec(spec *v1.PodSpec) error
func reconcileDeployment(d *apps.Deployment) error
```

**Naming patterns:**
- Verbs for functions: `createUser`, `validateInput`
- Nouns for types: `UserManager`, `ConfigValidator`
- Adjectives for booleans: `isValid`, `hasPermission`
- Match domain language

---

## Minor Issues (70%+ Comment)

### Performance

**Unnecessary allocations:**
```go
// ❌ Allocates every iteration
for _, item := range items {
    buf := make([]byte, size)
    // use buf
}

// ✅ Reuse
buf := make([]byte, size)
for _, item := range items {
    buf = buf[:0]
    // use buf
}
```

**Preallocate when size known:**
```go
// ❌ Grows dynamically
s := []T{}
for i := 0; i < n; i++ {
    s = append(s, item)
}

// ✅ Preallocate
s := make([]T, 0, n)
for i := 0; i < n; i++ {
    s = append(s, item)
}
```

**String concatenation:**
```go
// ❌ In loop
var s string
for _, item := range items {
    s += item.String()
}

// ✅ strings.Builder
var b strings.Builder
b.Grow(estimatedSize)
for _, item := range items {
    b.WriteString(item.String())
}
```

**Wrong data structure:**
```go
// ❌ O(n²)
for _, item := range items {
    for _, target := range targets {
        if item.ID == target.ID { /* ... */ }
    }
}

// ✅ O(n) with map
targetMap := make(map[string]*Target, len(targets))
for _, t := range targets {
    targetMap[t.ID] = t
}
for _, item := range items {
    if target, ok := targetMap[item.ID]; ok { /* ... */ }
}
```

### Documentation

**Missing comments:**
- Comment WHY, not WHAT
- Explain non-obvious decisions
- Document edge cases
- Link to issues/design docs

**Example:**
```go
// Use exponential backoff with jitter to prevent thundering herd
// when multiple clients reconnect simultaneously after network partition.
// See: https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
backoff := time.Duration(rand.Int63n(int64(baseDelay * (1 << attempt))))
```

### Variable Shadowing

```go
// ❌ Shadows outer
client := &http.Client{}
if condition {
    client := &http.Client{Timeout: 5 * time.Second}
    // Uses inner client
}
// Uses outer client (probably not intended)

// ✅ No shadowing
client := &http.Client{}
if condition {
    client = &http.Client{Timeout: 5 * time.Second}
}
```

---

## Quick Reference Table

| Priority | Category | Check |
|----------|----------|-------|
| 🔴 Critical | Errors | Ignored, wrong comparison (errors.Is/As), handled twice |
| 🔴 Critical | Concurrency | Data races, goroutine leaks, copying sync types |
| 🔴 Critical | Resources | Unclosed, slice/map leaks, time.After in loops |
| 🟡 Major | Interfaces | Pollution, producer-side, returning |
| 🟡 Major | Testing | No race flag, sleep usage, missing coverage |
| 🟡 Major | Organization | Duplicates (3+), deep nesting, generic names |
| 🟢 Minor | Performance | Unnecessary allocations, wrong data structure |
| 🟢 Minor | Docs | Missing WHY comments, no edge case docs |

---

## Skip These (Common False Positives)

- `_ = cleanup()` in defer (intentional)
- Generated code (`zz_generated.*`, `*.pb.go`)
- Test builder panics (intentional design)
- CI-covered (format, imports, lints)
- Concrete types for performance (intentional)
- Custom error handling patterns (verified intentional)

---

## Real-World OSS Patterns

**Sources:** kubernetes, prometheus, vitess, istio, grafana, volcano, traefik, temporal

**Priority Order:**
1. **Critical:** Concurrency races, resource leaks, error handling
2. **Major:** Testing quality, naming clarity, code duplication
3. **Minor:** Performance optimization, documentation

**When to Report:**
- Pattern clearly applies
- Improvement measurable (readability, performance, correctness)
- Consistent with project style
- Not over-engineering

**When to Skip:**
- Pattern doesn't fit context
- Would reduce clarity
- Premature optimization
- Project has different conventions
