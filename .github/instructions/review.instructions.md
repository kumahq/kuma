---
applyTo: "**"
excludeAgent: "coding-agent"
---

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

## Blast Radius (Downstream Impact)

**When the diff stops producing, renames, moves, or gates data, trace who still reads it.** Most Kuma state crosses untyped boundaries - `map[string]string` tags/labels, string-keyed xDS metadata, KDS-synced proto fields - so the compiler and unit tests do NOT catch a broken producer→consumer link. Verify it by hand.

**Triggers (any in the diff):**

- Removes/renames a struct field, tag key, label, or xDS metadata key (`envoy.lb`, `kuma.io/*`, `io.kuma.*`)
- Gates an existing data path behind a flag/condition (`KUMA_EXPERIMENTAL_*`, `if ...Disabled`)
- Stops writing something that was previously always set
- Changes what KDS syncs global↔zone, or a MeshContext cache-hash input

**Procedure:**

1. Name the datum (field / tag key / metadata key / label / flag).
2. Grep every reader - including string-literal lookups (`tags["kuma.io/zone"]`, `metadata["envoy.lb"]`, `GetLabels()[...]`), not just typed references. Cross package boundaries.
3. For each reader: is there a fallback when the datum is absent? If not → likely runtime break with no failing test.
4. Report the chain producer→…→consumer, not just the edit site. Ask for a test covering the affected flag/path combination.

**Kuma hot paths to trace:**

- Inbound tags ↔ Dataplane labels ↔ endpoint metadata ↔ LB/affinity (MeshLoadBalancingStrategy resolves locality by tag/label key)
- TargetRef / policy matching resolves subsets by tag key
- KDS field drop/rename → breaks cross-version zone sync
- MeshContext cache-hash inputs → dropping one yields stale xDS

**Canonical miss:** allow-listed node labels lived only on inbound tags; enabling `KUMA_EXPERIMENTAL_INBOUND_TAGS_DISABLED` emptied the tag bag; MeshLoadBalancingStrategy affinity (a string-key reader two hops downstream) silently lost locality - no compile error, no failing test. This is the class to catch.

**Severity:** 🔴 consumer breaks with no fallback • 🟡 plausible break unverified, or no test for the flag/path combination.

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
