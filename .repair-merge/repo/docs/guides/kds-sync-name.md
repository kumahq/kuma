# What is the name of the resource after KDS sync?

## Old policies (i.e. TrafficRoute, TrafficPermission)

* CRD scope: Cluster
* Kuma scope: Mesh

### Global(k8s/universal) -> Zone(k8s/universal)

```
nameInZone := "${truncate(nameInGlobal, 236)}-${hash(mesh, nameInGlobal)}"
```

Example:

User creates a policy on Global:
```yaml
type: TrafficPermission
name: my-policy
```

* Name in the Kuma API on Global CP: `my-policy`
* Name in the Kuma API on Zone CP: `my-policy-fvdw2c2wx44fv4xz`

### Zone(k8s/universal) -> Global(k8s/universal)

Not supported

## New policies (i.e. MeshTrafficPermission, MeshHTTPRoute)

* CRD scope: Namespace
* Kuma scope: Mesh

### Global(k8s) -> Zone(k8s)

```
nameInGlobal = trimNsSuffix(nameInGlobal)
nameInZone := "${truncate(nameInGlobal, 236)}-${hash(mesh, nameInGlobal)}.${systemNamespace}"
```

Example: 

User creates a CRD on Global:
```yaml
kind: MeshTrafficPermission
metadata:
  name: my-policy
  namespace: kuma-system-global
```

* Name in the Kuma API on Global CP: `my-policy.kuma-system-global`
* Name of the CRD on Zone CP: `my-policy-fvdw2c2wx44fv4xz`
* Name in the Kuma API on Zone CP: `my-policy-fvdw2c2wx44fv4xz.kuma-system-zone`

### Global(universal) -> Zone(k8s)

```
nameInZone := "${truncate(nameInGlobal, 236)}-${hash(mesh, nameInGlobal)}.${systemNamespace}"
```

Example:

User creates a policy on Global:
```yaml
type: MeshTrafficPermission
name: my-policy
```

* Name in the Kuma API on Global CP: `my-policy`
* Name of the CRD on Zone CP: `my-policy-fvdw2c2wx44fv4xz`
* Name in the Kuma API on Zone CP: `my-policy-fvdw2c2wx44fv4xz.kuma-system-zone`
* 
### Global(k8s) -> Zone(universal)

```
nameInGlobal = trimNsSuffix(nameInGlobal)
nameInZone := "${truncate(nameInGlobal, 236)}-${hash(mesh, nameInGlobal)}"
```

Example:

User creates a CRD on Global:
```yaml
kind: MeshTrafficPermission
metadata:
  name: my-policy
  namespace: kuma-system-global
```

* Name in the Kuma API on Global CP: `my-policy.kuma-system-global`
* Name in the Kuma API on Zone CP: `my-policy-fvdw2c2wx44fv4xz`

### Global(universal) -> Zone(universal)

```
nameInZone := "${truncate(nameInGlobal, 236)}-${hash(mesh, nameInGlobal)}"
```

Example:

User creates a policy on Global:
```yaml
type: MeshTrafficPermission
name: my-policy
```

* Name in the Kuma API on Global CP: `my-policy`
* Name in the Kuma API on Zone CP: `my-policy-fvdw2c2wx44fv4xz`

### Zone(k8s/universal) -> Global(k8s/universal)

Not supported

## DPP

* CRD scope: Namespace
* Kuma scope: Mesh

### Global(k8s/universal) -> Zone(k8s/universal)

Not supported

### Zone(k8s) -> Global(k8s)

```
nameInGlobal := "${zone}.${nameInZone}.${nsInZone}.${systemNamespce}"
```

Example:

User starts a DPP in Zone `zone-1`:
```yaml
kind: Dataplane
metadata:
  name: my-dpp
  namespace: kuma-demo
```

* Name in the Kuma API on Zone CP: `my-dpp.kuma-demo`
* Name of the CRD on Global CP: `zone-1.my-dpp.kuma-demo` 
* Name in the Kuma API on Global CP: `zone-1.my-dpp.kuma-demo.kuma-system`

Notes: 
* in KDS v1 instead of `.kuma-system` we have `.default`
* DPP name collision is possible https://github.com/kumahq/kuma/issues/8249

### Zone(universal) -> Global(k8s)

```
nameInGlobal := "${zone}.${nameInZone}.${systemNamespce}"
```

Example:

User starts a DPP in Zone `zone-1`:
```yaml
type: Dataplane
name: my-dpp
```

* Name in the Kuma API on Zone CP: `my-dpp`
* Name of the CRD on Global CP: `zone-1.my-dpp`
* Name in the Kuma API on Global CP: `zone-1.my-dpp.kuma-system`

Notes:
* in KDS v1 instead of `.kuma-system` we have `.default`
* DPP name collision is possible https://github.com/kumahq/kuma/issues/8249

### Zone(universal) -> Global(universal)

```
nameInGlobal := "${zone}.${nameInZone}"
```

Example:

User starts a DPP in Zone `zone-1`:
```yaml
type: Dataplane
name: my-dpp
```

* Name in the Kuma API on Zone CP: `my-dpp`
* Name in the Kuma API on Global CP: `zone-1.my-dpp`

## ZoneIngress and ZoneEgress

* CRD scope: Namespace
* Kuma scope: Global

### Zone(k8s) -> Global(k8s)

```
nameInGlobal := "${zone}.${nameInZone}.${nsInZone}.${systemNamespace}" 
```

Example:

User starts a ZoneIngress in Zone `zone-1`:
```yaml
kind: ZoneIngress
metadata:
  name: my-ingress
  namespace: kuma-system-zone
```

* Name in the Kuma API on Zone CP: `my-ingress.kuma-system-zone`
* Name of the CRD on Global CP: `zone-1.my-ingress.kuma-system-zone` 
* Name in the Kuma API on Global CP: `zone-1.my-ingress.kuma-system-zone.kuma-system-global`

Notes:
* in KDS v1 instead of `.kuma-system-global` we have `.default`

### Zone(universal) -> Global(k8s)

```
nameInGlobal := "${zone}.${nameInZone}.${systemNamespace}" 
```

Example: 

User starts a ZoneIngress in Zone `zone-1`:
```yaml
type: ZoneIngress
name: my-ingress
```

* Name in the Kuma API on Zone CP: `my-ingress`
* Name of the CRD on Global CP: `zone-1.my-ingress`
* Name in the Kuma API on Global CP: `zone-1.my-ingress.kuma-system-global`

Notes:
* in KDS v1 instead of `.kuma-system-global` we have `.default`

### Zone(k8s) -> Global(universal)

```
nameInGlobal := "${zone}.${nameInZone}.${nsInZone}" 
```

Example:

User starts a ZoneIngress in Zone `zone-1`:
```yaml
kind: ZoneIngress
metadata:
  name: my-ingress
  namespace: kuma-system-zone
```

* Name in the Kuma API on Zone CP: `my-ingress.kuma-system-zone`
* Name in the Kuma API on Global CP: `zone-1.my-ingress.kuma-system-zone`

### Zone(universal) -> Global(universal)

```
nameInGlobal := "${zone}.${nameInZone}" 
```

Example:

User starts a ZoneIngress in Zone `zone-1`:
```yaml
type: ZoneIngress
name: my-ingress
```

* Name in the Kuma API on Zone CP: `my-ingress`
* Name in the Kuma API on Global CP: `zone-1.my-ingress`

### Global(k8s) -> Zone(k8s)

```
nameInZone := "${nameInGlobal}.${systemNamespace}"
```

Example 1: Zone(k8s) -> Global(k8s) -> Zone(k8s)

User starts a ZoneIngress in Zone `zone-1`:
```yaml
kind: ZoneIngress
metadata:
  name: my-ingress
  namespace: kuma-system-zone-1
```

* Name in the Kuma API on Zone CP of `zone-1`: `my-ingress.kuma-system-zone-1`
* Name of the CRD on Global CP: `zone-1.my-ingress.kuma-system-zone-1`
* Name in the Kuma API on Global CP: `zone-1.my-ingress.kuma-system-zone-1.kuma-system-global`
* Name of the CRD on Zone CP of `zone-2`: `zone-1.my-ingress.kuma-system-zone-1.kuma-system-global`
* Name in the Kuma API on Zone CP of `zone-2`: `zone-1.my-ingress.kuma-system-zone-1.kuma-system-global.kuma-system-zone-2`

Example 2: Zone(universal) -> Global(k8s) -> Zone(k8s)

User starts a ZoneIngress in Zone `zone-1`:
```yaml
type: ZoneIngress
name: my-ingress
```

* Name in the Kuma API on Zone CP of `zone-1`: `my-ingress`
* Name of the CRD on Global CP: `zone-1.my-ingress`
* Name in the Kuma API on Global CP: `zone-1.my-ingress.kuma-system-global`
* Name of the CRD on Zone CP of `zone-2`: `zone-1.my-ingress.kuma-system-global`
* Name in the Kuma API on Zone CP of `zone-2`: `zone-1.my-ingress.kuma-system-global.kuma-system-zone-2`

Notes:
* in KDS v1 instead of `.kuma-system-global` and `.kuma-system-zone-2` we have `.default`

### Global(universal) -> Zone(k8s)

```
nameInZone := "${nameInGlobal}.${systemNamespace}"
```

Example 1: Zone(k8s) -> Global(universal) -> Zone(k8s)

User starts a ZoneIngress in Zone `zone-1`:
```yaml
kind: ZoneIngress
metadata:
  name: my-ingress
  namespace: kuma-system-zone-1
```

* Name in the Kuma API on Zone CP of `zone-1`: `my-ingress.kuma-system-zone-1`
* Name in the Kuma API on Global CP: `zone-1.my-ingress.kuma-system-zone-1`
* Name of the CRD on Zone CP of `zone-2`: `zone-1.my-ingress.kuma-system-zone-1`
* Name in the Kuma API on Zone CP of `zone-2`: `zone-1.my-ingress.kuma-system-zone-1.kuma-system-zone-2`

Example 2: Zone(universal) -> Global(universal) -> Zone(k8s)

User starts a ZoneIngress in Zone `zone-1`:
```yaml
type: ZoneIngress
name: my-ingress
```

* Name in the Kuma API on Zone CP of `zone-1`: `my-ingress`
* Name in the Kuma API on Global CP: `zone-1.my-ingress`
* Name of the CRD on Zone CP of `zone-2`: `zone-1.my-ingress`
* Name in the Kuma API on Zone CP of `zone-2`: `zone-1.my-ingress.kuma-system-zone-2`

Notes:
* in KDS v1 instead of `.kuma-system-zone-2` we have `.default`

### Global(k8s) -> Zone(universal)

```
nameInZone := "${nameInGlobal}"
```

Example 1: Zone(k8s) -> Global(k8s) -> Zone(universal)

User starts a ZoneIngress in Zone `zone-1`:
```yaml
kind: ZoneIngress
metadata:
  name: my-ingress
  namespace: kuma-system-zone-1
```

* Name in the Kuma API on Zone CP of `zone-1`: `my-ingress.kuma-system-zone-1`
* Name of hte CRD on Global CP: `zone-1.my-ingress.kuma-system-zone-1`
* Name in the Kuma API on Global CP: `zone-1.my-ingress.kuma-system-zone-1.kuma-system-global`
* Name in the Kuma API on Zone CP of `zone-2`: `zone-1.my-ingress.kuma-system-zone-1.kuma-system-global`

Example 2: Zone(universal) -> Global(k8s) -> Zone(universal)

User starts a ZoneIngress in Zone `zone-1`:
```yaml
type: ZoneIngress
name: my-ingress
```

* Name in the Kuma API on Zone CP of `zone-1`: `my-ingress`
* Name of hte CRD on Global CP: `zone-1.my-ingress`
* Name in the Kuma API on Global CP: `zone-1.my-ingress.kuma-system-global`
* Name in the Kuma API on Zone CP of `zone-2`: `zone-1.my-ingress.kuma-system-global`

Notes:
* in KDS v1 instead of `.kuma-system-zone-2` we have `.default`

### Global(universal) -> Zone(universal)

```
nameInZone := "${nameInGlobal}"
```

Example 1: Zone(k8s) -> Global(universal) -> Zone(universal)

User starts a ZoneIngress in Zone `zone-1`:
```yaml
kind: ZoneIngress
metadata:
  name: my-ingress
  namespace: kuma-system-zone-1
```

* Name in the Kuma API on Zone CP of `zone-1`: `my-ingress.kuma-system-zone-1`
* Name in the Kuma API on Global CP: `zone-1.my-ingress.kuma-system-zone-1`
* Name in the Kuma API on Zone CP of `zone-2`: `zone-1.my-ingress.kuma-system-zone-1`

Example 2: Zone(universal) -> Global(universal) -> Zone(universal)

User starts a ZoneIngress in Zone `zone-1`:
```yaml
type: ZoneIngress
name: my-ingress
```

* Name in the Kuma API on Zone CP of `zone-1`: `my-ingress`
* Name in the Kuma API on Global CP: `zone-1.my-ingress`
* Name in the Kuma API on Zone CP of `zone-2`: `zone-1.my-ingress`
