# Testing Guidelines

## Testing with Ginkgo/Gomega

**Structure:**
```go
var _ = Describe("Feature", func() {
    It("should behave correctly", func() {
        // setup
        manager := NewManager(...)

        // given
        input := Resource{...}

        // when
        err := manager.Create(ctx, &input, key)
        Expect(err).ToNot(HaveOccurred())

        // then
        Expect(actual.Field).To(Equal("expected"))
    })
})
```

**Table-driven:**
```go
DescribeTable("should validate",
    func(yaml string) {
        err := FromYAML([]byte(yaml), &spec)
        Expect(err).ToNot(HaveOccurred())
    },
    Entry("full", `targetRef: ...`),
    Entry("minimal", `targetRef: {kind: Mesh}`),
)
```

**Golden Files:**
```go
Expect(yaml).To(matchers.MatchGoldenYAML("testdata", "file.yaml"))
// Update: UPDATE_GOLDEN_FILES=true make test
```

**Builders:**
```go
dp := test_builders.Dataplane().WithName("dp1").Build()  // Validates, panics on error
```

**Common Mistakes:** ❌ Missing flow comments • Not updating golden files • No validation in builders • Using `err == nil` instead of Gomega matchers
