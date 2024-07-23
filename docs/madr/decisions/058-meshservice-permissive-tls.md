# Permissive mTLS and MeshService

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/10896

## Context and Problem Statement

Kuma supports permissive mTLS, which is an mTLS mode in which applications can accept mTLS and plaintext traffic.
This is useful for introducing mTLS in the cluster without any downtime, because distributing mTLS certs takes more than 0ms,
the client needs to know if it can send mTLS traffic to a server.

They way we solved this is we store `issuedBackend` that indicates if mTLS cert was issued for DPP.
Then, we take a list of `DataplaneInsights` and aggregate this information in `ServiceInsight`.
Then we determine if the service is ready to accept mTLS traffic with the condition of
```
issued certificates == data plane proxies offline + data plane proxies online
```

This approach has one problem. When we introduce new Dataplane of the service, then we might get into situation where `ServiceInsight` is updated before Dataplane is even online or was issued the cert.
In this case the condition described above is false, and we stop sending mTLS traffic until DPP is online, cert was issued and new reconciliation of `ServiceInsight` happens.

With the introduction of MeshService we cannot rely on `ServiceInsight`, because it's an aggregation of `kuma.io/service`.

## Decision Drivers

Feature parity for MeshService

## Considered Options

* Option 1 - Use DataplaneInsight
* Option 2 - Store it in the Dataplane object
* Option 3 - Store it in MeshService

## Decision Outcome

Chosen option: "Option 3 - Store it in MeshService", because it's one without perf and potential bugs implications. 

## Pros and Cons of the Options

### Option 1 - Use DataplaneInsight

Instead of deciding on the whole Envoy Cluster whether we set `transportSocket` with mTLS or not, we could track if an endpoint is protected by mTLS.
Then we can store this in Endpoint metadata `mtls: true` and use `transportSocketMatch` to send mTLS traffic only to those endpoints.
That would mean that we would not store this in MeshService at all.  However, that would also mean we need to pull this information from `DataplaneInsight`.
This has a significant performance impact, because we would have a dependency on `DataplaneInsights` for xDS reconciliation.
We would need to pull this into MeshContext and include it in the mesh hash.

### Option 2 - Store it in the Dataplane object

Instead of introducing dependency on `DataplaneInsight` we could store this in `Dataplane` object.
In case of Kubernetes, that would mean we have a double writer - Kubernetes Controller and XDS callback that would update Dataplane object.
In case of Universal, we would have "yet another writer" - one to create Dataplane when it connects/disconnects, HDS and XDS callbacks.

This can lead to problems where one thing overrides another.

### Option 3 - Store it in MeshService

We can follow similar path as with ServiceInsight so to store mTLS readiness in MeshService itself.

```
status:
  tls:
    status: Ready
```

However, we can be a bit smarter how we set it.
We know that when you enable mTLS all new data plane proxies always start with a certificate.
Envoy is only ready when Secret with cert is delivered.
Assuming we don't disable mTLS, the status can only go one way from `NotReady` to `Ready`.

MeshService status updater can then have a following logic
* If mTLS is disabled, set `tls.status` to `NotReady`
* If mTLS is enabled and `tls.status` is `NotReady`, check `issued certificates == data plane proxies offline + data plane proxies online`. If true, set `tls.status` to `Ready`

We can store this is `tls.status`. It's synced to global for visibility.
It's not synced cross-zone, because it's only useful in the context of local zone. For cross zone traffic we require mTLS anyway.

**How would MeshMTLS policy affect this?**
It's not decided if MeshMTLS will expose selective mTLS (mTLS only for a subset of proxies).
If so, we would need to do policy matching of proxies selected in MeshService selector to check if mTLS is enabled or not.
