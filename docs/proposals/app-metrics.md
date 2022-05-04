# Application metrics
 
## Context
 
Currently, when an application exposes metrics as regular service it's protected and traffic is secured (mTLS enabled), it's a bit tricky
to [expose those metrics without mTLS](https://kuma.io/docs/dev/policies/traffic-metrics/#expose-metrics-from-applications). Also, by configuring paths user's can by mistake expose more than it's necessary. To avoid it we should come up with a solution that fulfills all the requirements: scraping metrics with/without mTLS and safer exposing of metrics.
 
## Requirements
* support different application ports/paths
* support more than one container/application
* support no mTLS communication with Prometheus
 
## Idea
 
The idea is to create an aggregator in `kuma-dp`, to be specific Metrics Hijacker, that will be responsible for the aggregation of metrics from applications defined to scrap(within local env: pod/vm). Communication with `kuma-dp` might be configured as secure but it won't be required(and not recommended). It's worth mentioning that this idea can support only Prometheus endpoint scrapping and exposing metrics as Prometheus.
 
### Universal
At the mesh level, it will be possible to define the default exposed path and port at which metrics are going to be accessible.
```
type: Mesh
name: default
metrics:
 enabledBackend: prometheus-1
 backends:
 - name: prometheus-1
   type: prometheus
   conf:
    skipMTLS: false
    port: 5670 # default exposed port with metrics 
    path: /metrics
    tags:   
      kuma.io/service: dataplane-metrics
```
`Dataplane` configuration enables at which port/path `kuma-dp` is going to expose aggregated metrics from Envoy and other applications running within one workflow. Also, it's possible to override the `Mesh` configuration of endpoints to scrap by the `Dataplane` configuration. They are going to be identified by `name` and if `Mesh` and `Dataplane` configurations have definitions with the same name, then the `Dataplane` configuration has precedence before `Mesh` and is merged.
```
type: Dataplane
mesh: default
name: redis-1
networking:
 address: 192.168.0.1
 inbound:
 - port: 9000
   servicePort: 1234
   tags:
     kuma.io/service: backend
 metrics:
   type: prometheus
   conf:
     skipMTLS: true
     port: 1234
     path: /non-standard-path
     aggregate:
     - name: app
       port: 1236
       path: /metrics
     - name: opa
       port: 12345
       path: /stats
```
 
### Kubernetes
 
The same `Mesh` configuration works for K8s and Universal.
```
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
 name: default
spec:
 metrics:
   enabledBackend: prometheus-1
   backends:
   - name: prometheus-1
     type: prometheus
     conf:
       skipMTLS: false
       port: 5670
       path: /metrics
       tags: # tags that can be referred in Traffic Permission when metrics are secured by mTLS 
         kuma.io/service: dataplane-metrics
```
In the case of Kubernetes configuration of endpoints to scrap is done by label `prometheus.metrics.kuma.io/aggregate-{name}-port` and `prometheus.metrics.kuma.io/aggregate-{name}-path`, where `{name}` is used for distinguish between config for different containers/apps running in the pod.
 
```
apiVersion: apps/v1
kind: Deployment
metadata:
 name: example-app
 namespace: kuma-example
spec:
 ...
 template:
   metadata:
     ...
     labels:
       # indicate to Kuma that this Pod doesn't need a sidecar
       kuma.io/sidecar-injection: enabled
     annotations:
       prometheus.metrics.kuma.io/aggregate-app-port: 1234 # format is aggregate-{name}-port
       prometheus.metrics.kuma.io/aggregate-app-path: "/metrics"
       prometheus.metrics.kuma.io/aggregate-opa-port: 12345 # format is aggregate-{name}-port
       prometheus.metrics.kuma.io/aggregate-opa-path: "/stats"
   spec:
     containers:
       ...
```
 
## Use existing prometheus tags

Prometheus on K8s uses annotation `prometheus.io/{path/port/scrapping}`, we might use it to map them for kuma specifc 
annotations that are going to be used to scrap endpoints. At the begging we might not implement this feature but in future we might get this.

## Labels

We are not going to change any metric's label so stats from applications are going to be unchanged.

### How to get information about path/port for `kuma-dp` to expose metrics?
 
 * Bootstrap config is going to return information about path/part. This configuration is going to be static and won't be possible to change it during application runtime.
