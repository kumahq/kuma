# MeshTrafficPermission

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4222

## Context and Problem Statement

[New policy matching MADR](./005-policy-matching.md) introduces a new approach how Kuma applies configuration to proxies.
Rolling out strategy implies creating a new set of policies that satisfy a new policy matching.
Current MADR aims to define a MeshTrafficPermission.

## Considered Options

* Create a MeshTrafficPermission policy

## Decision Outcome

Chosen option: create a MeshTrafficPermission policy

### Overview

MeshTrafficPermission allows blocking unwanted incoming traffic for the service.
There are might be several reasons why it could be useful to block unwanted traffic:
* following the [principle of least privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege)
* cut off access for clients that shouldn't use your service (legacy clients, clients from another zone, etc.)

### Specification

Top-level targetRef can have all available kinds:

```yaml
targetRef:
  kind: Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshGatewayRoute|MeshHTTPRoute
  name: ...
```

MeshTrafficPermission is an inbound policy that's why it has a "from" array:

```yaml
from:
  - targetRef:
      kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
      name: ...
```

Configuration for MeshTrafficPermission has a single field "action" with 2 values "ALLOW" or "DENY".
So the overall specification looks like this:

```yaml
type: MeshTrafficPermission
mesh: default
spec:
  targetRef:
    kind: Mesh|MeshSubset|MeshService|MeshServiceSubset|MeshGatewayRoute|MeshHTTPRoute
    name: ...
  from:
    - targetRef:
        kind: Mesh|MeshSubset|MeshService|MeshServiceSubset
        name: ...
      conf:
        action: ALLOW|DENY
```

We don't want to mix mTLS configs (like permissive mTLS or TLS version) with MeshTrafficPermission. 
The main reason is that mTLS settings are applied for the entire inbound, while permissions could be set for a group of clients. 

### Envoy configuration

Envoy has [RBAC Network Filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/network_filters/rbac_filter) 
that allows configuring permissions for the list of principals.

**Principal** is an entity that can be authenticated.<br>
**Permission** is access granted to the principal.  

Each Kuma service has a generated certificate with URIs that identify service name and tags.
For example "backend" service in "us-east" zone of version v1 can have the following list of URIs:
* spiffe://mesh-1/backend
* kuma://kuma.io/zone/us-east
* kuma://version/v1

If we want to grant any permissions to these instances we can create the following Network RBAC:

```yaml
action: ALLOW
policies:
  "backend-us-east-v1-allow-all":
    permissions:
      - any: true
    principals:
      - and_ids:
          ids:
            - authenticated:
                principal_name:
                  exact: "spiffe://mesh-1/backend"
            - authenticated:
                principal_name:
                  exact: "kuma://kuma.io/zone/us-east"
            - authenticated:
                principal_name:
                  exact: "kuma://version/v1"
```

In order request to be allowed, at least 1 policy should be matched. 

#### Converting MeshTrafficPermission to Envoy NetworkRBAC

1. Concatenate "from" arrays from all matched MeshTrafficPermissions 

Input
```yaml
type: MeshTrafficPermission
name: global
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      conf:
        action: DENY
    - targetRef:
        kind: MeshSubset
        tags:
          kuma.io/zone: us-east
      conf:
        action: ALLOW
---
type: MeshTrafficPermission
name: backend
spec:
  targetRef:
    kind: MeshService
    name: backend
  from:
    - targetRef:
        kind: MeshSubset
        tags:
          env: dev
      conf:
        action: DENY
```

Concatenated "from" array
```yaml
from:
  - targetRef:
      kind: Mesh
    conf:
      action: DENY
  - targetRef:
      kind: MeshSubset
      tags:
        kuma.io/zone: us-east
    conf:
      action: ALLOW
  - targetRef:
      kind: MeshSubset
      tags:
        env: dev
    conf:
      action: DENY
```

2. Create a full rule-based view (with negations) 

Regular [rule-based view](005-policy-matching.md#rule-based-view) doesn't work well for NetworkRBAC generation,
because rule-based view relies on the order, but NetworkRBAC doesn't. 

In this example, if a request matches "rule 2" this means it doesn't match "rule 1":
```yaml
rules:
  - targetRef: # rule 1
      kind: MeshSubset
      tags:
        kuma.io/zone: us-east
        env: dev
    conf:
      action: DENY
  - targetRef: # rule 2
      kind: MeshSubset
      tags:
        kuma.io/zone: us-east
    conf:
      action: ALLOW
```

So "rule 2" implicitly matches requests from zone "us-east" that are **not** from env "dev".
That's why it's important to have a full rule-based view with negations ("!" means "not"):

```yaml
rules:
  - targetRef:
      kind: MeshSubset
      tags:
        kuma.io/zone: us-east
        env: dev
    conf:
      action: DENY
  - targetRef:
      kind: MeshSubset
      tags:
        kuma.io/zone: us-east
        env: !dev
    conf:
      action: ALLOW
  - targetRef:
      kind: MeshSubset
      tags:
        kuma.io/zone: !us-east
        env: dev
    conf:
      action: DENY 
  - targetRef:
      kind: MeshSubset
      tags:
        kuma.io/zone: !us-east
        env: !dev
    conf:
      action: DENY
```

See [how to generate a full rule-based view with negations](#how-to-generate-a-full-rule-based-view-with-negations)

3. For each "rule" that has "ALLOW" action generate a separate NetworkRBAC policy

```yaml
action: ALLOW
policies:
  "mtp-1":
    permissions:
      - any: true
    principals:
      - and_ids:
          ids:
            - authenticated:
                principal_name:
                  exact: "kuma://kuma.io/zone/us-east"
            - not_id:
                authenticated:
                  principal_name:
                    exact: "kuma://env/dev"
```

4. If MeshTrafficPermission is targeting proxy then NetworkRBAC is placed as a NetworkFilter on the inbound listener.
If MeshTrafficPermission is targeting MeshHTTPRoute then NetworkRBAC is placed as a NetworkFilter on the route.

### MeshTrafficPermission and Kubernetes

MeshTrafficPermission is going to be the first policy with a new policy matching. 
There are a few things worth mentioning regarding this fact. 

#### Mesh as label

Instead of creating a CRD with a "mesh" field, we can use a label to store mesh:

```yaml
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  labels:
    kuma.io/mesh: mesh-1
```

This allows filtering policies by mesh using kumactl.

#### Scope is namespaced

Today all policies are cluster-scoped, but new policies should be namespace scoped.
Policy applies only to the DPPs of the same namespace.
If policy is created in "kuma-system" then it's applied to all namespaces.

### Positive Consequences <!-- optional -->

* MeshTrafficPermission is more flexible and straight-forward than TrafficPermission
* MeshTrafficPermission supports l7 routes

### Negative Consequences <!-- optional -->

Not found

## Links <!-- optional -->

### How to generate a full rule-based view (with negations)

1. Each targetRef could be represented as a list of tags (tag has "key" and "value")
    * targetRef{kind:Mesh} -> []Tag{} (empty list of tags)
    * targetRef{kind:MeshSubset,tags:tags} -> []Tag{tags...}
    * targetRef{kind:MeshService,name:name} -> []Tag{{key:"kuma.io/service","value":name}}
    * targetRef{kind:MeshServiceSubset,name:name,tags:tags} -> []Tag{append(tags, {key:"kuma.io/service","value":name})}
    
Represent each targetRef in "from" array as tags:

```go
allTags := []Tag{}
for _, item := range from {
	allTags = append(allTags, representAsTags(item.TargetRef)...)
}
dedup(allTags) // tag1 == tag2 <=> (tag1.key == tag2.key && tag1.value == tag2.value)
```

2. We have to iterate over all possible combinations with negations. If we have 3 keys:
```go
allTags = []Tag{
    {key: "key1", value: "value1"},
    {key: "key2", value: "value2"},
    {key: "key3", value: "value3"},
}
```

then there are 2^3=8 combinations:

```go
comb1 = []Tag{
    {key: "key1", value: "value1"},
    {key: "key2", value: "value2"},
    {key: "key3", value: "value3"},
}
comb2 = []Tag{
    {key: "key1", value: "!value1"},
    {key: "key2", value: "value2"},
    {key: "key3", value: "value3"},
}
comb3 = []Tag{
    {key: "key1", value: "value1"},
    {key: "key2", value: "!value2"},
    {key: "key3", value: "value3"},
}
comb4 = []Tag{
    {key: "key1", value: "!value1"},
    {key: "key2", value: "!value2"},
    {key: "key3", value: "value3"},
}
comb5 = []Tag{
    {key: "key1", value: "value1"},
    {key: "key2", value: "value2"},
    {key: "key3", value: "!value3"},
}
comb6 = []Tag{
    {key: "key1", value: "!value1"},
    {key: "key2", value: "value2"},
    {key: "key3", value: "!value3"},
}
comb7 = []Tag{
    {key: "key1", value: "value1"},
    {key: "key2", value: "!value2"},
    {key: "key3", value: "!value3"},
}
comb8 = []Tag{
    {key: "key1", value: "!value1"},
    {key: "key2", value: "!value2"},
    {key: "key3", value: "!value3"},
}
```

Each combination uniquely defines a group of client and each client belongs to exactly one group.
It's possible to have combinations with 0 clients like:
```go
zeroClientsComb = []Tag{
    {key: "key1", value: "value1"},
    {key: "key1", value: "value2"},
}
```

3. Determine configuration value (ALLOW or DENY in the context of permissions) for each combination.
It could be done using initial "from" array by checking every combination against "from" items.

Example

We have a "from" array:

```yaml
from:
  - targetRef:
      kind: Mesh
    conf:
      action: ALLOW
  - targetRef: 
      kind: MeshSubset
      tags:
        zone: us-east
    conf:
      action: DENY
  - targetRef:
      kind: MeshSubset
      tags:
        env: dev
    conf:
      action: ALLOW
  - targetRef:
       kind: MeshSubset
       tags:
          env: prod
    conf:
       action: ALLOW
```

1. Create "allTags" array:
```go
allTags = []Tag{
	{key: "zone", value: "us-east"},
	{key: "env", value: "dev"},
	{key: "env", value: "prod"},
	
}
```

2. Create all possible combination with negations:

```go
comb1 = []Tag{
    {key: "zone", value: "us-east"},
    {key: "env", value: "dev"},
    {key: "env", value: "!prod"},
}
comb2 = []Tag{
    {key: "zone", value: "us-east"},
    {key: "env", value: "!dev"},
    {key: "env", value: "prod"},
}
comb3 = []Tag{
    {key: "zone", value: "!us-east"},
    {key: "env", value: "dev"},
    {key: "env", value: "!prod"},
}
comb4 = []Tag{
    {key: "zone", value: "!us-east"},
    {key: "env", value: "!dev"},
    {key: "env", value: "prod"},
}
comb5 = []Tag{
    {key: "zone", value: "us-east"},
    {key: "env", value: "!dev"},
    {key: "env", value: "!prod"},
}
comb6 = []Tag{
    {key: "zone", value: "!us-east"},
    {key: "env", value: "!dev"},
    {key: "env", value: "!prod"},
}
comb7 = []Tag{ // no clients can have env=dev and env=prod at the same time 
    {key: "zone", value: "us-east"},
    {key: "env", value: "dev"},
    {key: "env", value: "prod"},
}
comb8 = []Tag{ // no clients can have env=dev and env=prod at the same time
    {key: "zone", value: "!us-east"},
    {key: "env", value: "dev"},
    {key: "env", value: "prod"},
}
```

![img.png](assets/rules-example.png)

3. Determine configuration for each combination (only number 5-green is DENY):

```yaml
rules:
  - targetRef:
      kind: MeshSubset
      tags:
         zone: us-east
         env: dev
    conf:
      action: ALLOW
  - targetRef:
       kind: MeshSubset
       tags:
          zone: us-east
          env: prod
    conf:
       action: ALLOW
  - targetRef:
       kind: MeshSubset
       tags:
          zone: !us-east
          env: dev
    conf:
       action: ALLOW
  - targetRef:
       kind: MeshSubset
       tags:
          zone: !us-east
          env: prod
    conf:
       action: ALLOW
  - targetRef:
       kind: MeshSubset
       tags:
          zone: us-east
          env: !prod
          env: !dev
    conf:
       action: DENY
  - targetRef:
       kind: MeshSubset
       tags:
          zone: !us-east
          env: !prod
          env: !dev
    conf:
       action: ALLOW
```