# Reachable and Autoreachable services with MeshService and MeshExternalService

* Status: accepted

## Context and Problem Statement

Reachable and autoreachable services allow users to define which services they want to communicate with, reducing the memory footprint of a sidecar. Before the introduction of `MeshService` and `MeshExternalService`, a service wasn't a real object but a value of the tag `kuma.io/service`. Since the introduction of `MeshService` and `MeshExternalService`, that changed, and we can refer to them as real objects. Unfortunately, it doesn't work with the current logic and requires implementation.

### Reachable Services
The functionality allows the user to define a static list of services they want to communicate with. Currently, on Kubernetes, this is possible by setting the annotation `kuma.io/transparent-proxying-reachable-services` with the values of kuma.io/service

Example:

```yaml
...
  annotations:
    kuma.io/transparent-proxying-reachable-services: "demo-app_kuma-demo_svc_5000,redis_kuma-demo_svc_6379"
...
```

On Universal, it's a bit different and requires the user to provide this information in a `Dataplane` object.

Example:

```yaml
....
    transparentProxying:
      reachableServices:
      - demo-app_kuma-demo_svc_5000
      - redis_kuma-demo_svc_6379
....
```
This model is not extendable and doesn't fit the current resources.

### Autoreachable service

This is an experimental feature that achieves the same functionality as reachable services but dynamically, based on the `MeshTrafficPermission` policy. Users need to provide policies, after which the control plane creates a graph that restricts sidecars to only necessary services.

For example, the following policy enables `demo-app` to communicate with `redis`. When we enable the autoreachable services feature, the sidecar of `demo-app` will only monitor endpoints of the `redis` service.

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  namespace: kuma-system
  name: mtp-demo-to-redis
spec:
  targetRef:
    kind: MeshSubset
    labels:
      kuma.io/service: "redis_kuma-demo_svc_6379"
  from:
    - targetRef:
        kind: MeshService
        labels:
          kuma.io/servic: "demo-app_kuma-demo_svc_5000"
      default:
        action: Allow
```

More details in (MADR #036)

## Considered Options

* Support reachable and autoreachable services for `MeshService`, `MeshExternalService` and `MeshMultiZoneService`

## Decision Outcome

Chosen option: Support reachable and autoreachable services for `MeshService` and `MeshExternalService`

### Reachable services

In the case of Kubernetes, we can introduce a new annotation that has a more structured model and better flexibility.

```yaml
kuma.io/reachable-backends: |
  refs:
  - kind: MeshService
    name: demo-app
    namespace: kuma-demo # if missing, then my namespace
    port: 5000 # if missing then all ports of a given service
  - kind: MeshService
    name: redis
    namespace: redis-system
  - kind: MeshService
    labels:
      kuma.io/display-name: xyz
      kuma.io/zone: east
  - kind: MeshExternalService
    name: demo-app
    namespace: kuma-system
  - kind: MeshExternalService
    labels:
      kuma.io/display-name: httpbin
```

This model is more extendable, and if we introduce a new kind in the future, we can easily extend it. It is also similar to targetRef. The disadventage might be it takes much more space, and might be a bit difficult to manage within an annotation.

Universal needs another entry in a `Dataplane` object:

```proto
message TransparentProxying {
...
    message ReachableBackendRef {
        // Type of the backend: MeshService or MeshExternalService
        //  +required
        string kind = 1;

        // Name of the backend.
        //  +optional
        string name = 2;

        // Namespace of the backend. Might be empty
        //  +optional
        string namespace = 3;

        // Port of the backend.
        //  +optional
        google.protobuf.UInt32Value port = 4;

        // Labels used to select backends
        //  +optional
        map<string, string> labels = 5;
    }

    message ReachableBackends { repeated ReachableBackendRef refs = 1; }
    // Reachable backend via transparent proxy when running with
    // MeshExternalService, MeshService and MeshMultiZoneService. Setting an
    // explicit list of refs can dramatically improve the performance of the
    // mesh. If not specified, all services in the mesh are reachable.
    ReachableBackends reachable_backends = 7;
}
```

Most of the fields are optional and may not be used simultaneously. We need to support the following cases:
* When `name`, `namespace`, and `port` are defined(combination of them, all are not requried), you cannot use `labels`.
* When `labels` are defined, you cannot use `name`, `namespace`, and `port`.

### Dataplane backendRef

Another thing is adding another fields to the `BackendRef` in Dataplane object: namespace and labels
```proto
      message BackendRef {
        // Kind is a type of the object to target. Allowed: MeshService, MeshExternalService
        string kind = 1;
        // Name of the targeted object
        string name = 2;
        // Port of the targeted object. Required when kind is MeshService, MeshExternalService
        uint32 port = 3;
        // Namespace of the targeted object.
        //  +optional
        string namespace = 4;
        // Labels of the targeted object
        map<string, string> labels = 5;
      }
```

This enhancement should make it easier to match the reachable services provided by the user with the generated or provided outbounds. There is a disadventage that makes it maybe not the best option. If we expose it to the user, it might configure one `BackendRef` that reference many MeshServices. 

Because of that we will retrive `MeshService` and match based on its dpTags.

### Autoreachable services

From a code perspective, we used to build a map of supported tags: `kuma.io/service`, `k8s.kuma.io/namespace`, `k8s.kuma.io/service-name`, and `k8s.kuma.io/service-port`, and later match `MeshTrafficPermission` to dataplanes with these tags. Since the introduction of `MeshService` and `MeshExternalService`, we might not need to iterate over all dataplanes. Instead, we can match all `MeshService` and `MeshExternalService` entries. Based on this information, we know which dataplanes need to be configured. The only limitation is that `MeshService` doesn't carry all tags that a dataplane might have, which means that when using `MeshSubset`, users should be allowed to use the same tags as `dpTags` defined in `MeshService`.

Algorithm:

1. Take each `MeshService`
2. For each `MeshService`, take tags from dpTags (`+kuma.io/zone`, +`kuma.io/origin`) and build a fake "DPP" with one inbound
3. Execute MatchedPolicies on this "DPP" with a list of `MeshTrafficPermission` to get rules, in this case we have to trim tags from `MeshTrafficPermission` to have only one from `DPP`.
4. Store rules for each `MeshService`.
5. When we build XDS config for client DPP and check if we should add a cluster/listener for given `MeshService`, execute rules for `MeshService` using client DPP.

Targetting:

When a user wants to create a `MeshTrafficPermission` that targets a `MeshService`, they cannot do it directly in the top-level TargetRef. Instead, they can use `MeshSubset` to select a specific `MeshService`

```yaml
kind: MeshService
spec:
  selector:
    dpTags:
      app: backend
      k8s.kuma.io/namespace: xyz
---
kind: MeshTrafficPermission
spec:
  targetRef:
    kind: MeshSubset
    tags:
      app: backend
```

In the case of `MeshExternalService`, we need to decide how we want to target them (this needs to be clarified in a separate MADR). For now, we are going to provide all `MeshExternalService` entries until we figure out targeting.
Covered by the issue https://github.com/kumahq/kuma/issues/10868

### Positive Consequences

* Feature parity with the current solution
* Clearer and more extendable structure
* More intuitive selector section

### Negative Consequences
* More code that we need to support
* Annotation for reachable services can become quite large and difficult to manage 
