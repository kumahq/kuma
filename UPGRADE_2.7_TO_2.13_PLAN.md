# Kuma 2.7 → 2.13 Upgrade Plan

## Executive Summary

**Verdict**: Upgrade is possible but requires significant preparation due to multiple breaking changes across 6 major versions.

**Risk Level**: Medium-High
- Resource name validation changes may break existing resources
- DNS system replacement
- Kubernetes minimum version bump
- Multiple policy API changes

**Recommended Approach**: Staged upgrade through 2.11 first (largest breaking change set)

---

## Version-by-Version Breaking Changes

### 2.7 → 2.8

| Change | Impact | Action Required |
|--------|--------|-----------------|
| `MeshRetry.MaxConnectAttempt=0` rejected | Validation | Fix any policies with zero value |
| `MeshFaultInjection.responseBandwidth.limit` validation | Validation | Ensure valid bandwidth format |
| Legacy tokens (pre-2.1.x) must be renewed | Auth | Renew old DP tokens |
| Envoy range: v1.28-v1.31 | Compatibility | Verify Envoy version |

### 2.8 → 2.9

| Change | Impact | Action Required |
|--------|--------|-----------------|
| Virtual Probes deprecated | Health checks | Plan migration to Application Probe Proxy |
| `--redirect-inbound-port-ipv6` removed | Transparent proxy | Use `--ip-family-mode` instead |
| `redirectPortInboundV6` field removed from Dataplane | Config | Remove from Dataplane specs |
| MeshExternalService Unix socket removed | Connectivity | Use TCP/HTTP endpoints |
| `--kuma-dp-uid` deprecated | Config | Use `--kuma-dp-user` |
| MeshGatewayInstance `kuma.io/service` generation forbidden | Gateway | Update gateway configs |

### 2.9 → 2.10 (HIGH IMPACT)

| Change | Impact | Action Required |
|--------|--------|-----------------|
| **RFC 1123/1035 resource name validation** | **CRITICAL** | Audit ALL resource names - underscores now invalid |
| `targetRef.kind: MeshSubset` → `Dataplane` | Policy targeting | Update all policies using MeshSubset |
| New `rules` API replaces `from` section | Policy format | Migrate affected policies |
| MeshAccessLog, MeshCircuitBreaker, MeshRateLimit, MeshTimeout, MeshTls affected | Policies | Review and update |
| Cannot mix inbound/outbound in same policy | Policy structure | Split policies if needed |
| Metric name format change | Monitoring | Update dashboards/alerts |
| Envoy v1.32-v1.33 | Compatibility | Major Envoy jump |

**Metric name change example**:
```
Old: inbound:127.0.0.1:21011.rbac.allowed
New: http.127.0.0.1:21011.rbac.allowed
```

### 2.10 → 2.11 (HIGH IMPACT)

| Change | Impact | Action Required |
|--------|--------|-----------------|
| **Min Kubernetes: 1.25.x → 1.27.x** | **CRITICAL** | Verify/upgrade K8s clusters |
| Max Kubernetes: 1.31.x → 1.32.x | Compatibility | Check K8s version |
| **Embedded DNS server replaces CoreDNS** | DNS | Test DNS resolution thoroughly |
| Namespace labels required for injection | Config | Add `kuma.io/sidecar-injection` labels |
| ClusterRole permissions split | RBAC | Review/update RBAC |
| MeshAccessLog TCP newlines removed | Logging | Update log parsers if needed |
| Envoy v1.34.1 | Compatibility | Major Envoy jump |
| Delta XDS default | Performance | Monitor XDS behavior |

**Disable embedded DNS if issues**:
```yaml
KUMA_DNS_PROXY_PORT: "0"
```

### 2.11 → 2.12

| Change | Impact | Action Required |
|--------|--------|-----------------|
| `/status/zones` endpoints removed | API | Update any integrations using this endpoint |
| Readiness reporter TCP port deprecated | Health | Switch to Unix socket |
| Envoy v1.33-v1.34.3 | Compatibility | Verify Envoy |

### 2.12 → 2.13

| Change | Impact | Action Required |
|--------|--------|-----------------|
| **Strict Inbound Port Filtering enabled by default** | Traffic | May block undeclared ports |
| **Virtual Probes disabled by default** | Health | Ensure Application Probe Proxy works |
| Go module path: `github.com/kumahq/kuma` → `/v2` | Library users | Update imports if using as library |
| ServiceAccountName removed | Auth | Use `AllowedUsers` config |
| MeshTrust `spec.origin` → `status.origin` | Config | Update MeshTrust resources |
| Max Kubernetes: 1.34.3 | Compatibility | Check K8s version |
| Envoy v1.36.4 | Compatibility | Major Envoy jump |

**Disable strict port filtering initially**:
```bash
KUMA_DATAPLANE_RUNTIME_STRICT_INBOUND_PORTS_ENABLED=false
```

---

## Kubernetes Version Matrix

| Kuma Version | Min K8s | Max K8s |
|--------------|---------|---------|
| 2.7.x | 1.25.x | 1.31.x |
| 2.8.x | 1.25.x | 1.31.x |
| 2.9.x | 1.25.x | 1.31.x |
| 2.10.x | 1.25.x | 1.31.x |
| 2.11.x | **1.27.x** | 1.32.x |
| 2.12.x | 1.27.x | 1.32.x |
| 2.13.x | 1.27.x | 1.34.3 |

**Action**: Kubernetes clusters must be 1.27.x+ before upgrading to Kuma 2.11+

---

## Envoy Version Matrix

| Kuma Version | Envoy Version |
|--------------|---------------|
| 2.7.x | v1.27 - v1.31 |
| 2.8.x | v1.28 - v1.31 |
| 2.9.x | v1.30 - v1.31 |
| 2.10.x | v1.32 - v1.33 |
| 2.11.x | v1.34.1 |
| 2.12.x | v1.33 - v1.34.3 |
| 2.13.x | v1.36.4 |

---

## Pre-Upgrade Checklist

### 1. Resource Name Audit (CRITICAL for 2.10+)

```bash
# Find resources with underscores (will break in 2.10+)
kubectl get meshes,dataplanes,meshtrafficpermissions -A -o name | grep '_'

# Check kuma.io/service tags for underscores
kubectl get pods -A -o jsonpath='{range .items[*]}{.metadata.annotations.kuma\.io/service}{"\n"}{end}' | grep '_'
```

**RFC 1123 Rules**:
- Lowercase alphanumeric
- Hyphens (`-`) and dots (`.`) allowed
- **Underscores (`_`) NOT allowed**
- Max 253 characters
- RFC 1035: Max 63 chars, must start alphabetic

### 2. Policy Audit

```bash
# Find policies using MeshSubset (deprecated in 2.10)
kubectl get meshtrafficpermissions,meshtimeouts,meshretries -A -o yaml | grep 'kind: MeshSubset'

# Find policies using 'from' section (changed in 2.10)
kubectl get meshaccesslogs,meshcircuitbreakers,meshratelimits,meshtimeouts,meshtls -A -o yaml | grep -A5 'from:'
```

### 3. Kubernetes Version Check

```bash
kubectl version --short
# Must be >= 1.27.x for Kuma 2.11+
```

### 4. Existing Configuration Backup

```bash
# Export all Kuma resources
kubectl get meshes,dataplanes,meshtrafficpermissions,meshtimeouts,meshretries,meshaccesslogs,meshcircuitbreakers -A -o yaml > kuma-backup-$(date +%Y%m%d).yaml

# Backup Helm values
helm get values kuma -n kuma-system > kuma-helm-values-backup.yaml
```

### 5. Review Custom Integrations

- [ ] Scripts calling `/status/zones` endpoint (removed in 2.12)
- [ ] Monitoring dashboards with old metric names (changed in 2.10)
- [ ] Custom health check integrations (Virtual Probes → Application Probe Proxy)
- [ ] Any code importing Kuma Go modules (path changed in 2.13)

---

## Recommended Upgrade Path

### Option A: Direct Jump (Not Recommended)

2.7 → 2.13 directly

**Risks**:
- Too many changes to debug issues
- Difficult to identify which version caused problems
- Rollback complexity

### Option B: Staged Upgrade (Recommended)

```
2.7 → 2.9 → 2.11 → 2.13
```

**Rationale**:
1. **2.7 → 2.9**: Incremental, prepare for probe changes
2. **2.9 → 2.11**: Biggest jump - DNS, K8s min, Envoy upgrade
3. **2.11 → 2.13**: Strict port filtering, stabilization

### Option C: One-at-a-Time (Most Conservative)

```
2.7 → 2.8 → 2.9 → 2.10 → 2.11 → 2.12 → 2.13
```

**Use if**: Complex environment, many custom policies, low risk tolerance

---

## Step-by-Step Upgrade Procedure

### Phase 1: Preparation (Before Any Upgrade)

1. **Backup everything**
   ```bash
   kubectl get crd -o name | xargs -I{} kubectl get {} -A -o yaml > full-backup.yaml
   ```

2. **Fix resource names with underscores**
   - Rename services/dataplanes to use hyphens
   - Update all references

3. **Update policies using MeshSubset**
   ```yaml
   # Old
   targetRef:
     kind: MeshSubset
     name: backend
     tags:
       version: v1

   # New
   targetRef:
     kind: Dataplane
     labels:
       app: backend
       version: v1
   ```

4. **Verify K8s version >= 1.27**

### Phase 2: Upgrade to 2.11 (Staging Environment)

1. **Deploy 2.11 to staging**
   ```bash
   helm upgrade kuma kuma/kuma --version 2.11.x -n kuma-system -f values.yaml
   ```

2. **Verify embedded DNS**
   ```bash
   # Test DNS resolution from mesh pods
   kubectl exec -it <pod> -- nslookup <service>.mesh
   ```

3. **If DNS issues, disable temporarily**
   ```yaml
   # values.yaml
   controlPlane:
     envVars:
       KUMA_DNS_PROXY_PORT: "0"
   ```

4. **Monitor for 24-48 hours**

### Phase 3: Upgrade to 2.13 (Staging Environment)

1. **Deploy 2.13 with strict port filtering disabled**
   ```yaml
   # values.yaml
   controlPlane:
     envVars:
       KUMA_DATAPLANE_RUNTIME_STRICT_INBOUND_PORTS_ENABLED: "false"
   ```

2. **Verify Application Probe Proxy**
   ```bash
   kubectl logs <pod> -c kuma-sidecar | grep "probe proxy"
   ```

3. **Test all services under load**

4. **Enable strict port filtering**
   - Remove the env var
   - Rolling restart dataplanes
   - Monitor for blocked traffic

### Phase 4: Production Rollout

1. **Follow same staging procedure**
2. **Use canary deployment if possible**
3. **Have rollback ready**
   ```bash
   helm rollback kuma -n kuma-system
   ```

---

## Rollback Plan

### Quick Rollback (Same Version Line)

```bash
helm rollback kuma <revision> -n kuma-system
```

### Major Version Rollback

**Warning**: May require resource recreation due to API changes

1. Scale down control plane
2. Restore CRDs from backup if needed
3. Reinstall previous version
4. Apply resource backup

---

## Post-Upgrade Validation

### Functional Tests

- [ ] Service-to-service communication works
- [ ] mTLS handshakes succeed
- [ ] DNS resolution works (`*.mesh` domains)
- [ ] Health checks pass (liveness/readiness probes)
- [ ] Policies applied correctly (check XDS)

### Monitoring Checks

- [ ] Update Grafana dashboards for new metric names
- [ ] Verify alerting rules still trigger
- [ ] Check control plane metrics
- [ ] Monitor dataplane memory/CPU

### Commands for Validation

```bash
# Check control plane health
kubectl get pods -n kuma-system
kubectl logs -n kuma-system deployment/kuma-control-plane

# Verify mesh connectivity
kumactl inspect dataplanes
kumactl inspect meshes

# Check policy application
kumactl inspect dataplane <name> --type=config

# Verify DNS
kubectl exec -it <mesh-pod> -- nslookup backend.mesh
```

---

## Known Issues & Workarounds

### DNS Resolution Failures (2.11+)

**Symptom**: Services can't resolve `*.mesh` domains
**Workaround**:
```yaml
KUMA_DNS_PROXY_PORT: "0"  # Falls back to CoreDNS
```

### Blocked Traffic (2.13)

**Symptom**: Services receiving connection refused on valid ports
**Workaround**:
```yaml
KUMA_DATAPLANE_RUNTIME_STRICT_INBOUND_PORTS_ENABLED: "false"
```

### Probe Failures (2.9+)

**Symptom**: Pods failing liveness/readiness after upgrade
**Cause**: Virtual Probes deprecated, Application Probe Proxy transition
**Fix**: Restart pods to pick up new probe configuration

---

## Timeline Estimate

| Phase | Duration |
|-------|----------|
| Preparation (resource audit, fixes) | 1-2 weeks |
| Staging 2.11 deployment + testing | 1 week |
| Staging 2.13 deployment + testing | 1 week |
| Production canary | 1 week |
| Full production rollout | 1-2 days |
| **Total** | **4-6 weeks** |

---

## References

- [UPGRADE.md](./UPGRADE.md) - Official upgrade guide
- [CHANGELOG.md](./CHANGELOG.md) - Full changelog
- [Kuma Documentation](https://kuma.io/docs/)
