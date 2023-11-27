# Adding inbounds in runtime / Changing Kube Service for the Pod 

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/5086

## Context and Problem Statement

Changing Kubernetes Service for a Pod results in changing / adding a new inbound in Dataplane object.

Adding a new inbound for a Dataplane object results in two parallel operations:
* Updating changed data plane proxy with new certificate
* Updating all clients with new endpoint targeting changed data plane proxy

If second applies before the first, it can result in broken traffic.

This can happen with ArgoCD and Blue Green rollout.
First the pod is selected by `service-preview` and if it passes the analysis the pod is targeted by the `service-active`.

Here is an example of `redis` from `kuma-demo` going through blue-green promotion 
```yaml
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: kuma-demo
spec:
  ports:
    - port: 6379
      protocol: TCP
      targetPort: 6379
  selector:
    app: redis
    rollouts-pod-template-hash: 778dbdddff
---
apiVersion: v1
kind: Service
metadata:
  name: redis-preview
  namespace: kuma-demo
spec:
  ports:
    - port: 6379
      protocol: TCP
      targetPort: 6379
  selector:
    app: redis
    rollouts-pod-template-hash: 646998df5c
```
Once it's finished, `redis` will change selector to `rollouts-pod-template-hash: 646998df5c` and `redis-preview` will keep `rollouts-pod-template-hash: 646998df5c` selector. 

## Decision Drivers

* Enable users to use ArgoCD blue-green rollout without traffic interruption

## Considered Options

* Wait for certs using Dataplane Insights
* Wait for certs using XDS callbacks
* Ignore labels in selector to add ignored listener

## Decision Outcome

Chosen option: "Ignore labels in selector to add ignored listener". While it's the hardest to explain. It has no impact on the perf.

## Pros and Cons of the Options

To consider options, we need to first introduce new concept of "Ignored" listener.
Instead of `healthy.ready: false/true` we also need another state of the listener.

```yaml
type: Dataplane
mesh: default
name: redis-646998df5c-nzx4b.kuma-demo
networking:
  address: 10.42.0.25
  inbound:
    - port: 6379
      tags:
        app: redis
        kuma.io/service: redis_kuma-demo_svc_6379
        rollouts-pod-template-hash: 646998df5c
      state: Ready | NotReady | Ignored
```

* Ready - listener is ready to receive the traffic (equivalent of existing `healthy.ready: true`)
* NotReady - listener is not ready to receive the traffic (equivalent of existing `healthy.ready: false`)
* Ignored - listener is not created. It cannot be targeted by policies. It will however receive a cert with `kuma.io/service` of this listener.

`state` field would replace `healthy` branch. 

**Why we need Ignored state. Can't we just use NotReady?**
* Targeting outbound. Let say we have `backend` and `backend-preview` services. We want to redirect the traffic from `backend-preview` to `redis-staging`.
  We create traffic policies `backend->redis` and `backend-preview->redis-staging`.
  Let's say we have a data plane proxy in the middle of ArgoCD rollout, so with Ready `backend-preview` listener and NotReady `backend`.
  It may match either `backend->redis` or `backend-preview->redis-staging`, but we only really want to match `backend-preview->redis-staging`. 
* GUI handling. Data plane proxy with at least one unhealthy inbound is considered offline. Service with unhealthy inbound is considered partially degraded.
  This might be confusing for users as rollout is an operation that brings more attention to inspect what is going on in infra.
  Ignored listeners can be just not be considered when computing status of data plane proxy / Service.

### Wait for certs using Dataplane Insights

Ideally we would like to just mark the inbound as `healthy.ready: false` until we know we received certificate for `kuma.io/service` in this listener.
Inbounds with `healthy.ready: false` are not included in endpoints set for clients.

However, implementing this is quite difficult.
With multiple instances of the control plane, it might be so one instance is responsible for generating Dataplane object (leader - Pod Controller)
and another instance is handling XDS of this data plane proxy - it has information that data plane proxy received cert.
To synchronize this information between instances we would need to store it in DataplaneInsight.
Then, when generating Dataplane from Pod we need to check DataplaneInsight.

Disadvantages:
* It means that Pod Controller has to watch DataplaneInsights (which are changed quite often), this would introduce a significant performance impact on the system.

This option does not require introducing new "Ignored" state.

### Wait for certs using XDS callbacks

Pod Controller could detect that we are just about to add listener in runtime for healthy pod, and it would mark listener as "Ignored"
We would have a new XDS Callbacks that whenever we deliver a new cert and data plane proxy confirms them (sent ACK), check if Dataplane object has "Ignored" listener and update it to be ready.

Disadvantages:
* We have two writers, which can lead to races, bugs etc.
* We need to fetch (and potentially update) Dataplane object whenever there is cert delivery = more requests to the database.

### Ignore labels in selector to add ignored listener

We could introduce a new setting `KUMA_RUNTIME_KUBERNETES_INJECTOR_IGNORED_SERVICE_SELECTOR_LABELS`.
This setting would cause selecting more services for inbounds. However, not fully matched services would create "Ignored" listener.

Example. Let's say we configure this setting to "rollouts-pod-template-hash".
Then, we have Pod with labels `app=redis,rollouts-pod-template-hash=646998df5c`.
If we have services defined in "Context and Problem Statement", the pod would receive:
* "Ignored" listener for "redis" service
* Ready listener for "redis-preview"

This setting has to be configured only by mesh operator, therefore it cannot be annotation on the service.
Otherwise, service owner could put `app` in the annotation and potentially receive identity of majority services in the mesh.

Advantages:
* No performance impact.

Disadvantages:
* It's opt-in setting, so it won't work out of the box.
