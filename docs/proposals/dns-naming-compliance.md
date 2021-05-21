# DNS compliant names in Kuma

## Current state

### Service Naming

In a broader sense, Kuma uses the `kuma.io/service` tag to mark a certain combination of ip/address and port as a particular service, which then
can be used in the Policies selectors.

On Kubernetes, this tag is generated automatically and is an underscore delimited combination of the K8s Service Name, Namespace, and Port: `<name>_<namespace>_svc_<port>`.

On Universal, one still has to generate the name in an arbitrary form.

### Cross-zone communication

To account for multizone and hybrid deployments, we have extended the naming scheme through a DNS resolving service by adding a `.mesh` suffix.
This unifies the naming of services in a way that it does not matter where the service was instantiated. Of course the naming scheme in Kubernetes
is still applied, but now we can also also call `<name>_<namespace>_svc_<port>.mesh`.

By adding the transparent proxy capabilities to Universal, we can now refer to services in the `.mesh` domain too.

The DNS server in the kuma-cp will generate and assign virtual IPs (VIPs) so that any service name resolves to it, and that
the side-cars get an additional listener for the service cluster. The default VIPs CIDR is `240.0.0.1`.

### Exposing services outside the mesh

The typical way to expose services from within the mesh to the external consumers is through a gateway dataplane proxy (one the Kuma side) and
a reverse proxy on the customer-facing side. In this architecture, the traffic routes on the proxy should refer to Kuma services so that the
listeners on the side-car Envoy, are hit with the traffic and then load-balanced through the available local and remote endpoints.

On Kubernetes this happens using an Ingress resource of the form:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: marketplace
  namespace: default
  annotations:
    kubernetes.io/ingress.class: kong
    kubernetes.io/ingress.global-static-ip-name: "kong-demo-ip"
spec:
  rules:
  - host: localhost
    http:
      paths:
      - path: /
        backend:
          serviceName: frontend
          servicePort: 80
```

Where `serviceName: frontend` is the name of a Kubernetes Service object, which will be used by the proxy to redirect the incoming requests.
This same service in Kuma namespace will be called `frontend_default_svc_80` and can also be resolved as a DNS name `frontend_default_svc_80.mesh`.

However, the problem is that the Kubernetes Ingress Resource enforces 2 rules here:
* `serviceName` can refer only to a Kubernetes Service
* this service can not contain `_` or `.`

### Illustrating the problem
To illustrate the problem further, let's look at the option to manually create a Service of type `ExternalName`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: frontend
  namespace: default
  annotations:
    ingress.kubernetes.io/service-upstream: "true"
spec:
  type: ExternalName
  externalName: frontend_default_svc_80.mesh
```

```shell
The Service "frontend" is invalid: spec.externalName: Invalid value: "frontend_default_svc_80.mesh": a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')
```

The DNS name we try to supply is invalid according to the relevant RFCs and the Kubernetes embedded checks. So we are not even able to apply that resource.

## Proposal for improvements

In addition to the already existing Kubernetes generated `frontend_default_svc_80.mesh`, we can add a couple of compliant aliases:
* use dashes as separators `frontend-default-svc-80.mesh`. This can get compllicated, cosider a service name `my-complex-service`
  which naturally translates to `my-complex-service_default_svc_80.mesh`, and then the all dashes version `my-complex-service-default-svc-80.mesh`
* use dots as separators `frontend.default.svc.80.mesh`. The complex service version then will be slightly more readable `my-complex-service.default.svc.80.mesh`

All three of these should refer to the same VIP. We can have a `kumactl` command that enumerates the services and ServiceInsights and expose
the DNS names associated with it.
