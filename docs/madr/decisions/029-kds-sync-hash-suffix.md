# Use hash-suffix for policy naming during KDS sync

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/3837

## Context and Problem Statement

Primary keys are different for resources on Universal and Kubernetes:

* Universal: (name, namespace, mesh, type)
* Kubernetes: (name, namespace, kind)

This results in inconsistent state when Global(Universal) and Zone(Kubernetes):

* Global(Universal):
```yaml
type: MeshTrafficPermission
name: allow-all
mesh: mesh-1
spec: {}
---
type: MeshTrafficPermission
name: allow-all
mesh: mesh-2
spec: {}
```

* Zone(Kubernetes):
```yaml
kind: MeshTrafficPermission
metadata:
  name: allow-all
  namespace: kuma-system
  labels:
    kuma.io/mesh: mesh-1
spec: {}
```

MeshTrafficPermission from `mesh-2` is not synced to Zone because there is already MeshTrafficPermission with 
a key `(allow-all, kuma-system, MeshTrafficPermission)`.

## Considered Options

* use hash-suffix for a name when syncing policies
* remove `mesh` from Universal primary key

## Decision Outcome

Chosen option: use hash-suffix for a name when syncing policies

### Implementation

KDS has client and server, server owns the resource, client receive update when resource is changing. 
Depending on the resource type Global and Zone CP both act as a client and as a server:
* Policies are created on Global CP (KDS Server), Zone CP receives copies of these policies (KDS Client). 
* Data plane proxies are created on Zone CP (KDS Server), Global CP receives copies of DPPs (KDS Client).

When this feature is implemented, KDS Server is going to name resources differently. Resource name is constructed like:

```go
name := fmt.Sprintf("%s-%s", resourceName, hash(resourceMesh, resourceZone, resourceNamespace))
```

Since renaming is happening on the KDS Server, KDS Client doesn't need to know anything about hashes. 
It will automatically remove old policies and create new policies. 

Information that we're using to build hash has to be set as `labels` on the resource. Today, 
when syncing DPP from Zone to Global we have the following name on Global:

```
zone-1.my-dpp.ns-from-zone
```

but with hash, the name will look like

```
my-dpp-c8fh421p
```

Because we don't want to lose the information about zone and namespace, we have to put it as labels:

* Zone(Kubernetes)
```yaml
kind: Dataplane
metadata:
  name: my-dpp
  namespace: my-ns
  labels: 
    kuma.io/mesh: mesh-1
```

* Global(Kubernetes)
```yaml
kind: Dataplane
metadata:
  name: my-dpp-c8fh421p
  namespace: kuma-system
  labels: 
    kuma.io/mesh: mesh-1
    kuma.io/namespace: ns-from-zone
    kuma.io/zone: zone-1
```

Universal doesn't have `labels` in the resource schema. That's why we have to introduce a new attribute 
in `resources` table in Postgres called `labels`. 

Implementation can be shipped in 2 steps:

1. Do a bare minimum to unblock [#3837](https://github.com/kumahq/kuma/issues/3837). This means we add hash-suffixes
to policies when syncing them from Global to Zone. Since such hash-suffix is built using only `mesh` we don't need to
implement Universal labels at this step. 

2. Implement Universal labels. This will unblock the use-case with Zone(Universal) and Global(Kubernetes). Also solves
the issue with long names on k8s (because we're adding zone+namespace when syncing DPPs). This step also unblocks the 
ongoing work related to Zone to Global policy sync. When implementing this step make sure Inspect API works in Multizone.

## Pros and Cons of the Options <!-- optional -->

### Use hash-suffix for a name when syncing policies

* Good, because it's not a breaking change 
* Good, because solves the problem of long names on k8s
* Bad, because name doesn't tell much about the origin of the policy (user has to check labels)

### Remove `mesh` from Universal primary key

* Good, because naming is the same on k8s and Universal
* Bad, because it's a breaking change, upgrade path is quite tricky
