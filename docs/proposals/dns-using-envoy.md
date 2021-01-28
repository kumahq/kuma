# Use Envoy DNSFilter for DNS resolution

## Context

Currently, Kuma relies on an external DNS to provide VIPs which are then added as services which proxy back to the actual Cluster in Envoy.
For VIPs to be resolved Kuma CP runs a DNS server which maintains this mapping.

This current setup has some limitations:

- Apps need to set their DNS to point to the CP's DNS (or CNIs need to configure DNS to be correct)
- The CP needs to run a DNS and handle the load
- When the CP is down DNS requests will fail. This is usually not the case in a Service Mesh
- DNS configuration is global to all meshes
- DNS is a separate component which makes it less customisable compared to things that end up in XDS
- This DNS is not aware of the origin of requests, therefore isolation of meshes breaks as anyone can resolve any domain

In this proposal we suggest using [Envoy's DNSFilter](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/udp_filters/dns_filter) to replace the CP DNS.
We add a new section in the Mesh Resource to be able to control the configuration of the DNS.

## Out of scope

* UDP transparent proxying
* Handling SRV records. While DNSFilter supports it, it's probably not necessary in v1

## Requirements

* The top-level domain should be configurable (defaults to `mesh`)
* Drop in replacement to Kuma DNS (Default configuration should match Kuma's current configuration)
* ability to set upstream dns resolvers to have a single resolver in the configuration
* Select the tag(s) to use for lower level domain names
* Provide complete isolation between meshes. .i.e. it shouldn't be possible from a mesh to figure out services running in other meshes by querying the DNS
* DNS query should still succeed even if the CP is down

## Configuration model

Here is a full example of the configuration on the mesh.

```yaml
type: Mesh
name: default 
dns:
  domain: "kuma" # [Optional defaults to "mesh"] the TLD to use for services exposed with the DNS
  upstreamResolvers: # [Optional] A list of resolvers to use when we resolving names not in the TLD
    - "8.8.8.8"
    - "4.4.4.4"
  address: "169.254.53.53" # [Optional defaults to 0.0.0.0] the address on which the UDP listener should bind in envoy
  port: "53" # [Optional defaults to 53 or redirectOutbound if transparent proxying is set] the port on which the UDP listener should bind in envoy
  tags: ["kuma.io/protocol", "kuma.io/service"] # [Optional defaults to ["kuma.io/service"]] A list of tags to use to build domnains they are joined with "." for example: "http.my-service.mesh" would be a final name.
```

## What needs to be done

- Add the configuration to the Mesh resource 
- Add possibility to add UDP listeners to XDS
- Add a new generator which will add the UDP listener with the DNSFilter to the XDS configuration

## Backward compatibility

- By default, this feature won't be used until adding DNS 
- The default configuration will make DNS names similar to what is present by default in Kuma
- Users pick the address and port at which the DNS server runs, thus being able to run both setups concurrently

In a future version it will be possible to have a default DNS configuration to replace Kuma DNS altogether.

## Notes

- POC: https://github.com/lahabana/kuma/commit/f27cd908d937d3acfa9f11b3c7e9d566001e89b4
- Original issue: https://github.com/kumahq/kuma/issues/1450

