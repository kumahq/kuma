# What is the name of the resource after KDS sync?

## Old policies (i.e. TrafficRoute, TrafficPermission)

* CRD scope: Cluster
* Kuma scope: Mesh

### Global(k8s/universal) -> Zone(k8s/universal)

```
nameInZone := "${truncate(nameInGlobal, 236)}-${hash(mesh, nameInGlobal)}"
```

Example: `my-policy -> my-policy-fvdw2c2wx44fv4xz`

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

Example: `my-policy.kuma-system-global -> my-policy-fvdw2c2wx44fv4xz.kuma-system-zone`

### Global(universal) -> Zone(k8s)

```
nameInZone := "${truncate(nameInGlobal, 236)}-${hash(mesh, nameInGlobal)}.${systemNamespace}"
```

Example: `my-policy -> my-policy-fvdw2c2wx44fv4xz.kuma-system-zone`

### Global(k8s) -> Zone(universal)

```
nameInGlobal = trimNsSuffix(nameInGlobal)
nameInZone := "${truncate(nameInGlobal, 236)}-${hash(mesh, nameInGlobal)}"
```

Example: `my-policy.kuma-system-global -> my-policy-fvdw2c2wx44fv4xz`

### Global(universal) -> Zone(universal)

```
nameInZone := "${truncate(nameInGlobal, 236)}-${hash(mesh, nameInGlobal)}"
```

Example: `my-policy -> my-policy-fvdw2c2wx44fv4xz`

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

Example: `my-dpp.kuma-demo -> zone-1.my-dpp.kuma-demo.kuma-system` 

Notes: 
* in KDS v1 (default at this moment) instead of `.kuma-system` we have `.default`
* DPP name collision is possible https://github.com/kumahq/kuma/issues/8249

### Zone(universal) -> Global(k8s)

```
nameInGlobal := "${zone}.${nameInZone}.${systemNamespce}"
```

Example: `my-dpp -> zone-1.my-dpp.kuma-system`

Notes:
* in KDS v1 (default at this moment) instead of `.kuma-system` we have `.default`
* DPP name collision is possible https://github.com/kumahq/kuma/issues/8249

### Zone(universal) -> Global(universal)

```
nameInGlobal := "${zone}.${nameInZone}"
```

Example: `my-dpp -> zone-1.my-dpp`

## ZoneIngress and ZoneEgress

* CRD scope: Namespace
* Kuma scope: Global

### Zone(k8s) -> Global(k8s)

```
nameInGlobal := "${zone}.${nameInZone}.${nsInZone}.${systemNamespace}" 
```

Example: `my-ingress.kuma-system-zone -> zone-1.my-ingress.kuma-system-zone.kuma-system-global`

Notes:
* in KDS v1 (default at this moment) instead of `.kuma-system-global` we have `.default`

### Zone(universal) -> Global(k8s)

```
nameInGlobal := "${zone}.${nameInZone}.${systemNamespace}" 
```

Example: `my-ingress -> zone-1.my-ingress.kuma-system-global`

Notes:
* in KDS v1 (default at this moment) instead of `.kuma-system-global` we have `.default`

### Zone(k8s) -> Global(universal)

```
nameInGlobal := "${zone}.${nameInZone}.${nsInZone}" 
```

Example: `my-ingress.kuma-system-zone -> zone-1.my-ingress.kuma-system-zone`

### Zone(universal) -> Global(universal)

```
nameInGlobal := "${zone}.${nameInZone}" 
```

Example: `my-ingress -> zone-1.my-ingress`

### Global(k8s) -> Zone(k8s)

```
nameInZone := "${nameInGlobal}.${systemNamespace}"
```

Example: `zone-1.my-ingress.kuma-system-zone-1.kuma-system-global -> zone-1.my-ingress.kuma-system-zone-1.kuma-system-global.kuma-system-zone-2`

Notes:
* in KDS v1 (default at this moment) instead of `.kuma-system-global` and `.kuma-system-zone-2` we have `.default`

### Global(universal) -> Zone(k8s)

```
nameInZone := "${nameInGlobal}.${systemNamespace}"
```

Example: `zone-1.my-ingress.kuma-system-zone-1 -> zone-1.my-ingress.kuma-system-zone-1.kuma-system-zone-2`

Notes:
* in KDS v1 (default at this moment) instead of `.kuma-system-zone-2` we have `.default`

### Global(k8s) -> Zone(universal)

```
nameInZone := "${nameInGlobal}"
```

Example: `zone-1.my-ingress.kuma-system-zone-1.kuma-system-global -> zone-1.my-ingress.kuma-system-zone-1.kuma-system-global`

Notes:
* in KDS v1 (default at this moment) instead of `.kuma-system-zone-2` we have `.default`

### Global(universal) -> Zone(universal)

```
nameInZone := "${nameInGlobal}"
```

Example: `zone-1.my-ingress.kuma-system-zone-1 -> zone-1.my-ingress.kuma-system-zone-1`
