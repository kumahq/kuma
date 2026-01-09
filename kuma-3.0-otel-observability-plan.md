# Kuma 3.0 OpenTelemetry & Observability Completeness Plan

## Executive Summary

Kuma 3.0 needs comprehensive OpenTelemetry support and improved observability dashboards. Current implementation has critical gaps in protocol support, context propagation, and user-facing visualization tools.

---

## 1. Current State Assessment

### OpenTelemetry Implementation Status

**✅ Implemented:**
- MeshTrace policy with OTel backend (gRPC/OTLP only)
- MeshMetric policy with OTel backend (experimental, gRPC only)
- MeshAccessLog policy with OTel backend (gRPC only)
- W3C TraceContext propagation in control plane
- Basic sampling configuration (overall, client, random)

**❌ Missing:**
- HTTP/HTTPS OTLP endpoints ([#9459](https://github.com/kumahq/kuma/issues/9459))
- W3C Baggage propagation
- B3 propagator (Zipkin/Istio compatibility)
- Configurable resource attributes
- Semantic conventions compliance
- TLS/authentication for OTLP
- Advanced sampling (parent-based, tail-based)
- Profiles signal support
- OTel Collector integration patterns

### Dashboard & Observability Status

**✅ Existing:**
- 6 Grafana dashboards (mesh, CP, dataplane, service, service-to-service, gateway)
- `kumactl install observability` (Prometheus, Grafana, Loki, Jaeger)
- Envoy/Prometheus metrics collection

**❌ Missing:**
- OTel-native dashboards for MeshMetric backend
- Golden Signals (RED/USE) dashboards
- SLO/SLI tracking dashboards
- Security/policy compliance dashboards
- Multi-zone observability dashboards
- Distributed tracing integration in Grafana
- Application-level metrics dashboards

---

## 2. Gap Analysis & Impact

### Critical OTel Gaps

#### Gap 1: HTTP/HTTPS OTLP Protocol Support ([#9459](https://github.com/kumahq/kuma/issues/9459))
**Current:** Only gRPC (port 4317) supported
**Required:** HTTP (4318), HTTPS with TLS
**Impact:**
- Blocks integration with many OTel collectors requiring HTTP
- Cloud providers often prefer HTTP for cost/compatibility
- TLS requirement for secure transmission
**Affected Policies:** MeshTrace, MeshMetric, MeshAccessLog

#### Gap 2: MeshMetric OTel Backend Experimental Status ([#11870](https://github.com/kumahq/kuma/issues/11870))
**Current:** Beta/experimental flag
**Required:** GA/stable for 3.0
**Impact:**
- Users won't adopt experimental features in production
- Blocks OTel-first observability strategy
- Inconsistent with MeshTrace stability

#### Gap 3: Backend Configuration Inconsistencies ([#8884](https://github.com/kumahq/kuma/issues/8884), [#5868](https://github.com/kumahq/kuma/issues/5868))
**Current:** Different validation/formats across policies
**Required:** Unified backend configuration schema
**Impact:**
- Poor UX, users confused by inconsistencies
- Hard to maintain/extend
- Documentation complexity
**Related:** [#5868](https://github.com/kumahq/kuma/issues/5868) proposes Backends resource for grouping observability metadata

#### Gap 4: Context Propagation Incompleteness
**Current:** W3C TraceContext + Datadog only (hardcoded)
**Required:**
- W3C Baggage (standard companion to TraceContext)
- B3 propagator (Zipkin, Istio compatibility)
- Configurable propagator list
**Impact:**
- Trace context lost in heterogeneous environments
- Breaks distributed tracing with Istio/Zipkin
- No baggage for cross-service metadata

#### Gap 5: Resource Attributes & Semantic Conventions
**Current:** No configurable resource attributes
**Required:**
- service.name, service.version, service.namespace
- deployment.environment, k8s.* attributes
- Auto-detection from environment
**Impact:**
- Non-compliant with OTel semantic conventions
- Poor correlation in OTel backends
- Manual attribute management burden

#### Gap 6: Advanced Sampling
**Current:** Basic random/probability sampling
**Required:**
- Parent-based sampling (honor upstream decisions)
- Tail-based sampling (sample after span completion)
- Adaptive sampling (dynamic rates)
**Impact:**
- Inefficient trace collection
- Missing critical traces or collecting too many
- No smart sampling for errors/high latency

#### Gap 7: TLS/Authentication for OTLP
**Current:** No secure connection support
**Required:**
- TLS configuration (certs, CA, verification)
- Bearer token auth
- API key headers
**Impact:**
- Cannot connect to secured OTel endpoints
- Security compliance issues
- Blocks cloud OTel services

### Dashboard Gaps

#### Gap 8: OTel-Native Dashboards
**Missing:** Dashboards visualizing MeshMetric OTel backend data
**Impact:** Users can't monitor OTel metrics pipeline

#### Gap 9: Golden Signals Dashboards
**Missing:** Explicit RED (Rate, Errors, Duration) dashboards
**Impact:** Industry-standard monitoring pattern unavailable

#### Gap 10: SLO/SLI Tracking
**Missing:** Service Level Objective/Indicator dashboards
**Impact:** No reliability tracking, SRE practices unsupported

#### Gap 11: Security/Policy Dashboards
**Missing:** mTLS status, RBAC enforcement, policy violations
**Impact:** Security posture invisible to operators

#### Gap 12: Multi-Zone Observability
**Missing:** Cross-zone traffic, zone health, federation metrics
**Impact:** Multi-zone deployments hard to monitor

#### Gap 13: Distributed Tracing in Grafana
**Missing:** Tempo/Jaeger integration in Grafana dashboards
**Impact:** Separate tools needed, poor UX

---

## 3. Proposed Solutions

### Phase 1: Critical OTel Compliance (P0 for 3.0)

#### 1.1 HTTP/HTTPS OTLP Support
**Implementation:**
- Add `protocol` field to backend config (grpc|http|https)
- Implement HTTP OTLP client using `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp`
- Add TLS configuration struct (certFile, keyFile, caFile, insecureSkipVerify)
- Update validation for all three policies
- Add e2e tests with HTTP OTel Collector

**Files Affected:**
- `pkg/plugins/policies/meshtrace/api/v1alpha1/validator.go`
- `pkg/plugins/policies/meshmetric/api/v1alpha1/validator.go`
- `pkg/plugins/policies/meshaccesslog/api/v1alpha1/validator.go`
- New: `pkg/observability/otlp/http_client.go`

**Configuration Example:**
```yaml
backend:
  openTelemetry:
    endpoint: https://otel-collector:4318
    protocol: https
    tls:
      mode: provided
      ca: /path/to/ca.crt
      cert: /path/to/client.crt
      key: /path/to/client.key
```

#### 1.2 Unified Backend Configuration
**Implementation:**
- Create shared `BackendConfig` proto in `api/common/v1alpha1/`
- Migrate all policies to use common config
- Unified validation package
- Single source of truth for OTLP backends

**Files Affected:**
- New: `api/common/v1alpha1/backend.proto`
- Update all policy protos to reference common backend

#### 1.3 Graduate MeshMetric OTel to Stable
**Implementation:**
- Remove experimental flags/warnings
- Complete test coverage (unit, integration, e2e)
- Performance benchmarks
- Update documentation to GA status
- Add to compatibility matrix
- Fix unbounded metric cardinality ([#10158](https://github.com/kumahq/kuma/issues/10158))

**Acceptance Criteria:**
- 90%+ test coverage
- Performance: <5% overhead at 10k rps
- Validated with 3+ OTel backends (Collector, Datadog, Grafana Cloud)
- Bounded metric cardinality

#### 1.4 W3C Baggage Propagation
**Implementation:**
- Add W3C Baggage propagator to control plane
- Enable by default alongside TraceContext
- Support baggage in MeshTrace custom tags
- Add tests for baggage propagation

**Files Affected:**
- `pkg/xds/generator/tracing_generator.go`
- `app/kuma-cp/pkg/config/app/kuma-cp.defaults.yaml`

#### 1.5 Configurable Resource Attributes
**Implementation:**
- Add `resourceAttributes` map to MeshMetric/MeshTrace
- Auto-detect from k8s labels/annotations
- Support interpolation from dataplane metadata
- Apply semantic conventions defaults

**Configuration Example:**
```yaml
spec:
  default:
    backends:
      - openTelemetry:
          endpoint: collector:4317
          resourceAttributes:
            service.name: "{{kuma.io/service}}"
            service.version: "{{version}}"
            deployment.environment: "{{env}}"
            k8s.namespace.name: "{{k8s.namespace}}"
```

### Phase 2: Enhanced Compatibility (P1)

#### 2.1 B3 Propagator Support
**Implementation:**
- Add B3 Single/Multi propagators
- Configurable propagator list in MeshTrace
- Support propagator priority

**Configuration:**
```yaml
spec:
  default:
    propagators:
      - w3c_tracecontext
      - w3c_baggage
      - b3
      - datadog
```

#### 2.2 Parent-Based Sampling
**Implementation:**
- Add parent-based sampler wrapper
- Configure fallback sampler when no parent
- Respect parent sampling decisions

#### 2.3 OTLP Authentication
**Implementation:**
- Add `headers` map to backend config
- Support for Authorization, API-Key headers
- Environment variable substitution for secrets

**Configuration:**
```yaml
backend:
  openTelemetry:
    endpoint: https://api.honeycomb.io
    headers:
      X-Honeycomb-Team: "${HONEYCOMB_API_KEY}"
      X-Honeycomb-Dataset: "kuma-prod"
```

#### 2.4 Timeout/Retry Configuration ([#8348](https://github.com/kumahq/kuma/issues/8348))
**Implementation:**
- Add `timeout`, `retryPolicy` to backend config
- Configure max retries, backoff
- Circuit breaker for failing backends

### Phase 3: Advanced Observability (P2)

#### 3.1 Profiles Signal Support
**Implementation:**
- Add MeshProfile policy (when OTel profiles stabilize)
- Continuous profiling integration
- Pyroscope/Phlare compatibility

#### 3.2 Tail-Based Sampling
**Implementation:**
- Requires OTel Collector as proxy
- Add collector sidecar pattern docs/examples
- Configure tail sampling in collector

#### 3.3 OTel Collector Patterns
**Deliverables:**
- Sidecar collector deployment manifests
- Gateway collector configuration
- Multi-cluster collector federation
- Example: processors, batch export, filtering

---

## 4. Dashboard Implementation Plan

### 4.1 OTel Metrics Dashboard
**Metrics Visualized:**
- OTLP export success/failure rates
- Backend latency/errors
- Dropped/sampled metric counts
- Resource attribute cardinality

**Panels:**
- Export success rate by backend
- Backend latency p50/p95/p99
- Metric/trace/log drop rate
- Active exporters status

### 4.2 Golden Signals Dashboard
**RED Method per Service:**
- Rate: requests/sec by service
- Errors: error rate % by service
- Duration: latency p50/p95/p99 by service
- Retry metrics ([#5052](https://github.com/kumahq/kuma/issues/5052))

**USE Method for Proxies:**
- Utilization: CPU/memory by dataplane
- Saturation: queue depths, conn limits
- Errors: proxy errors, upstream failures

### 4.3 SLO/SLI Dashboard
**Features:**
- Define SLOs (e.g., 99.9% availability, p95 < 200ms)
- Error budget tracking
- Burn rate alerts
- SLI trend graphs

### 4.4 Security/Policy Dashboard
**Visualizations:**
- mTLS enabled services %
- RBAC policy violations
- Traffic permissions denials
- Certificate expiry warnings ([#14949](https://github.com/kumahq/kuma/issues/14949))
- Policy compliance score

### 4.5 Multi-Zone Dashboard
**Metrics:**
- Cross-zone traffic volume
- Zone CP connectivity status
- Zone sync latency
- Zone ingress/egress health
- Global/zone resource counts

### 4.6 Distributed Tracing Integration
**Implementation:**
- Add Tempo data source to Grafana
- Link trace IDs in dashboards
- Service map from traces
- Trace exemplars in metrics

### 4.7 Cost/Resource Dashboard
**Visualizations:**
- Resource utilization by service
- Proxy overhead metrics
- Network bandwidth by zone
- Capacity planning projections

---

## 5. Implementation Considerations

### Backwards Compatibility
- Existing gRPC-only configs continue working
- HTTP/HTTPS opt-in via protocol field
- Old propagator behavior preserved unless configured
- Dashboard updates non-breaking

### Performance Impact
- HTTP OTLP: similar overhead to gRPC
- Resource attributes: <1% CPU increase
- Baggage propagation: <0.5% overhead
- Consider migrating from summary to histogram metrics ([#13646](https://github.com/kumahq/kuma/issues/13646))
- All changes benchmarked before merge

### Migration Strategy
- Document OTel best practices
- Provide migration guide from Prometheus to OTel
- Example configurations for popular backends
- Automated config converter tool (stretch)

### Documentation Requirements
- OTel policy reference updates
- Protocol selection guide
- Semantic conventions guide
- Dashboard import instructions
- Troubleshooting guide
- Backend-specific guides (Datadog, Grafana Cloud, Honeycomb, etc.)

### Testing Strategy
- Unit tests for all new validators/generators
- Integration tests with real OTel Collector
- E2e tests for HTTP/gRPC protocols
- Performance benchmarks
- Conformance tests against OTel spec
- Dashboard validation (query correctness)

---

## 6. Deliverables Summary

### Code Changes
1. HTTP/HTTPS OTLP client implementation
2. Unified backend configuration
3. W3C Baggage propagator
4. B3 propagator
5. Resource attributes configuration
6. Parent-based sampling
7. TLS/auth for OTLP
8. Timeout/retry logic

### Policy Updates
- MeshTrace: protocol, tls, auth, propagators, resourceAttributes
- MeshMetric: protocol, tls, auth, resourceAttributes (+ GA status)
- MeshAccessLog: protocol, tls, auth, resourceAttributes

### Dashboards (7 new + improvements)
1. `kuma-otel-metrics.json` - OTel pipeline monitoring
2. `kuma-golden-signals.json` - RED/USE metrics
3. `kuma-slo.json` - SLO/SLI tracking
4. `kuma-security.json` - Security/policy compliance
5. `kuma-multizone.json` - Multi-zone observability (update KDS metrics [#13569](https://github.com/kumahq/kuma/issues/13569))
6. `kuma-tracing.json` - Distributed tracing integration
7. `kuma-resources.json` - Cost/capacity planning

**Existing Dashboard Improvements:**
- Gateway: Add paths/listeners visibility ([#4656](https://github.com/kumahq/kuma/issues/4656))
- Control Plane: Improve XDS metrics protocol distinction ([#13462](https://github.com/kumahq/kuma/issues/13462))
- All: Add retry panels ([#5052](https://github.com/kumahq/kuma/issues/5052))

### Documentation
- OTel configuration guide
- Dashboard import guide
- Migration guide (Prometheus → OTel)
- Backend integration guides
- Troubleshooting runbook
- Semantic conventions reference

### Examples
- HTTP OTLP configuration
- Datadog OTel integration
- Grafana Cloud integration
- Honeycomb integration
- OTel Collector sidecar pattern
- OTel Collector gateway pattern

---

## 7. Open Questions for Team Discussion

### Scope
1. **All P0 for 3.0?** Or document known limitations and target 3.1?
2. **Dashboard format?** JSON only or Jsonnet for templating? ([#7167](https://github.com/kumahq/kuma/issues/7167))
3. **MeshMetric GA criteria?** What testing/validation needed?
4. **Prometheus 3 compatibility?** Ensure UTF-8 metric names ([#14426](https://github.com/kumahq/kuma/issues/14426))

### Design
5. **Propagator configuration location?** MeshTrace policy or global config?
6. **Resource attributes precedence?** Policy vs auto-detected vs dataplane metadata?
7. **Backend config migration?** Breaking change or dual support?

### Testing
8. **OTel conformance?** Use official OTel test suite?
9. **Dashboard testing?** Automated validation of queries?
10. **Performance targets?** Acceptable overhead thresholds?

### Integration
11. **Ship OTel Collector?** Bundle or document external deployment?
12. **Default backend?** Should Kuma ship with OTel Collector by default?
13. **Observability installer?** Update `kumactl install observability` for OTel-first?
14. **Dashboard deployment?** Support ConfigMap installation for existing Grafana ([#10369](https://github.com/kumahq/kuma/issues/10369))
15. **PodMonitor support?** Define container ports for MeshMetric ([#13281](https://github.com/kumahq/kuma/issues/13281))

### Migration
16. **Dual metrics?** Support Prometheus + OTel simultaneously?
17. **Dashboard versioning?** How to handle dashboard updates?
18. **Config converter?** Tool to migrate old configs to new format?

---

## 8. Success Criteria

### OTel Compliance
- ✅ Support HTTP + gRPC OTLP protocols
- ✅ W3C TraceContext + Baggage propagation
- ✅ B3 propagator for ecosystem compatibility
- ✅ Semantic conventions compliance (resource attributes)
- ✅ Secure connections (TLS/auth)
- ✅ All signals GA (traces, metrics, logs)

### User Experience
- ✅ Users can import 7+ dashboards for complete observability
- ✅ Works with 5+ popular OTel backends out-of-box
- ✅ Migration guide from old observability
- ✅ Clear documentation for each backend
- ✅ <30min to full observability stack

### Performance
- ✅ <5% overhead with OTel enabled
- ✅ <2% overhead for propagation features
- ✅ Handles 10k+ rps per dataplane
- ✅ No memory leaks in long-running tests

### Compatibility
- ✅ Backwards compatible with existing configs
- ✅ Works with Istio (B3 propagation)
- ✅ Works with Zipkin (B3 propagation)
- ✅ Works with all major OTel backends

---

## References

### GitHub Issues - OTel/Observability

**Critical:**
- [#9459](https://github.com/kumahq/kuma/issues/9459) - HTTP/HTTPS OTLP endpoint support
- [#11870](https://github.com/kumahq/kuma/issues/11870) - MeshMetric OTel experimental status
- [#8884](https://github.com/kumahq/kuma/issues/8884) - Backend configuration inconsistencies
- [#8492](https://github.com/kumahq/kuma/issues/8492) - Attribute interpolation for logging
- [#8348](https://github.com/kumahq/kuma/issues/8348) - Timeout/retry configuration

**Backend/Configuration:**
- [#5868](https://github.com/kumahq/kuma/issues/5868) - Add Backends resource for grouping observability metadata
- [#11693](https://github.com/kumahq/kuma/issues/11693) - Review observability setup
- [#10158](https://github.com/kumahq/kuma/issues/10158) - Unbounded metric cardinality with socket for OTLP metrics

**Dashboards:**
- [#7167](https://github.com/kumahq/kuma/issues/7167) - Grafana dashboards as code
- [#5052](https://github.com/kumahq/kuma/issues/5052) - Grafana panels for retries
- [#10369](https://github.com/kumahq/kuma/issues/10369) - Install grafana dashboards as config maps
- [#4656](https://github.com/kumahq/kuma/issues/4656) - Gateway grafana dashboards improvements

**Metrics/Monitoring:**
- [#14426](https://github.com/kumahq/kuma/issues/14426) - Prometheus 3 UTF-8 compatible metrics
- [#13281](https://github.com/kumahq/kuma/issues/13281) - Define container ports for MeshMetric (PodMonitor support)
- [#14949](https://github.com/kumahq/kuma/issues/14949) - Expose metric for mTLS cert expiry
- [#13569](https://github.com/kumahq/kuma/issues/13569) - Update KDS metrics
- [#13462](https://github.com/kumahq/kuma/issues/13462) - Improve XDS metrics to distinguish protocol versions
- [#13646](https://github.com/kumahq/kuma/issues/13646) - Consider using histogram to replace summary type metrics in CP

### Existing Dashboards
- Location: `app/kumactl/data/install/k8s/metrics/grafana/`
- Files: kuma-mesh.json, kuma-cp.json, kuma-dataplane.json, kuma-service.json, kuma-service-to-service.json, kuma-gateway.json
- Note: Gateway dashboard needs improvements for paths/listeners ([#4656](https://github.com/kumahq/kuma/issues/4656))

### OTel Resources
- OpenTelemetry Specification 1.52.0+
- Semantic Conventions for Service Mesh
- OTLP Protocol Specification
- Context Propagation Standards (W3C TraceContext, Baggage, B3)

---

**Next Steps:** Team review, prioritization, timeline assignment, resource allocation
