# Kuma 3.0 Observability

TODO:
- unified naming
- remove kumactl install observability
- remove the possibility to configure metrics with pod annotations
- remove observability config from Mesh

## Current State

### OpenTelemetry Implementation

#### ✅ Done

- MeshTrace policy with OTel backend (gRPC/OTLP only)
- MeshMetric policy with OTel backend (experimental, gRPC only)
- MeshAccessLog policy with OTel backend (gRPC only)
- W3C TraceContext propagation in control plane
- Basic sampling configuration (overall, client, random)

#### ❌ Missing

- HTTP/HTTPS OTLP endpoints ([#9459](https://github.com/kumahq/kuma/issues/9459))
- TLS/authentication for OTLP

##### ❌ Missing - nice to have:

- Configurable resource attributes
- W3C Baggage propagation
- B3 propagator (Zipkin/Istio compatibility)
- Advanced sampling (parent-based, tail-based)

### Dashboard & Observability

#### ✅ Done

- 6 Grafana dashboards (mesh, CP, dataplane, service, service-to-service)
- `kumactl install observability` (Prometheus, Grafana, Loki, Jaeger)
- Envoy/Prometheus metrics collection

#### ❌ Missing

- OTel-native dashboards for MeshMetric backend
- Golden Signals (RED/USE) dashboards
- SLO/SLI tracking dashboards
- Security/policy compliance dashboards
- Multi-zone observability dashboards
- Distributed tracing integration in Grafana

---

## Gaps

### OTel Gaps

#### Backend Configuration Inconsistencies ([#8884](https://github.com/kumahq/kuma/issues/8884))

**Current:** Different validation/formats across policies
**Required:** Unified backend configuration schema
**Impact:**
- Poor UX, users confused by inconsistencies
- Hard to maintain/extend
- Documentation complexity

#### HTTP/HTTPS OTLP Protocol Support ([#9459](https://github.com/kumahq/kuma/issues/9459))

**Current:** Only gRPC (port 4317) supported
**Required:** HTTP (4318), HTTPS with TLS
**Impact:**
- Blocks integration with many OTel collectors requiring HTTP
- Cloud providers often prefer HTTP for cost/compatibility
- TLS requirement for secure transmission
**Affected Policies:** MeshTrace, MeshMetric, MeshAccessLog

#### MeshMetric OTel Backend Experimental Status ([#11870](https://github.com/kumahq/kuma/issues/11870))

**Current:** Beta/experimental flag
**Required:** GA/stable for 3.0
**Impact:**
- Users won't adopt experimental features in production
- Blocks OTel-first observability strategy
- Inconsistent with MeshTrace stability

#### TLS/Authentication for OTLP

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

#### OTel-Native Dashboards

**Missing:** Dashboards visualizing MeshMetric OTel backend data
**Impact:** Users can't monitor OTel metrics pipeline

#### Golden Signals

**Missing:** Explicit RED (Rate, Errors, Duration) dashboards
**Impact:** Industry-standard monitoring pattern unavailable

#### SLO/SLI Tracking

**Missing:** Service Level Objective/Indicator dashboards
**Impact:** No reliability tracking, SRE practices unsupported

#### Security/Policy

**Missing:** mTLS status, RBAC enforcement, policy violations
**Impact:** Security posture invisible to operators

#### Multi-Zone Observability

**Missing:** Cross-zone traffic, zone health, federation metrics
**Impact:** Multi-zone deployments hard to monitor

### Nice to Have

#### Context Propagation Incompleteness

**Current:** W3C TraceContext + Datadog only (hardcoded)
**Required:**
- W3C Baggage (standard companion to TraceContext)
- B3 propagator (Zipkin, Istio compatibility)
- Configurable propagator list
**Impact:**
- Trace context lost in heterogeneous environments
- Breaks distributed tracing with Istio/Zipkin
- No baggage for cross-service metadata

#### Custom Resource Attributes

**Current:** No configurable resource attributes
**Required:**
- ability for user to manually configure resource attributes
**Impact:**
- Poor correlation in OTel backends
- Manual attribute management burden

#### Advanced Sampling

**Current:** Basic random/probability sampling
**Required:**
- Parent-based sampling (honor upstream decisions)
- Tail-based sampling (sample after span completion)
- Adaptive sampling (dynamic rates)
**Impact:**
- Inefficient trace collection
- Missing critical traces or collecting too many
- No smart sampling for errors/high latency

---

## Open Questions for Discussion

### Scope

1. **All P0 for 3.0?** Or document known limitations and target 3.1?
2. **Dashboard format?** JSON only or Jsonnet for templating? ([#7167](https://github.com/kumahq/kuma/issues/7167))
3. **Prometheus 3 compatibility?** Ensure UTF-8 metric names ([#14426](https://github.com/kumahq/kuma/issues/14426))

### Integration

1. **Dashboard deployment?** Support ConfigMap installation for existing Grafana ([#10369](https://github.com/kumahq/kuma/issues/10369))
2. **PodMonitor support?** Define container ports for MeshMetric ([#13281](https://github.com/kumahq/kuma/issues/13281))

## References

### GitHub Issues

Umbrella: [#11693](https://github.com/kumahq/kuma/issues/11693) - Review observability setup

**Critical:**
- [#9459](https://github.com/kumahq/kuma/issues/9459) - HTTP/HTTPS OTLP endpoint support
- [#11870](https://github.com/kumahq/kuma/issues/11870) - MeshMetric OTel experimental status
- [#8884](https://github.com/kumahq/kuma/issues/8884) - Backend configuration inconsistencies
- [#8492](https://github.com/kumahq/kuma/issues/8492) - Attribute interpolation for logging
- [#8348](https://github.com/kumahq/kuma/issues/8348) - Timeout/retry configuration

**Backend/Configuration:**
- [#10158](https://github.com/kumahq/kuma/issues/10158) - Unbounded metric cardinality with socket for OTLP metrics

**Dashboards:**
- [#7167](https://github.com/kumahq/kuma/issues/7167) - Grafana dashboards as code
- [#10369](https://github.com/kumahq/kuma/issues/10369) - Install Grafana dashboards as config maps

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

### OTel Resources

- OpenTelemetry Specification 1.52.0+
- Semantic Conventions for Service Mesh
- OTLP Protocol Specification
- Context Propagation Standards (W3C TraceContext, Baggage, B3)
