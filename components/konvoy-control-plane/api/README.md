# Konvoy Control Plane API

## Using

`Konvoy Control Plane API` consists of:

* `Discovery API`
* `Mesh API`

### Discovery API

`Discovery API` lets you integrate `Konvoy Control Plane` with a *service registry* specific to your organization, e.g. `Consul`, `Zookeeper`, etc.

Use `Discovery API` model to describe `Services` and `Workloads` you want to make available within *Konvoy Service Mesh*.

#### Examples

Yaml:
```yaml
# inventory
items:
- service:
    id:
      name: example
      namespace: demo
    endpoints:
    - address: 192.168.0.1
      port: 8080
      workload:
        id:
          name: example-xgg9df
          namespace: demo
        meta:
          labels:
            app: example
            version: 0.0.1
            kuma.io/region: us-east-1
            kuma.io/zone: us-east-1b
- workload:
    id:
      name: daily-report-f5seg
      namespace: demo
    meta:
      labels:
        job: daily-report
        kuma.io/region: us-west-2
        kuma.io/zone: us-west-2c
```

Golang
```go
import (
    discovery "github.com/Kong/konvoy/components/konvoy-control-plane/api/discovery/v1alpha1"
)

inventory := discovery.Inventory{
    Items: []*discovery.Inventory_Item{
      &discovery.Inventory_Item{
        ItemType: &discovery.Inventory_Item_Service{
          Service: &discovery.Service{
            Id: &discovery.Id{
              Namespace: "demo",
              Name:      "example",
            },
            Endpoints: []*discovery.Endpoint{
              &discovery.Endpoint{
                Address: "192.168.0.1",
                Port:    8080,
                Workload: &discovery.Workload{
                  Id: &discovery.Id{
                    Namespace: "demo",
                    Name:      "example-xgg9df",
                  },
                  Meta: &discovery.Meta{
                    Labels: map[string]string{
                      "app":                 "example",
                      "version":             "0.0.1",
                      "kuma.io/region": "us-east-1",
                      "kuma.io/zone":   "us-east-1b",
                    },
                  },
                },
              },
            },
          },
        },
      },
      &discovery.Inventory_Item{
        ItemType: &discovery.Inventory_Item_Workload{
          Workload: &discovery.Workload{
            Id: &discovery.Id{
              Namespace: "demo",
              Name:      "daily-report-f5seg",
            },
            Meta: &discovery.Meta{
              Labels: map[string]string{
                "job":                 "daily-report",
                "kuma.io/region": "us-east-1",
                "kuma.io/zone":   "us-east-1b",
              },
            },
          },
        },
      },
    },
  }
```

### Mesh API

TODO(yskopets):

## Developing

See [Developer Guide](DEVELOPER.md) for further details.
