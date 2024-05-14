# MeshExternalService API

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/6330

## Context and Problem Statement

The current implementation of `ExternalService` has significant limitations and warrants a complete redesign.

* Lack of support for * subdomains
* Lack of support for thick clients (think Kafka or C* where discovery is done as part of the client).
* Need for pluggability (for example lambda support or equivalent)
* Working well with gateways (what if your ExternalService relies on SNI/host header to route through a gateway)
* Applying policies with and without egress


## Considered Options

* Creating MeshExternalService + HostnameGenerator
* Creating MeshExternalService
* Creating MeshExternalService and using HostnameGenerator for InternalVIP

## Decision Outcome

Chosen option: "Creating MeshExternalService".

### Positive Consequences

* Clearer structure.
* Enhanced functionalities.
* Better policy matching capabilities.
* No need for a HostnameGenerator

### Negative Consequences

* Whole new resource with a lot of new code.
* Temporarily increased complexity of a product until the migration is done.
* One external service might have many internal VIPs

## Pros and Cons of the Options

### Creating MeshExternalService

We're introducing a new object called `MeshExternalService`. This object will be a namespace resource and will only be creatable within the `kuma-system` namespace. It can be created within a specific zone, but in this case, it will only be accessible within the zone where it was created.

```yaml
kind: MeshExternalService
metadata:
  name: example
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
    kuma.io/zone: east-1
spec:
  match:
  - type: InternalVIP # Kuma will generate a domain
    value: example.ext.svc.local
    port: 80
    protocol: http
  - type: Domain # Existing domain
    value: httpbin.com
    port: 80
    protocol: http
  - type: CIDR
    value: 10.1.1.0/24
    port: 80
    protocol: http
  - type: IP
    value: 192.168.0.1
    port: 80
    protocol: http
  destination:
    type: Regular # Regular|Passthrough|Extension
    extension:
      type: Lambda 
      config: # type JSON
        arn: arn:aws:lambda:us-west-2:123456789012:function:my-function
    endpoints:
    - address: 1.1.1.1
      port: 12345
    - address: example.com
    - address: unix://....
    tls:
      enabled: true
      version:
        min: TLSv1_2 # or TLS_AUTO, TLSv1_0, TLSv1_1, TLSv1_2, TLSv1_3
        max: TLSv1_3 # or TLS_AUTO, TLSv1_0, TLSv1_1, TLSv1_2, TLSv1_3
      allowRenegotiation: false
      verification:
        skip: true # if this is true then subjectAltNames don't take effect
        subjectAltNames: # if subjectAltNames is not defined then take domain or ips
          - example.com
          - "spiffe://example.local/ns/local"
      caCert: 
        inline: 123
      clientCert:
        secret: 123
      clientKey:
        secret: 123
status: 
  vips:
  - ip: 242.0.0.1
    type: Kuma
    hostname: example.ext.svc.local
```
* **spec**:
  * **match**: defines a list of internalVIPs/domains/CIDR/IPs that should be routed through the sidecar
    * **type**: type of the match, one of `InternalVIP`, `Domain`, `CIDR` and `IP` are available
      * `InternalVIP`: allocates a VIP for a domain provided in the `value` field
      * `Domain`: handles traffic to the specified domain
      * `CIDR`: handles traffic to specified addresses range
      * `IP` handles the traffic to specified IP
    * **value**: depends of the type can be an existing domain, new domain name, CIDR or IP
    * **port**: defines a port to which a user does requests
    * **protocol**: defines a protocol of the communication. Possible values:
      * `tls`: should be used when TLS traffic is originated by the client application
      * `tcp`: WARNING: shouldn't be used when match has only domains. On the TCP level we are not able to disinguish domain, in this case it is going to hijack whole traffic on this port.
      * `grpc`
      * `http`
      * `http2`
  * **destination**: defines where matched requests should be routed
    * **type**: defines what kind of destination it is, one of `Regular`, `Passthrough`, or `Extension`, (Default: `Passthrough`)
      * `Regular`: allows creating a set of destination endpoints and `TLS` configuration, when defined section `extension` is not available.
      * `Passthrough`: traffic just passes a proxy without any modifications to the original destination, only available when defined sections `endpoints`, `tls` and `extension` are not defined.
      * `Extension`: allows specifying a custom plugin for example, user can create a plugin which support AWS Lambda, when defined sections `endpoints` and `tls` are not available.
    * **extension**: struct for a plugin configuration
      * **type**: defines what kind of plugin to use, it's a string type so any new plugins should works.
      * **config**: json map that is mapped to configuration provided in the type.
    * **endpoints**: defines a list of endpoints
      * **address**: defines an address to which a user want to send a request. Is possible to provide `domain`, `ip` and `unix` sockets
      * **port**: defines a port of a destination.
    * **tls**: provides a TLS configuration when proxy is resposible for a TLS origination
      * **enabled**: defines if proxy should originate TLS.
      * **version**: section for providing version specification.
        * **min**: defines minmum supported version. One of `TLS_AUTO`, `TLSv1_0`, `TLSv1_1`, `TLSv1_2`, `TLSv1_3`
        * **max**: defines maximum supported version. One of `TLS_AUTO`, `TLSv1_0`, `TLSv1_1`, `TLSv1_2`, `TLSv1_3`
      * **allowRenegotiation**: defines if TLS sessions will allow renegotiation.
      * **verification**: section for providing TLS verification details.
        * **skip**: defines if proxy should skip SAN verification. Default `false`.
        * **subjectAltNames**: list of names to verify in the certificate.
      * **caCert**: defines a certificate of CA.
        * one of `inline`, `inlineString` or `secret`.
      * **clientCert**: defines a certificate of a client.
        * one of `inline`, `inlineString` or `secret`.
      * **clientKey**: defines a client private key.
        * one of `inline`, `inlineString` or `secret`.
* **status**: status of an object managed by Kuma control-plane
  * **vips**: list of allocated VIPs
    * **ip**: allocated IP for a provided domain with `InternalVIP` type in a match section
    * **type**: provides information about the way IP was provided
    * **hostname**: provides a domain with `InternalVIP` type in a match section for which IP was allocated. In case of many entries it helps corelating entries.

#### Cluster name

Currently, we name clusters after their respective policies. However, when dealing with multiple protocols, we find it necessary to include the port number in the cluster name.
Example:

```yaml
kind: MeshExternalService
metadata:
  name: httpbin
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
    kuma.io/zone: east-1
spec:
  match:
  - type: Domain
    value: httpbin.com
    port: 80
    protocol: http2
  - type: Domain
    value: httpbin.com
    port: 443
    protocol: tls
 destination:
   type: Passthrough
```

In this case, for port `80`, we require HTTP/2 configuration in the cluster settings. This necessitates having a separate configuration for each cluster.

clusterName: `{policyName}_{port}`, so for the above policy name would looks like: `httpbin_80` and `httpbin_443`

#### Extensability

Provided model allows creating separate plugins which can integrate with different cloud providers or components. The user can create it's own custom configuration and later in the code create XDS configuration based on it. 

Example:

```yaml
kind: MeshExternalService
metadata:
  name: example
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
    kuma.io/zone: east-1
spec:
  match:
  - type: InternalVIP
    value: example.ext.svc.local
    port: 80
    protocol: http
 destination:
   type: Extension
   extension:
     type: Lambda 
     config:
       arn: arn:aws:lambda:us-west-2:123456789012:function:my-function
```

#### IP intersections 

It's possible for users to create a policy where the IP ranges intersect with each other. We cannot validate this during policy creation, but during configuration generation, we are able to check if IPs intersect. In such cases, we can log a message stating:

> external-service1 and external-service2 have overlapping IPs X.X.X.X, which can disrupt your traffic.

#### Domain generation

The `ExternalService` resource automatically allocated an internal IP each real domain (e.g. `httpbin.com`). While this approach offered benefits such as avoiding the allocation of a listener for a specific port listening on `0.0.0.0`, it obscured the existing domain behind a custom IP. With the introduction of `MeshExternalService`, we expose to users a type called `InternalVIP`, enabling the creation of domains for specific external services.

```yaml
kind: MeshExternalService
metadata:
  name: mongo
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  match:
  - type: InternalVIP # Kuma will generate a domain
    value: mongo.ext.svc.local
    port: 27017
    protocol: tcp
 destination:
   type: Regular # Regular|Passthrough|Extension
   endpoints:
   - address: 10.0.0.1
     port: 27017
   - address: 10.0.0.2
     port: 27017
status: 
  vips:
  - ip: 242.0.0.1
    type: Kuma
    hostname: mongo.ext.svc.local
```

In this scenario, we generate a custom domain `mongo.ext.svc.local` and allocate the address `242.0.0.1` to it.
For the purpose of `MeshExternalServices` we are going to take a completely new CIDR by default `242.0.0.0/8`. This CIDR can be configured by the user.

#### Wildcard domain

Wildcard domains are supported only when `destination.type: Passthrough`, which is a default one.

```yaml
kind: MeshExternalService
metadata:
  name: kafka
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  match: 
  - type: Domain
    value: *.eu-west-3.aws.cloud 
    port: 9092
    protocol: tls
```

### Other options

#### Creating MeshExternalService + HostnameGenerator

```yaml
kind: MeshExternalService
metadata:
  name: myservice
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
spec:
  domains:
  - *.myservice.svc.local
  - httpbin.com
  addresses:
  - 10.1.1.0/24
  - 192.168.0.1
  ports:
  - port: 443
    targetPort: 8443
    protocol: tcp
  type: Regular # Regular|Passthrough|Extension
  extension:
    type: Lambda 
    config: # type JSON
      arn: arn:aws:lambda:us-west-2:123456789012:function:my-function
  destinations:
  - address: 1.1.1.1
    port: 12345
  - address: httpbin.com
  - address: unix://....
  tls:
    version:
      min: TLS_12 # or TLS_13
      max: TLS_12 # setting min=max means we require specific version
    allowRenegotiation: false
    verification:
      skip: true # if this is true then subjectAltNames don't take effect
      subjectAltNames: # if subjectAltNames is not defined then take domains
        - httpbin.com
        - "spiffe://httpbin.com"
    caCert: 
      inline: 123
    clientCert:
      secret: 123
    clientKey:
      secret: 123
status: # managed by CP. Not shared cross zone, but synced to global
  addresses:
  - hostname: myservice.svc.meshext.local
    status: Available # | NotAvailable
    origin: 
      kind: HostGenerator
      name: "k8s-zone-generator"
    reason: "not available because of the clash with ..."
  vips:
  - ip: <kuma VIP>
    type: Kuma | # External
```

* **spec**:
  * **domains**: defines a list of existing domains that are going to be routed through proxy
  * **ports**: defines a list of ports and protocols
    * **port**: defines a port to which a user does requests
    * **protocol**: defines a protocol of the communication. Possible values:
      * `tls`: should be used when TLS traffic is originated by the client application
      * `tcp`: WARNING: shouldn't be used when match has only domains. On the TCP level we are not able to disinguish domain, in this case it is going to hijack whole traffic on this port.
      * `grpc`
      * `http`
      * `http2`
    * **targetPort**: defines a target port to which traffic should be sent.
  * **type**: defines what kind of destination is it, one of `Regular`, `Passthrough`, or `Extension`, (Default: `Passthrough`)
      * `Regular`: allows creating a set of destination endpoints and `TLS` configuration, when defined section `extension` is not available.
      * `Passthrough`: traffic just passes a proxy without any modifications to the original destination, when defined sections `endpoints`, `tls` and `extension` are not available.
      * `Extension`: allows specifying a custom plugin for example, user can create a plugin which support AWS Lambda, when defined sections `endpoints` and `tls` are not available.
    * **extension**: struct for a plugin configuration
      * **type**: defines what kind of plugin to use, it's a string type so any new plugins should works.
      * **config**: json map that is mapped to configuration provided in the type. 
  * **destinations**: defines a list of domains, unix sockets or ips:
    * **address**: defines an address to which a user want to send a request. Is possible to provide `domain`, `ip` and `unix` sockets
    * **port**: defines a port of a destination.
  * **tls**: provides a TLS configuration when proxy is resposible for a TLS origination
    * **enabled**: defines if proxy should originate TLS.
    * **version**: section for providing version specification.
      * **min**: defines minmum supported version. One of `TLS_AUTO`, `TLSv1_0`, `TLSv1_1`, `TLSv1_2`, `TLSv1_3`
      * **max**: defines maximum supported version. One of `TLS_AUTO`, `TLSv1_0`, `TLSv1_1`, `TLSv1_2`, `TLSv1_3`
    * **allowRenegotiation**: defines if TLS sessions will allow renegotiation.
    * **verification**: section for providing TLS verification details.
      * **skip**: defines if proxy should skip SAN verification. Default `false`.
      * **subjectAltNames**: list of names to verify in the certificate.
    * **caCert**: defines a certificate of CA.
      * one of `inline`, `inlineString` or `secret`.
    * **clientCert**: defines a certificate of a client.
      * one of `inline`, `inlineString` or `secret`.
    * **clientKey**: defines a client private key.
      * one of `inline`, `inlineString` or `secret`.
* **status**: status of an object managed by Kuma control-plane
  * **addresses**:
    * **hostname**: domain generated by HostnameGenerator
    * **status**: if a domain is already assigned, possible values `Available` | `NotAvailable`
    * **origin**: provided an information how the domain was generated
      * **kind**: kind of resource responsible for hostname generation
      * **name**: name of the resource
    * **reason**: set when there was error when generating a hostname
  * **vips**: list of allocated VIPs
    * **ip**: allocated IP for a provided domain with `InternalVIP` type in a match section
    * **type**: provides information about the way IP was provided
    * **hostname**: provides a domain with `InternalVIP` type in a match section for which IP was allocated. In case of many entries it helps corelating entries.

##### Hostname and VIP generation

Currently, Kuma allocates a domain for an ExternalService based on the kuma.io/service label and the real destination domain. However, in MeshExternalService, we aim to discontinue this practice. Instead, we propose utilizing the HostGenerator for domain generation. To facilitate this change, we intend to introduce a new selector called meshExternalServiceSelector, enabling users to specify particular MeshExternalServices that necessitate custom domains. Only MeshOperator should create a HostGenerator policy and when a user needs a specific domain for a MeshExternalService, the user needs to reach out to the MeshOperator. HostGenerator can be created only on GlobalCP and later is synchronized to all zones. More about HostGenerator in [MADR-046](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/046-meshservice-hostname-vips.md?plain=1#L58).

```yaml
kind: HostnameGenerator
metadata:
  name: k8s-zone-hostnames
spec:
  targetRef:
    kind: MeshExternalService
    tags:
      my.label.io/service: my-service 
  template: {{ name }}.svc.meshext.local
```

For example, given this `MeshExternalService`

```yaml
kind: MeshExternalService
metadata:
  name: mydomain
  namespace: kuma-system
  labels:
    my.label.io/service: my-service 
    kuma.io/mesh: default
    kuma.io/zone: east-1
    kuma.io/origin: zone
spec:
  destinations:
  - address: 192.168.0.1
    port: 9090 
...
```

We would generate such hostname `mydomain.svc.meshext.local`

Possible template functions/keys are:

`{{ name }}` - name of the `MeshExternalService`
`{{ label "x" }}` - value of label x.

If the template cannot be resolved (label is missing), the hostname won't be generated with a NotAvailable status and an appropriate reason.

#### Positive Consequences

* Simpler model, but not as clear
* One domain per external service 

#### Negative Consequences

* Using HostnameGenerator might cause unexpected domains to be created, when match is too wide

#### Creating MeshExternalService and using HostnameGenerator for InternalVIP

This option is similar to the chosen one, but instead of explicitly providing a domain in the policy, users would need to create a `HostnameGenerator` to generate a domain.

The API model does not include a `value` field for the `InternalVIP` type. The domain name is generated by `HostnameGenerator` and later provided under a status section.

```yaml
kind: MeshExternalService
metadata:
  name: example
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
    kuma.io/zone: east-1
spec:
  match:
  - type: InternalVIP # Kuma will generate a domain based on HostnameGenerator
    port: 80
    protocol: http
...
status:
  addresses:
  - hostname: httpbin.svc.meshext.local
    status: Available # | NotAvailable
    origin: 
      kind: HostGenerator
      name: "k8s-zone-generator"
    reason: "not available because of the clash with ..."
  vips:
  - ip: <kuma VIP>
    type: Kuma | # External
```

#### Positive Consequences

* Separation between hostname generation and possible better security (but not a huge problem when only MeshOperator is allowed to create the policy)

#### Negative Consequences

* Less explicit
