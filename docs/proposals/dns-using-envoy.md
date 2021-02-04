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
* Select the tag(s) to use for lower level domain names
* Provide complete isolation between meshes. .i.e. it shouldn't be possible from a mesh to figure out services running in other meshes by querying the DNS
* DNS query should still succeed even if the CP is down

## Configuration model

Here is a full example of the configuration on the mesh.

```yaml
type: Mesh
name: default 
dns:
  port: "53" # [Optional defaults to 53] the port on which the UDP listener should bind in envoy, if set to "transparent" use the same port as for transparent proxying.
  template: "{{ kuma.io/service | replace '_' '.' }}.mesh" # [Optional defaults to `{{ kuma.io/service }}` ] a gotemplate which defines the domain this resolves to.
  tags: ["kuma.io/protocol", "kuma.io/service"] # [Optional defaults to ["kuma.io/service"]] A list of tags to use to build domnains they are joined with "." for example: "http.my-service.mesh" would be a final name.
```

## K8s compatibility

In K8s we will update the pod's [DNSConfig](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-dns-config) to add the sidecar as a nameserver.
We will add an annotation `kuma.io/local-dns-disabled` which will enable users to avoid this injection (This can be useful for troubleshooting).

## Universal compatibility

In universal the user will need to have their /etc/resolv.conf or iptables set to redirect DNS to the envoy sidecar.

## What needs to be done

- Add the configuration to the Mesh resource 
- Add possibility to add UDP listeners to XDS
- Add a new generator which will add the UDP listener with the DNSFilter to the XDS configuration
- Add modification of K8s pod DNSConfig to add the local nameserver when using this feature.

## Backward compatibility

- By default, this feature won't be used until adding DNS 
- The default configuration will make DNS names similar to what is present by default in Kuma
- Users pick the address and port at which the DNS server runs, thus being able to run both setups concurrently

In a future version it will be possible to have a default DNS configuration to replace Kuma DNS altogether.

## Notes

- POC: https://github.com/lahabana/kuma/commit/f27cd908d937d3acfa9f11b3c7e9d566001e89b4
- Original issue: https://github.com/kumahq/kuma/issues/1450

