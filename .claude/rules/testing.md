# Testing patterns

## Suite structure

Each package has `*_suite_test.go` with:
```go
func TestXxx(t *testing.T) { test.RunSpecs(t, "Suite Name") }
```
For E2E tests, use `test.RunE2ESpecs()` instead (sets higher Gomega timeouts).

## Imports

Use dot imports for Ginkgo/Gomega:
```go
. "github.com/onsi/ginkgo/v2"
. "github.com/onsi/gomega"
```

## Golden file matchers

- `matchers.MatchGoldenYAML("testdata", "file.yaml")`: YAML comparison
- `matchers.MatchGoldenJSON("testdata", "file.json")`: JSON comparison
- `matchers.MatchGoldenEqual("testdata", "file.txt")`: raw string equality
- `matchers.MatchGoldenXML("testdata", "file.xml")`: XML comparison
- `matchers.MatchProto()`: protobuf message comparison with detailed diffs

Update all golden files: `UPDATE_GOLDEN_FILES=true make test`

## Table-driven tests

Use `DescribeTable`/`Entry` for parameterized tests:
```go
DescribeTable("description",
    func(input, expected string) { ... },
    Entry("case 1", "input1", "expected1"),
)
```

Auto-load from testdata: `test.EntriesForFolder("testdata/folder")` scans for `.input.yaml` files.

## Test resource builders

Fluent API in `pkg/test/resources/builders/`:
```go
builders.Dataplane().WithName("dp-1").WithMesh("default").WithAddress("127.0.0.1").Build()
```
Pre-built samples in `pkg/test/resources/samples/`.

## Mocks

Hand-written stubs only, no mockgen or counterfeiter. Implement the minimal interface in the test file itself.

## Async testing

- `test.Within(timeout, func() { ... })`: wraps async task with timeout and GinkgoRecover
- `Eventually()` / `Consistently()` for async Gomega assertions
- Spawn goroutines with `defer GinkgoRecover()` for panic safety

## Test ordering

- `Ordered` modifier on `Describe()` for sequential test execution
- `Serial` for individual serial tests
- Default: parallel execution within suites
