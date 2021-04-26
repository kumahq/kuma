# VirtualHost policy

## Context

Since Kuma 1.1.4 the dataplane now embeds a CoreDNS instance which forwards all DNS traffic to envoy proxy.
This opens the door to more flexible implementation of vips.
We propose here to create a new policy which will enable users to create arbitrary hostnames and ports that will route to a specific sets of tags.

## New Policy

We will add a new policy type:

```yaml
type: VirtualHost
mesh: string # The mesh this policy applies to
name: string #The name of the policy
tags:
    # A map of tags that should be taken for creating the tag selector. 
    # It can either have a value or have "*" in which case a different vip will be generated for each tuple
host: string # A go template with a special function `tag` which can use tags present in the `tags` section.
port: int # A port to expose the virtual host on
```

It is impossible to create a VirtualHost that hasn't got a tag `kuma,io/service`.

### Example

Simple service with no wildcard replacement:

```yaml
type: VirtualHost
mesh: default
name: my-service
tags:
  kuma.io/service: my-service
host: httpbin.mesh
port: 8080
```

Create different hostnames for each version of a service:

```yaml
type: VirtualHost
mesh: default
name: my-service-by-version
tags:
  kuma.io/service: "*"
  version: "*"
host: "{{tag version}}.{{tag kuma.io/service}}.mesh"
port: 8080
```

A policy to enable things to be backward compatible:

```yaml
type: VirtualHost
mesh: default
name: my-service-by-version
tags:
  kuma.io/service: "*"
host: "{{tag kuma.io/service}}.mesh"
port: 80
```

A virtual host to get per zone domains

```yaml
type: VirtualHost
mesh: default
name: my-service-by-zone
tags:
  kuma.io/service: "*"
  kuma.io/zone: "*"
host: "{{tag kuma.io/service}}.{{kuma.io/zone}}.mesh"
port: 80
```


### Implementation

The vip management and DNS propagation is already present.
The main changes will be:
- Replace `vips` by `virtualHosts` which will be a map of a selector to an ip and a port.
- The selector for vips is: `<policyName>{key1=value1,key2=value2}` (similar to what is done for SNIs in ingress).
- Expand the `vip_allocator` to allocate vips based on tuples of matches defined in the tags section of each policy.

### Backward compatibility

- This new setup will have a different IPAM and will run concurrently to the existing one.
- Add a flag to disable the old way to allocate vips.
- Once it is proven reliable enough we will be able to add a default `virtualHost` and remove the auto generated vips.

## Summary

This provides us with two new features:

- Possibility to have vips with specific ports.
- Possibility to have multiple hostnames for a service or a subset of a service. This might be useful for more complex setups or simply for debugging by adding temporary policies.
