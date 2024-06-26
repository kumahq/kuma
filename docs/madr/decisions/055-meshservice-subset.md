# Subsets of Dataplanes per MeshService

* Status: accepted

## Context and Problem Statement

Referring to a subset of endpoints for a given Service is important for service
mesh. It's most commonly use to implement things like blue/green deployments or canary
deployments.

For example if we have a `backend` service that we want to upgrade from `v1` to
`v2`. With `kuma.io/service`, this would mean having `Dataplanes` with
`kuma.io/service: backend` and an additional tag either `version: v1` or
`version: v2`. In order to do weighted routing to these different versions, a
user would create a `MeshHTTPRoute`:

```yaml
backendRefs:
- kind: MeshServiceSubset
  tags:
    version: v2
  name: backend
  weight: 10
- kind: MeshServiceSubset
  tags:
    version: v1
  name: backend
  weight: 90
```

Now that we have `MeshService` objects, the user has something like the following:

```yaml
kind: MeshService
spec:
  selector:
    dataplaneTags:
      app: backend
```

and matched `Dataplanes` have `version` tags set.

The question this MADR answers is how do users handle the same concept with
`MeshService`?

## Considered Options

* Support `MeshServiceSubset` with `MeshService` objects
* Users create additional `MeshServices`
  * On Kubernetes the user creates `Services`

## Decision Outcome

Chosen option: additional `MeshServices`, no more `MeshServiceSubset`

Concretely:

```yaml
---
kind: MeshService
name: backend
spec:
  selector:
    dataplaneTags:
      app: backend
---
kind: MeshService
name: backend-v1
spec:
  selector:
    dataplaneTags:
      app: backend
      version: v1
---
kind: MeshService
name: backend-v2
spec:
  selector:
    dataplaneTags:
      app: backend
      version: v2
```

```yaml
        backendRefs:
        - kind: MeshService
          name: backend-v2
          weight: 10
        - kind: MeshService
          name: backend-v1
          weight: 90
```

After the move to v2 is tested and complete, probably followed by:

```yaml
---
kind: MeshService
name: backend
spec:
  selector:
    dataplaneTags:
      app: backend
      version: v2
```

```yaml
        backendRefs:
        - kind: MeshService
          name: backend
          weight: 100
```

and eventually once `v1` is no longer running:

```yaml
---
kind: MeshService
name: backend
spec:
  selector:
    dataplaneTags:
      app: backend
```

### Positive Consequences

* Closer to widespread Kubernetes UX, liberal creation of Services
 * GAMMA
 * Argo
 * Flagger
* Easier implementation since all potential endpoint sets are explicit
  * xds
  * no more metrics merging

### Negative Consequences

* Universal UX. It's tedious to create and apply MeshServices with kumactl in deployment pipelines
  and effort to manage RBAC and permissions for service owners. But we can and should improve tooling here generally

## Pros and Cons of the Options

`MeshServiceSubset` is implicit, multiple `MeshServices` is explicit. This is
both an advantage and disadvantage.

### `MeshServiceSubset` support

This option would allow users to continue defining implicit subsets in routes,
endpoint selection would be:

1. apply `MeshService` `dataplaneTags` selector
2. apply `backendRef.tags` selector

#### Pros

* Similar to existing implicit subsetting of Kuma

#### Cons

* Fake kind `MeshServiceSubset` in targetRef

### Additional `MeshService` support

#### Pros

* Similar to how Kubernetes ecosystem handles this
 * GAMMA
 * Argo
 * Flagger
* Easier metrics management, no more merging needed
* xds implementation is easier since all potential subsets are always enumerated explicitly
* No more fake `MeshServiceSubset` kind

#### Cons

* Universal UX. It's a relative pain to create `MeshServices` manually at the
  moment
