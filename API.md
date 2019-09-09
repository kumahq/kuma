# Kuma Mesh API (user-facing)

## CRDs

List of supported CRDs:

* `ProxyTemplate`

### ProxyTemplate

`ProxyTemplate` CRD allows users to fully customize configuration of `Envoy` proxies.

At a high level, `ProxyTemplate` specifies a list of configuration sources
that contribute to the overall `Envoy` config.

#### Sources of configuration

`ProxyTemplate` supports the following sources of configuration:

|   | Source                      | Status               |
| - | --------------------------- |:--------------------:|
| 1 | Predefined profiles         | draft implementation |
| 2 | Raw `Envoy` resources       | draft implementation |
| 3 | Templating engine (Jsonnet) | proposal             |
| 4 | User-defined profiles       | proposal             |

#### Usage

`Kuma Control Plane` generates configuration for a given `Envoy` sidecar using the following algorithm:
* `Kuma Control Plane` checks whether a `Pod` defintion contains `kuma.io/proxy-template` annotation
* If `kuma.io/proxy-template` annotation is present on a `Pod`, its value must be a name of a `ProxyTemplate` resource in the same namespace
* If `ProxyTemplate` resource with that name actually exists, `Kuma Control Plane` will use it to generate `Envoy` configuration
* In all other cases `Kuma Control Plane` will fall back to a "default" `ProxyTemplate`

E.g.,

`kuma.io/proxy-template` inside `Pod` definition:
```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    kuma.io/proxy-template: custom-template
  ...
spec:
  ...
```

`kuma.io/proxy-template` inside `Deployment` definition:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    # not the right place for `proxy-template` annotation
  ...
spec:
  template: # PodTemplateSpec
    metadata:
      annotations:
        kuma.io/proxy-template: custom-template
  ...
```

`kuma.io/proxy-template` inside `Job` definition:
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  annotations:
    # not the right place for `proxy-template` annotation
  ...
spec:
  template: # PodTemplateSpec
    metadata:
      annotations:
        kuma.io/proxy-template: custom-template
  ...
```

#### Examples

##### Predefined profiles

```yaml
apiVersion: kuma.io/v1alpha1
kind: ProxyTemplate
metadata:
  name: uses-predefined-profiles
spec:
  sources:
  - profile:
      name: default-proxy
```

##### Raw Envoy xDS resources

`Envoy` resource as YAML string:
```yaml
apiVersion: kuma.io/v1alpha1
kind: ProxyTemplate
metadata:
  name: raw-envoy-xds-resources
spec:
  sources:
  - raw:
      resources:
      - name: localhost:8080
        resource: |
          '@type': type.googleapis.com/envoy.api.v2.Cluster
          connectTimeout: 5s
          name: localhost:8080
          loadAssignment:
            clusterName: localhost:8080
            endpoints:
            - lbEndpoints:
              - endpoint:
                  address:
                    socketAddress:
                      address: 127.0.0.1
                      portValue: 8080
          type: STATIC
        version: v1
```

`Envoy` resource as JSON string:
```yaml
apiVersion: kuma.io/v1alpha1
kind: ProxyTemplate
metadata:
  name: raw-envoy-xds-resources
spec:
  sources:
  - raw:
      resources:
      - name: localhost:8080
        resource: |
            {
              "@type": "type.googleapis.com/envoy.api.v2.Cluster",
              "connectTimeout": "5s",
              "loadAssignment": {
                "clusterName": "localhost:8080",
                "endpoints": [
                  {
                    "lbEndpoints": [
                      {
                        "endpoint": {
                          "address": {
                            "socketAddress": {
                              "address": "127.0.0.1",
                              "portValue": 8080
                            }
                          }
                        }
                      }
                    ]
                  }
                ]
              },
              "name": "localhost:8080",
              "type": "STATIC"
            }
        version: v1
```

##### Templating engine (Jsonnet)

WARNING: This feature hasn't been implemented yet

```yaml
apiVersion: kuma.io/v1alpha1
kind: ProxyTemplate
metadata:
  name: raw-envoy-xds-resources-generated-by-jsonnet-script
spec:
  sources:
  - generator:
      jsonnet:
        script: |
          ...
        params:
          a: b
          c: d
```

##### User-defined profiles

WARNING: This feature hasn't been implemented yet

```yaml
apiVersion: kuma.io/v1alpha1
kind: Profile
metadata:
  name: custom-profile
spec:
  params:
  - name: param1
  - name: param2
  generator:
    jsonnet:
      script: |
        ...
---
apiVersion: kuma.io/v1alpha1
kind: ProxyTemplate
metadata:
  name: uses-custom-profile
spec:
  sources:
  - profile:
      name: custom-profile
      params:
        param1: value1
        param2: value2
```

#### Predefined profiles

|   | Profile                      | Description | Generated Envoy xDS resources  |
| - | ---------------------------- | - | ----------------------------- |
| 1 | `default-proxy`  | Setup proxying according to `Dataplane` descriptor: <br><br> 1. Setup transparent proxying <br> 2. Forward *inbound* requests to local `Clusters` <br> 3. (transparent proxying) Forward *outbound* requests to original destinations | 1. `Listener` on port 15001 (with `UseOriginalDst: true`) <br> 2. "Pass-through" `Cluster` with `LbPolicy: ORIGINAL_DST_LB` <br> 3. `Listener` per each `Inbound` interfaces of the `Dataplane` <br> 4. Local `Cluster` per each unique `WORKLOAD_PORT` in `Inbound` interfaces of the `Dataplane` |

#### Known limitations

1. "Default" `ProxyTemplate` is hardcoded inside `Kuma Control Plane`
2. `kuma.io/proxy-template` annotation must be attached directly to a `Pod` or `PodTemplateSpec` (see examples above)
3. Only 1 predefined profile
4. If multiple configuration sources produce an xDS resource with the same name, the latest definition wins

Such constraints greatly simplify the initial implementation.
