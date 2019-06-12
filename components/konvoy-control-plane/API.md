# Konvoy Mesh API (user-facing)

## CRDs

List of supported CRDs:

* ProxyTemplate

### ProxyTemplate

`ProxyTemplate` CRD allows users to fully customize configuration of `Envoy` proxies.

At a high level, `ProxyTemplate` specifies a list of configuration sources
that contribute to the overall `Envoy` configuration.

#### Sources of configuration

`ProxyTemplate` supports the following sources of configuration:

|   | Source                      | Status               |
| - | --------------------------- |:--------------------:|
| 1 | Predefined profiles         | draft implementation |
| 2 | Raw `Envoy` resources       | draft implementation |
| 3 | Templating engine (Jsonnet) | proposal             |
| 4 | User-defined profiles       | proposal             |

#### Usage

`Konvoy Control Plane` generates configuration for a given `Envoy` sidecar using the following algorithm:
* `Konvoy Control Plane` checks whether a `Pod` defintion contains `mesh.getkonvoy.io/proxy-template` annotation
* If `mesh.getkonvoy.io/proxy-template` annotation is present on a `Pod`, its value must be a name of a `ProxyTemplate` resource in the same namespace
* If `ProxyTemplate` resource with that name actually exists, `Konvoy Control Plane` will use it to generate `Envoy` configuration
* In all other cases `Konvoy Control Plane` will fall back to a default `ProxyTemplate`

E.g.,

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    mesh.getkonvoy.io/proxy-template: custom-template
  name: nginx-98f78d5f6-qpwh6
  namespace: default
spec:
  ...
```

##### Known limitations

1. Default `ProxyTemplate` is hardcoded inside `Konvoy Control Plane`
2. `mesh.getkonvoy.io/proxy-template` annotation must be attached directly to a `Pod` (rather than to `Deployment`, `ReplicaSet`, `Job`, etc)

It greatly simplifies the initial implementation implementation.

#### Examples

##### Predefined profiles

```yaml
apiVersion: mesh.getkonvoy.io/v1alpha1
kind: ProxyTemplate
metadata:
  name: default
spec:
  sources:
  - profile:
      name: transparent-inbound-proxy
  - profile:
      name: transparent-outbound-proxy
```

##### Raw Envoy resources

`Envoy` resource as YAML string:
```yaml
---
apiVersion: mesh.getkonvoy.io/v1alpha1
kind: ProxyTemplate
metadata:
  name: default-and-raw
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
        version: v1
```

`Envoy` resource as JSON string:
```yaml
apiVersion: mesh.getkonvoy.io/v1alpha1
kind: ProxyTemplate
metadata:
  name: default-and-raw
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

WARNING: This is feature hasn't been implemented yet

```yaml
apiVersion: mesh.getkonvoy.io/v1alpha1
kind: ProxyTemplate
metadata:
  name: default-and-scripted
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

WARNING: This is feature hasn't been implemented yet

```yaml
apiVersion: mesh.getkonvoy.io/v1alpha1
kind: Profile
metadata:
  name: custom-profile
spec:
  generator:
    jsonnet:
    script: |
        ...
    params:
    - name: param1
    - name: param2
---
apiVersion: mesh.getkonvoy.io/v1alpha1
kind: ProxyTemplate
metadata:
  name: default-and-scripted
spec:
  sources:
  - profile:
      name: custom-profile
      params:
        param1: value1
        param2: value2
```
