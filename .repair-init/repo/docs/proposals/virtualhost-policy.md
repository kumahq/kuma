# VirtualOutbound policy

## Context

Since Kuma 1.1.4 the dataplane now embeds a CoreDNS instance which forwards all DNS traffic to envoy proxy.
This opens the door to more flexible implementation of vips.
We propose here to create a new policy which will enable users to create arbitrary hostnames and ports that will route to a specific sets of tags.

## New Policy

We will add a new policy type:

```yaml
type: VirtualOutbound
mesh: string # The mesh this policy applies to
name: string #The name of the policy
selectors:
  - match:
    # A dataplane selector that should be taken for selecting dps that should have a hostPort generated.
    # It can either have a value or have "*" in which case a different vip will be generated for each tuple
conf:
    host: string # A go template that uses tags to compile the template
    port: int # A port to expose the virtual host on
    tags:
      # A set of key values that map tags to keys that can be used in the template of host
```

### Example

Simple service with no wildcard replacement:

```yaml
type: VirtualOutbound
mesh: default
name: my-service
selectors:
  - match:
    kuma.io/service: my-service
conf:
  host: httpbin.mesh
  port: 8080
```

Create different hostnames for each version of a service:

```yaml
type: VirtualOutbound
mesh: default
name: my-service-by-version
selectors:
  - match:
    kuma.io/service: "*"
    version: "*"
conf:
  host: "{{version}}.{{service}}.mesh"
  port: 8080
  tags:
    "kuma.io/service": service
    "version": version
```

A policy to enable things to be backward compatible:

```yaml
type: VirtualOutbound
mesh: default
name: default
selectors:
  - match:
    kuma.io/service: "*"
conf:
  host: "{{service}}.mesh"
  port: 80
  tags:
    "kuma.io/service": service
```

A virtual host to get per zone domains

```yaml
type: VirtualOutbound
mesh: default
name: service-by-zone
selectors:
  - match:
    kuma.io/service: "*"
conf:
  host: "{{service}}.{{zone}}.mesh"
  port: 80
  tags:
    "kuma.io/service": service
    "kuma.io/zone": zone
```

### Implementation
The vip management and DNS propagation is already present we will simply add AAAA record to add IPv6 support to the DNS.

We'll add a `hostname` field in the `Dataplane_Networking_Outbound`.
This field will be visible to users in their dataplane insights.

Hostnames that overlaps won't be treated any differently. The user has the endpoint view to troubleshoot.
In the future we might want to add a weight to prioritize a `VirtualOutbound` over another.

VIPs will be allocated per DP and won't need to be persisted anywhere outside of the DP.
This implies that a service may have a different ip on different DPs.

### Steps to implement

- Add hostname to outbound and populate it in `VipOutbounds()`.
- Update `dns_generator` to read the hostname from the `hostname` field in the outbound instead of generating one.
- Add `VirtualOutbound` policy and on reconcile populate the outbounds of the matching dataplanes accordingly.

### Backward compatibility

- Make sure that the new VirtualOutbound use ips not already used by the current vip allocator (execute them sequentially).
- Once it is proven reliable enough we will be able to add a default `virtualOutbound` and remove the auto generated vips.

## Summary

This provides us with two new features:

- Possibility to have vips with specific ports.
- Possibility to have multiple hostnames for a service or a subset of a service. This might be useful for more complex setups or simply for debugging by adding temporary policies.
