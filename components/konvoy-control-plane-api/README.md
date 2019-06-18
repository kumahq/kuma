# Konvoy Control Plane API

## Using

`Konvoy Control Plane API` consists of:

* `Discovery API`
* `Mesh API`

### Discovery API

`Discovery API` lets you integrate `Konvoy Control Plane` with a *service registry* specific to your organization, e.g. `Consul`, `Zookeeper`, etc.

Use `Discovery API` model to describe `Services`, `Endpoints` and `Workloads` you want to make available within *Konvoy Service Mesh*.

#### Examples

Yaml:
```yaml
TODO(yskopets):
```

Golang
```go
import (
    discovery "github.com/Kong/konvoy/components/konvoy-control-plane-api/discovery/v1alpha1"
)

svc := discovery.Service{
    Name:      "example",
    Endpoints: []*discovery.Endpoint{
        &{
            Address:  "192.168.0.1",
            Port:     8080,
            Workload: &discovery.Workload{
                Id:   &discovery.Id{
                    Name: "example-xgg9df",
                },
                Meta: &discovery.Meta{
                    Labels: map[string]string{
                        "app": "example",
                        "version": "0.0.1",
                        "getkonvoy.io/region": "us-east-1",
                        "getkonvoy.io/zone": "us-east-1b",
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
