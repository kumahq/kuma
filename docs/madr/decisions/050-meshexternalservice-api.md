# MeshExternalService API

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/6330

## Context and Problem Statement

The current implementation of `ExternalService` has significant limitations and warrants a complete redesign.

* Lack of support for * subdomains
* Lack of support for thick clients (think Kafka or Cassandra where discovery is done as part of the client).
* Need for pluggability (for example lambda support or equivalent)
* Working well with gateways (what if your ExternalService relies on SNI/host header to route through a gateway)
* Applying policies with and without egress

## Considered Options

* Creating MeshExternalService resource and MeshPassthrough policy
* Creating MeshExternalService
* Creating MeshExternalService + HostnameGenerator
* Creating MeshExternalService and using HostnameGenerator for InternalVIP

## Decision Outcome

Chosen option: "Creating MeshExternalService resource and MeshPassthrough policy".

### Positive Consequences

* Clearer structure.
* Explicit showing what is passthough and what is external service
* Enhanced functionalities.
* Better policy matching capabilities.

### Negative Consequences

* 2 new resources
* Temporarily increased complexity of a product until the migration is done.

## Pros and Cons of the Options

### Creating MeshExternalService

We're introducing a new object resource `MeshExternalService`. This object will be a namespace resource and will only be creatable within the `kuma-system` namespace. It can be created within a specific zone, but in this case, it will only be accessible within the zone where it was created.

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
    type: HostnameGenerator # Kuma will generate a domain
    port: 80
    protocol: http
  extension:
    type: Lambda 
    config: # type JSON
      arn: arn:aws:lambda:us-west-2:123456789012:function:my-function
  endpoints: # in the presence of an extension `endpoints` and `tls` are not required anymore, it's up to the extension to validate them independently.
    - address: 1.1.1.1
      port: 12345
    - address: example.com
    - address: unix://....
  tls:
    version:
      min: TLS12 # or TLSAuto, TLS10, TLS11, TLS12, TLS13
      max: TLS13 # or TLSAuto, TLS10, TLS11, TLS12, TLS13
    allowRenegotiation: false
    verification:
      mode: Secured
      subjectAltNames: # if subjectAltNames is not defined then take domain or ips
        - type: Exact
          value: example.com
        - type: Prefix
          value: "spiffe://example.local/ns/local"
      caCert: 
        inline: 123
      clientCert:
        secret: 123
      clientKey:
        secret: 123
status: 
  vip:
    value: 242.0.0.1
    type: Kuma
  addresses:
  - hostname: example.ext.svc.local
    status: NotAvailable
    origin:
      kind: HostnameGenerator
      name: k8s-example-service-hostname
    reason: "addresses are overlapping with ext2"
```
* **spec**:
  * **match**: defines traffic that should be routed through the sidecar
    * **type**: type of the match, only `HostnameGenerator` is available at the moment
      * `HostnameGenerator`: matches hostnames provided by HostnameGenerator
    * **port**: defines a port to which a user does requests
    * **protocol**: defines a protocol of the communication. Possible values:
      * `tcp`
      * `grpc`
      * `http`
      * `http2`
  * **extension**: struct for a plugin configuration, in the presence of an extension `endpoints` and `tls` are not required anymore - it's up to the extension to validate them independently.
    * **type**: defines what kind of plugin to use, it's a string type so any new plugins should work.
    * **config**: json map that is mapped to configuration provided in the type.
  * **endpoints**: defines a list of endpoints.
    * **address**: defines an address to which a user want to send a request. Is possible to provide `domain`, `ip` and `unix` sockets
    * **port**: defines a port of a destination.
  * **tls**: provides a TLS configuration when proxy is resposible for a TLS origination
    * **enabled**: defines if proxy should originate TLS.
    * **version**: section for providing version specification.
      * **min**: defines minmum supported version. One of `TLSAuto`, `TLS10`, `TLS11`, `TLS12`, `TLS13`
      * **max**: defines maximum supported version. One of `TLSAuto`, `TLS10`, `TLS11`, `TLS12`, `TLS13`
    * **allowRenegotiation**: defines if TLS sessions will allow renegotiation.
    * **verification**: section for providing TLS verification details.
      * **mode**: defines if proxy should skip verification, one of `SkipSAN`, `SkipCA`, `Secured`, `SkipALL`. Default `Secured`.
      * **subjectAltNames**: list of names to verify in the certificate.
        * **type**: specify matching type, one of `Exact`, `Prefix`. Default: `Exact`
        * **value**: name to verify.
      * **caCert**: defines a certificate of CA.
        * one of `inline`, `inlineString` or `secret`.
      * **clientCert**: defines a certificate of a client.
        * one of `inline`, `inlineString` or `secret`.
      * **clientKey**: defines a client private key.
        * one of `inline`, `inlineString` or `secret`.
* **status**: status of an object managed by Kuma control-plane
  * **vip**: section for allocated IP
    * **value**: allocated IP for a provided domain with `HostnameGenerator` type in a match section or provided IP
    * **type**: provides information about the way IP was provided
  * **addresses**: section for generated domain
    * **hostname**: generated domain 
    * **status**: indicate if an address is available
    * **origin**: section providing information what generated the vip
      * **kind**: points to kind that generated domain
      * **name**: name of the `HostnameGenerator`
    * **reason**: holds error messages if there are any 

#### Default CA certificate

By default, when TLS is enabled and CA certificate is not set, we are going to use system bundled one. We can achive that by scanning in `kuma-dp` [default paths](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ssl#example-configuration) for each environment and later send that information in the bootstrap request to the control-plane for a configuration generation.

#### Policy matching

`MeshExternalServices` should work the same with policies as `MeshService`. We are going to introduce a new `kind: MeshExternalService`, which allows targeting them with policies.

We will create separate listeners for each `MeshExternalService` that bind to VIP provided by HostnameGenerator. It's easier to do it this way instead of having multiple filter chain matches on `passthrough` listener because we this piece of code already implemented and we don't have to search deep to find which FilterChain is responsible for which External Service.

#### Envoy resources naming

Currently, we name clusters after their respective policies. Since the model includes only one match we can stick to the same pattern but prefix it with `meshexternalservice_`.
Example:

```yaml
kind: MeshExternalService
metadata:
  name: lambda
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
    kuma.io/zone: east-1
spec:
  match:
    type: HostnameGenerator # Kuma will generate a domain
    port: 80
    protocol: http
  extension:
    type: Lambda 
    config: # type JSON
      arn: arn:aws:lambda:us-west-2:123456789012:function:my-function
  endpoints:
    - address: 1.1.1.1
      port: 12345
```

Cluster name: `meshexternalservice_{resourceName}`, so for the resource name above, it would look like: `meshexternalservice_lambda`.

Currently the name is `outbound:{vip_ip}:{vip_port}` - we think that VIP IP does not tell the user too much information so we suggest changing it.
We suggest naming them `meshexternalservice_{resourceName}` because each listener points to the cluster.

#### Extensability

Provided model allows creating separate plugins which can integrate with different cloud providers or components. The user can create it's own custom configuration and later in the code create XDS configuration based on it. 

Example:

```yaml
kind: MeshExternalService
metadata:
  name: lambda-example
  namespace: kuma-system
  labels:
    kuma.io/mesh: default
    kuma.io/zone: east-1
spec:
  match:
    type: HostnameGenerator # Kuma will generate a domain
    port: 80
    protocol: http
  extension:
    type: Lambda 
    config: # type JSON
      arn: arn:aws:lambda:us-west-2:123456789012:function:my-function
```

`endpoints` and `tls` sections in the configuration are extension specific and might not be required.

#### Universal without Transparent Proxy

It should work similar to the current way. In the `Dataplane` object an user needs to provide information about outbounds. The external service is going to be available on this port.

```yaml
type: Dataplane
mesh: default
name: redis-1
networking:
  ...
  outbound:
  - port: 54321
    backendRef:
      kind: MeshExternalService
      name: ext-svc
```

#### Domain generation

Currently, Kuma allocates a domain for an `ExternalService` based on the `kuma.io/service` label and the real destination domain. However, in `MeshExternalService`, we aim to discontinue this practice. Instead, we propose utilizing the `HostnameGenerator` for domain generation. To facilitate this change, we intend to introduce a new kind `MeshExternalService`, enabling users to specify particular `MeshExternalServices` that necessitate custom domains. Only MeshOperator should create a `HostnameGenerator` resource and when a user needs a specific domain for a `MeshExternalService`, the user needs to reach out to the MeshOperator. `HostnameGenerator` can be created only on GlobalCP and later is synchronized to all zones. More about HostnameGenerator in [MADR-046](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/046-meshservice-hostname-vips.md?plain=1#L58).

```yaml
kind: HostnameGenerator
metadata:
  name: k8s-meshext-hostnames
spec:
  targetRef:
    kind: MeshExternalService
    tags:
      my-service.io/access: "true" 
  template: {{ name }}.svc.meshext.local
```

For example, given this `MeshExternalService`

```yaml
kind: MeshExternalService
metadata:
  name: mydomain
  namespace: kuma-system
  labels:
    my-service.io/access: "true" 
    kuma.io/mesh: default
    kuma.io/zone: east-1
    kuma.io/origin: zone
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
  - address: 192.168.0.1
    port: 9090 
...
```

In this scenario, we generate a custom domain `mydomain.svc.meshext.local` and allocate the address `242.0.0.1` to it.
For the purpose of `MeshExternalServices` we are going to take a completely new CIDR by default `242.0.0.0/8`. This CIDR can be configured by the user.

#### Wildcard domain

Wildcard domains are not going to be supported as a part of `MeshExternalService`. We want to introduce a new policy `MeshPassthrough` which is going to be resposible for exposing wildcard and domains.

### Other options

#### Creating MeshExternalService

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
  type: Managed # Managed|Passthrough|Extension
  extension:
    type: Lambda 
    config: # type JSON
      arn: arn:aws:lambda:us-west-2:123456789012:function:my-function
  managed:
    endpoints:
    - address: 1.1.1.1
      port: 12345
    - address: example.com
    - address: unix://....
    tls:
      enabled: true
      version:
        min: TLS12 # or TLSAuto, TLS10, TLS11, TLS12, TLS13
        max: TLS13 # or TLSAuto, TLS10, TLS11, TLS12, TLS13
      allowRenegotiation: false
      verification:
        skipSAN: true # if this is true then subjectAltNames don't take effect
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
  vip:
    ip: 242.0.0.1
    type: Kuma
  addresses:
  - address: example.ext.svc.local
    status: Available
  - address: httpbin.com
    status: Available
  - address: 192.168.0.1
    status: Available
  - address: 10.1.1.0/24
    status: NotAvailable
    reason: "addresses are overlapping with ext2"
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
      * `tls`: should be used when TLS traffic is originated by the client application in the case the `kuma.io/protocol` would be tcp
      * `tcp`: WARNING: shouldn't be used when match has only domains. On the TCP level we are not able to disinguish domain, in this case it is going to hijack whole traffic on this port. We are going to validate configuration and do not apply config when protocol is tcp and `type: Domain`.
      * `grpc`
      * `http`
      * `http2`
  * **type**: defines what kind of destination it is, one of `Managed`, `Passthrough`, or `Extension`, (Default: `Passthrough`)
    * `Managed`: allows creating a set of destination endpoints and `TLS` configuration, when defined section `extension` is not available.
    * `Passthrough`: traffic just passes a proxy without any modifications to the original destination, only available when defined sections `endpoints`, `tls` and `extension` are not defined.
    * `Extension`: allows specifying a custom plugin for example, user can create a plugin which support AWS Lambda, when defined sections `endpoints` and `tls` are not available.
  * **extension**: struct for a plugin configuration, only for a `type: Extension`
    * **type**: defines what kind of plugin to use, it's a string type so any new plugins should works.
    * **config**: json that is mapped to configuration provided in the type.
  * **managed**: defines where matched requests should be routed, only for a `type: Managed`
    * **endpoints**: defines a list of endpoints
      * **address**: defines an address to which a user want to send a request. Is possible to provide `domain`, `ip` and `unix` sockets
      * **port**: defines a port of a destination.
    * **tls**: provides a TLS configuration when proxy is resposible for a TLS origination
      * **enabled**: defines if proxy should originate TLS.
      * **version**: section for providing version specification.
        * **min**: defines minmum supported version. One of `TLSAuto`, `TLS10`, `TLS11`, `TLS12`, `TLS13`. Default: `TLSAuto`
        * **max**: defines maximum supported version. One of `TLSAuto`, `TLS10`, `TLS11`, `TLS12`, `TLS13`. Default: `TLSAuto`
      * **allowRenegotiation**: defines if TLS sessions will allow renegotiation.
      * **verification**: section for providing TLS verification details.
        * **skipSAN**: defines if proxy should skip SAN verification. Default `false`.
        * **subjectAltNames**: list of names to verify in the certificate.
        * **caCert**: defines a certificate of CA.
          * one of `inline`, `inlineString` or `secret`.
        * **clientCert**: defines a certificate of a client.
          * one of `inline`, `inlineString` or `secret`.
        * **clientKey**: defines a client private key.
          * one of `inline`, `inlineString` or `secret`.
* **status**: status of an object managed by Kuma control-plane
  * **vip**: section for allocated IP
    * **ip**: allocated IP for a provided domain with `InternalVIP` type in a match section
    * **type**: provides information about the way IP was provided
  * **addresses**: list of domains and addresses
    * **address**: IP address, CIDR, or domain name user
    * **status**: indicate if an address is available
    * **reason**: holds error messages if there are any 

##### Hostname and VIP generation

Currently, Kuma allocates a domain for an ExternalService based on the kuma.io/service label and the real destination domain. However, in `MeshExternalService`, we aim to discontinue this practice. Instead, we propose utilizing the HostGenerator for domain generation. To facilitate this change, we intend to introduce a new selector called meshExternalServiceSelector, enabling users to specify particular MeshExternalServices that necessitate custom domains. Only MeshOperator should create a HostGenerator policy and when a user needs a specific domain for a MeshExternalService, the user needs to reach out to the MeshOperator. HostGenerator can be created only on GlobalCP and later is synchronized to all zones. More about HostGenerator in [MADR-046](https://github.com/kumahq/kuma/blob/master/docs/madr/decisions/046-meshservice-hostname-vips.md?plain=1#L58).

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
  type: Managed # Managed|Passthrough|Extension
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

* Clearer structure.
* Enhanced functionalities.
* Better policy matching capabilities.
* No need for a HostnameGenerator

#### Negative Consequences

* Whole new resource with a lot of new code.
* Temporarily increased complexity of a product until the migration is done.
* Problem with supporting Universal without transparent proxy, because one MeshExternalService points to many real domains
* Problem with matching and applying policy when type `Passthrough`. We can't apply policies on passthrough because the cluster type is `original_dst` and these configuration options are not available there.

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
  type: Managed # Managed|Passthrough|Extension
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
  * **type**: defines what kind of destination is it, one of `Managed`, `Passthrough`, or `Extension`, (Default: `Passthrough`)
      * `Managed`: allows creating a set of destination endpoints and `TLS` configuration, when defined section `extension` is not available.
      * `Passthrough`: traffic just passes a proxy without any modifications to the original destination, when defined sections `endpoints`, `tls` and `extension` are not available.
      * `Extension`: allows specifying a custom plugin for example, user can create a plugin which support AWS Lambda, when defined sections `endpoints` and `tls` are not available.
    * **extension**: struct for a plugin configuration
      * **type**: defines what kind of plugin to use, it's a string type so any new plugins should works.
      * **config**: json map that is mapped to configuration provided in the type. 
  * **destinations**: defines a list of domains, unix sockets or ips:
    * **address**: defines an address to which a user want to send a request. Is possible to provide `domain`, `ip` and `unix` sockets
    * **port**: defines a port of a destination.
  * **tls**: provides a TLS configuration when proxy is resposible for a TLS origination
    * **enabled**: defines if proxy should originate TLS. If no certs provided uses default system bundled.
    * **version**: section for providing version specification.
      * **min**: defines minmum supported version. One of `TLSAuto`, `TLS10`, `TLS11`, `TLS12`, `TLS13`
      * **max**: defines maximum supported version. One of `TLSAuto`, `TLS10`, `TLS11`, `TLS12`, `TLS13`
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
* Match section looks a bit awkward with a hostname generator

## Common concerns

### Migration

It is vital that we provide easy and safe migration path.
MeshExternalService is quite different from other policies so having shadow mode will result in big Envoy config diffs that would be hard to predict if they would work correctly.

#### Introduce switch on the DP (env variable / label) to disable specific ExternalServices

We can have a switch to disable a list of ExternalServices on the DP level.

The migration path will look like this:
1. Define MeshExternalService "x" (that replaces ExternalService "x")
2. Disable ExternalService "x" on one instance
3. Check that the traffic is working fine
4. Remove ExternalService
5. Remove config from point 2

We can implement support for `kuma.io/ignore` for `ExternalService`. When labels is added, the whole resource is not taken into configuration generation.
