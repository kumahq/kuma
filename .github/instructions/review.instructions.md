# PR Review Guidelines

## CoVe Methodology (MANDATORY)

**3-Step Process for EVERY review:**

1. **Verify:** Bug or valid approach? Used elsewhere? Intentional? Runtime issue or style?
2. **Answer:** Check codebase patterns, context, execution path
3. **Self-Critique:** False positive? CI handles? >80% confidence? Concrete fix? Precedent? Substance vs style?

**Post only if:** >80% confidence after verification • Not CI-handled • Specific fix provided • Substance over style

**Kuma-Specific Verification:**
- **Policy:** api/plugin/xds structure? Similar policy? K8s+Universal? TargetRef complete? Inbound+outbound?
- **XDS:** Envoy config valid? Golden file? All protocols? Allocations? Caching?
- **KDS:** Consistency? Cross-zone isolation? Version compat? Sync failure? Resource isolation?
- **Errors:** Wrapped? No reflect.TypeOf? Rollback handled? Custom Is()? multierr?

**Skip (False Positives):**
- `_ = cleanup()` in defer • Generated code (zz_generated.*, *.pb.go) • Test builder panics • CI-covered (format, imports, lints) • Concrete types for performance

**Confidence Calibration:**
- **Lower:** No codebase check, performance intent, generated/test code, CI-handled, organization not correctness
- **Higher:** Violates rules (reflect.TypeOf, >50 lines), security, runtime failure, inconsistent patterns, breaks isolation

---

## Severity & Checklist

**Levels:**
- 🔴 **Block (95%+):** Security, breaking changes, data loss, incorrect xDS
- 🟡 **Change (80%+):** Missing tests, logic errors, performance, policy violations
- 🟢 **Comment (70%+):** Style, optimizations

**Security:** mTLS/certs secure • Multi-zone auth • RBAC • No secrets in logs • Input validation • Cross-zone isolation

**Correctness:** Valid xDS • Policy edge cases • KDS consistency • Resource versions • Context propagation • Rollback handling

**Go Standards:**
```go
// ✅ Wrap errors
return errors.Wrap(err, "context")

// ✅ Custom error Is() - NO reflect.TypeOf
func (e *NotFound) Is(err error) bool { _, ok := err.(*NotFound); return ok }

// ✅ Multi-error
return multierr.Append(rollbackErr, err)

// ❌ Functions >50 lines, raw errors, ignored rollbacks
```

**Policy:** api/plugin/xds structure • Validator complete • TargetRef kinds • Inbound+outbound

**XDS:** Efficient ResourceSet • Listener/cluster/route • Protocol handling • Metadata • Cache invalidation

**KDS:** Global↔Zone sync • Version compat • Consistency • Failure handling

**Tests:** Ginkgo structure (setup/given/when/then) • Table-driven validators • Golden files • K8s+Universal • E2E for user features

**Performance:** Minimize allocations • Efficient algorithms • Caching (MeshContext) • Optimized queries • Batching • No unnecessary goroutines

---

## Path Rules & CI Coverage

**pkg/plugins/policies/**: Validator complete • api/plugin/xds structure • TargetRef • Protocol handling
**pkg/xds/**: Efficient config • Caching • ResourceSet • Metadata • No hot-path allocations
**pkg/kds/**: Consistency • Version compat • Failure handling • Resource mapping
**test/**: K8s+Universal • Cleanup • Idempotent • Clear errors • Reasonable timeouts

**CI Handles (Skip):** Format (gofmt, gci) • Lints (golangci-lint) • Imports • Generated code • Commit format • License

---

## Anti-Patterns

❌ **Always Flag:** reflect.TypeOf • Ignored errors `_` • Panics in prod • Functions >50 lines • Missing test markers (setup/given/when/then)

---

## Quick Reference

**Check:** Security (mTLS, auth, secrets) • Correctness (xDS, policies, KDS) • Tests (unit+E2E, K8s+Universal, golden) • Performance (allocations, caching) • Go (wrapped errors, <50 lines, no reflect)

**Thresholds:** 95%+ Block • 80%+ Change • 70%+ Comment • <70% Skip
