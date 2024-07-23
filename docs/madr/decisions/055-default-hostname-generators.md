# Default HostnameGenerators

* Status: accepted

## Context and Problem Statement

We introduced HostnameGenerator for MeshService, MeshExternalService and MeshMultizoneService.
This MADR proposes default generators taking into account different scenarios.

## Considered Options

Template and matching:
* Proposed options

Shipping:
* Document
* Create

Suffix:
* `.{{ type }}.mesh.{{ zone }}`
* `.{{ type }}.{{ optional zone }}.mesh.local`

## Decision Outcome

Template and matching - Proposed options.
Shipping - short term document. Once we get all prerequisites and MeshService/MeshExternalService/MeshMultizoneService goes GA, we create by default.
Suffix - `.{{ type }}.{{ optional zone }}.mesh.local`, because it's more "correct" (no conflicts).

## Pros and Cons of the Options - Template and matching

### Prerequisites

**Labels**
* `kuma.io/origin`. The value is either `zone` or `global`.
  The initial idea was to preserve the label between Zones, so that `kuma.io/origin: zone` in Zone A means that it's also the same on Global and on Zone B (if resource is synced).
  However, we diverged from this, so that Zone A has `kuma.io/origin: zone`, Global has `kuma.io/origin: zone` and Zone B has `kuma.io/origin: global` (because it was synced from global).
  We could adjust the implementation to initial assumptions, but it requires careful handling. If we were to do this we would have to introduce something like `kuma.io/synced: "true"`.
  We found existing state to be useful for hostname generators, and we did not find good use case to leverage this label preservation across zones.
* `kuma.io/env`. The value is either `kubernetes` or `universal`.
  Introducing this label is out of scope of this MADR.

**Display name**
`{{ .DisplayName }}` resolves to `kuma.io/display-name` if it exists and it fallbacks to name.
We can remove the fallback if we add `kuma.io/display-name` to Universal resources by default.
It's better than `{{ label "kuma.io/display-name" }}`, because looking at Kube resources it's annotation, not a label, so it might be confusing for users.

### MeshService

#### Local Kubernetes MeshService

It's rather unlikely for user to use custom hostnames to consume MeshService that is local in the Kubernetes cluster.
Therefore, no default HostnameGenerator is needed for it. If there is a need for it. We can define it as follows

```yaml
name: local-kube-mesh-service
spec:
  template: '{{ label "k8s.kuma.io/service-name" }}.{{ .Namespace }}.svc.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/env: kubernetes # short term "kuma.io/managed-by: kube-controller"
        kuma.io/origin: zone # only select local resources
        k8s.kuma.io/is-headless-service: false
```

This HostnameGenerator could be automatically created by Zone CP.

#### Synced Kubernetes MeshService

```yaml
name: synced-kube-mesh-service
spec:
  template: '{{ label "k8s.kuma.io/service-name" }}.{{ .Namespace }}.svc.{{ .Zone }}.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/env: kubernetes # short term "kuma.io/managed-by: kube-controller"
        kuma.io/origin: global # only select synced resources
        k8s.kuma.io/is-headless-service: false
```

It's important to only select Kubernetes MeshServices, so we won't try and fail to evaluate this template for Universal MeshServices.
MeshService controller will add `kuma.io/managed-by: kube-controller` to all services created by it.
It means that short term we won't take into account manually applied MeshServices on Kubernetes.
This is not something we want to recommend to users anyway. They should use Service object. 

This HostnameGenerator could be automatically created by Global CP so that it's synced down to every Zone CP.

#### Local Headless Kubernetes MeshService

It's rather unlikely for user to use custom hostnames to consume MeshService that is local in the Kubernetes cluster.
Therefore, no default HostnameGenerator is needed for it. If there is a need for it. We can define it as follows

```yaml
name: local-headless-kube-mesh-service
spec:
  template: '{{ label "statefulset.kubernetes.io/pod-name" }}.{{ label "k8s.kuma.io/service-name" }}.{{ .Namespace }}.svc.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/env: kubernetes # short term "kuma.io/managed-by: kube-controller"
        kuma.io/origin: zone # only select local resources
        k8s.kuma.io/is-headless-service: true
```

This HostnameGenerator could be automatically created by Zone CP.

#### Synced Headless Kubernetes MeshService

```yaml
name: synced-headless-kube-mesh-service
spec:
  template: '{{ label "statefulset.kubernetes.io/pod-name" }}.{{ label "k8s.kuma.io/service-name" }}.{{ .Namespace }}.svc.{{ .Zone }}.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/env: kubernetes # short term "kuma.io/managed-by: kube-controller"
        kuma.io/origin: global # only select synced resources
        k8s.kuma.io/is-headless-service: true
```

This HostnameGenerator could be automatically created by Global CP so that it's synced down to every Zone CP.

#### Local Universal MeshService

```yaml
name: local-universal-mesh-service
spec:
  template: '{{ .Name }}.svc.mesh.local' # .DisplayName would work as well
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: zone # only take into account local MeshServices
```

Universal zones do not have hostnames, so we need to create HostnameGenerator for them.

This HostnameGenerator could be automatically created by Universal Zone CP.

#### Synced Universal MeshService

```yaml
name: synced-universal-mesh-service
spec:
  template: '{{ .DisplayName }}.svc.{{ .Zone }}.mesh.local'
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: global
        kuma.io/env: universal # short term the users need to apply it themselves 
```

This HostnameGenerator could be automatically created by Global CP so that it's synced down to every Zone CP.

### MeshExternalService

#### MeshExternalService applied on Zone CP

```yaml
name: local-mesh-external-service
spec:
  template: '{{ .DisplayName }}.extsvc.mesh.local'
  selector:
    meshExternalService:
      matchLabels:
        kuma.io/origin: zone # only consider local MeshExternalServices
```

This HostnameGenerator could be automatically created by Zone CP.

On Kubernetes, MeshExternalService can only be applied in the system namespace, so for example hostname for this MeshExternalService

```yaml
kind: MeshExternalService
metadata:
  name: httpbin
  namespace: kuma-system
```

Would be `httpbin.extsvc.mesh.local`

#### MeshExternalService applied on Global CP

```yaml
name: synced-mesh-external-service
spec:
  template: '{{ .DisplayName }}.extsvc.mesh.local'
  selector:
    meshExternalService:
      matchLabels:
        kuma.io/origin: global # only consider synced MeshExternalServices
```

This HostnameGenerator could be automatically created by Global CP.

### MeshMultizoneService 

#### MeshMultizoneService applied on Global CP

```yaml
name: synced-mesh-multi-zone-service
spec:
  template: '{{ .DisplayName }}.mzsvc.mesh.local'
  selector:
    meshMultiZoneService: {}
```

This HostnameGenerator could be automatically created by Global CP.

#### MeshMultizoneService applied on Zone CP

The plan is to only allow applying MeshMultizoneService on Global CP.
However, if we want to extend this we could just use the same hostname generator as above. 

## Pros and Cons of the Options - Shipping

### Documenting

We can just document HostnameGenerators.

### Creating

We can create HostnameGenerators for users. The what and where is described in the "Pros and Cons of the Options - Template and matching"

## Pros and Cons of the Options - Suffix

Types:
* `svc` for MeshService
* `extsvc` - MeshExternalService
* `mzsvc` - MeshMultiZoneService

### `.{{ type }}.mesh.{{ zone }}`

Examples:
* `redis.kuma-demo.svc.mesh.local` - local Kubernetes MeshService
* `elasticsearch.svc.mesh.local` - local Universal MeshService
* `redis.kuma-demo.svc.mesh.east` - synced Kubernetes MeshService
* `elasticsearch.svc.mesh.west` - synced Universal MeshService
* `httpbin.extsvc.mesh.global` - synced MeshExternalService
* `auth.mzsvc.mesh.global` - synced MeshMultiZoneService

Advantages:
* It's familiar to Kubernetes pattern of `svc.cluster.local`
* Clear distinction between `local` and `global` hostnames 

Disadvantages:
* `{{ zone }}` might violate TLD. Although we can validate zone name to prevent this.
* `{{ zone }}` can clash with existing TLD. What if someone name the zone with `com`?

### `.{{ type }}.{{ optional zone }}.mesh.local`

Examples:
* `redis.kuma-demo.svc.mesh.local`  - local Kubernetes MeshService
* `elasticsearch.svc.mesh.local` - local Universal MeshService
* `redis.kuma-demo.svc.east.mesh.local` - synced Kubernetes MeshService
* `elasticsearch.svc.west.mesh.local` - synced Universal MeshService
* `httpbin.extsvc.mesh.local` - synced MeshExternalService
* `auth.mzsvc.mesh.local` - synced MeshMultiZoneService

Advantages:
* Stable TLD. It's possible to add `.mesh.local` or `.svc.mesh.local` to search section in `/etc/resolv.conf` to simplify hostnames.
* `.local` TLD is reserved so there is no way of clashing with existing TLD

Disadvantages:
* Awkward `.local` when the service is "global".
