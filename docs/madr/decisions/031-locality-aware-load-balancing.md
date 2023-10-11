# MeshLoadBalancingStrategy - locality aware
* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/2689

## Context and Problem Statement
Right now we have really simple locality aware load balancing, we only care about zones. Dataplane will prefer sending traffic to local zone with fallback to other zones. We would like to add more sophisticated mechanism of locality aware load balancing. We would like to allow user to specify load balancing rules based on tags. For example, big cluster is deployed in multiple AZ’s, and we would like to prefer staying in a single AZ to reduce traffic cost.
Considered Options

## Decision Outcome

## Solution
As stated earlier, we would like to improve our current locality aware load balancing, and make it more flexible for users.
### Use cases
After gathering client's expectations and a couple of brainstorming sessions, we ended up with multiple use cases for locality aware load balancing:

1. Keeping traffic in single zone
2. Load balancing based on priorities within a zone. Example: node → AZ → region
3. Prefer traffic in single zone with fallback to other zones
4. Load balance equally to list of zones
5. When leaving zone, target specific zones first. Example zones physically close to each other
6. Load balance equally to local dpp and other zone
7. Exclude zone from load balancing

### API design
Looking at use cases, it is clear that we have two types of load balancing strategies, inside zone and cross zone. We can easily route to services with proper tags inside the zone, but immediately after we cross zone border and hit Zone Ingress we lose the ability to control locality.
Because of this we have decided that we will split our API in two sections: `inZone` and `crossZone`.

Here is the full example of policy:

```yaml
targetRef:
  kind: Mesh
to:
  targetRef:
    kind: MeshService
    name: backend
  defaults:
    localityAware:
      inZone: ["k8s.io/node", "k8s.io/az"] # (1)
      crossZone: # (2)
        rules: # (3)
          - groups: # (4)
            - ["zone-1", "zone-2"]
            - ["zone-3", "zone-4"]
          - groups:
            - ["zone-5"]
        fallback: any # (5)
```


(1) In `inZone` section, you configure the priorities of tags used for load balancing.

(2) In `crossZone` section, you configure zones priority.

(3) In `rules` section, you configure list of zone groups priorities, first groups will have the highest load balancing priority and last in order will have the lowest priority.

(4) In `groups` section, you configure logical zone groups. Let's assume `zone-1` and `zone-2` are located in the same datacenter, so you want to load balance equally between them. `Zone-5` is located in another datacenter so it should have lower priority, that is why it is in the second groups entry.

(5) At last we should apply some feedback if no dataplanes from previous configuration are available, at this moment we allow `fail` and `any` fallback

### API examples based on use cases

#### Keeping traffic in single zone

![Use case 1](assets/031/use_case_1.png)

```yaml
targetRef:
  kind: Mesh
to:
  targetRef:
    kind: MeshService
    name: backend
  defaults:
    localityAware:
      inZone: []
```

#### Load balancing based on priorities within a zone. Example: node → AZ → region

![Use case 2](assets/031/use_case_2.png)

```yaml
targetRef:
  kind: Mesh
to:
  targetRef:
    kind: MeshService
    name: backend
  defaults:
    localityAware:
      inZone: ["k8s.io/node", "k8s.io/az"]
```

#### Prefer traffic in single zone with fallback to other zones

![Use case 3](assets/031/use_case_3.png)

```yaml
targetRef:
  kind: Mesh
to:
  targetRef:
    kind: MeshService
    name: backend
  defaults:
    localityAware:
        inZone: []
    crossZone:
      fallback: any
```

#### Load balance equally to list of zones

![Use case 4](assets/031/use_case_4.png)

```yaml
targetRef:
  kind: Mesh
to:
  targetRef:
    kind: MeshService
    name: backend
  defaults:
    localityAware:
      inZone: []
    crossZone:
      - groups:
        - ["zone-1", "zone-2"]
```

#### When leaving zone, target specific zones first

![Use case 5](assets/031/use_case_5.png)

```yaml
targetRef:
  kind: Mesh
to:
  targetRef:
    kind: MeshService
    name: backend
  defaults:
    localityAware:
      inZone: []
    crossZone:
      - groups:
        - ["zone-1"]
      - groups:
        - ["zone-3"]
```

#### Exclude single zone from load balancing

![Use case 6](assets/031/use_case_6.png)

```yaml
targetRef:
  kind: Mesh
to:
  targetRef:
    kind: MeshService
    name: backend
  defaults:
    localityAware:
      inZone: []
    crossZone:
      - groups:
        - ["zone-1"]
      fallback: fail
```

TODO What if i have lots of zones and don’t want to write all of them during the incident. I would like to just exclude single zone, do we want to add something like:

```yaml
targetRef:
  kind: Mesh
to:
  targetRef:
    kind: MeshService
    name: backend
  defaults:
    localityAware:
      inZone: []
    crossZone:
      - exclude: ["zone-3"]
```

#### Load balance equally to local and zone

![Use case 7](assets/031/use_case_7.png)

We have the couple of use cases for this, but no config yet:
- You are migrating from universal to k8s, both zones are in the same AZ and I would like to move traffic gradually from one to another
- You have have some zone with testing or preview environment and you would like to handle live traffic on it for testing. Is it a use case for locality aware lb? Or should this be done with MeshHTTPRoute?
The question is if we really need it, or can this be done currently in some other way?


### Implementation

We want to enhance current `MeshLoadBalancingStrategy` so code related to the changes should be placed in the plugin.

#### Matching endpoints
Each instance of the service might be in a different locality: zone, node. Users need to properly tag these dataplanes so we are able to match them. We need to retrieve all tags from all inbounds and, later, based on these tags, match outbounds that are in the same location.

Example:

On Kubernetes, all tags are populated automatically, while on Universal, users need to populate tags themselves:

```yaml
type: Dataplane
mesh: default
name: demo
networking:
  address: 127.0.0.1
  inbound:
    - port: 9999
      servicePort: 10000
      serviceAddress: 127.0.0.1
      tags:
        kuma.io/service: demo
        kuma.io/protocol: tcp
        k8s.io/az: us-east-1
```


Later, based on these tags, the control-plane creates `ClusterLoadBalancing` and puts all tags, for each endpoint, into cluster metadata::

```json
"metadata":{
  "filter_metadata": {
    "envoy.lb": {
      "az/test": "b",
      "kuma.io/protocol": "tcp",
      "k8s.io/az": "us-east-1"
    }
  }
}
```


We can retrieve them from `ClusterLoadBalancing` and match them with the policy. First, we need to retrieve all tags from inbounds so we can determine our location and match it with endpoints.

Pseudo algorithm:

1. Get all inbound tags.
2. Take the matched rules.
3. Check if inbound has the tag defined in the rule. 
4. If yes, iterate over all endpoints of the service and group them in to priority groups, by matching with the specific rule( first inZone, then  cross Zone). If no, get all the endpoints in the zone. 
5. Create `LocalityLbEndpoints` with selected endpoints. 
6. Override `ClusterLoadAssignment` for the dataplane in the `ResourceStore`. 

Dataplanes in the specific location are going to receive only endpoints in the specific locality. For cross zone traffic without egress the control-plane delivers only ingresses of specific zones.

##### Egress
Egress is not as simple as ingress. Currently, we support Locality Aware when atleast one client requires it, so we cannot distinguish between clients. Control-plane needs to send all dataplanes to the egress sidecar because there might be services sending requests to all zones. There are 2 options:
- LoadBalacing based on metadata
- Separate clusters

Load balancing based on the metadata seems to be able to solve an issue but it has limitations. Because, egress knows all endpoints and does routing based on matching metadata the request might not obey the priority of zones. In this case we cannot configure different priorities for different clients.

Separate cluster seems to be the only option that could solve this problem. When there is a different configuration of Load Balancing for the one of dataplane, control-plane sends to egress new a cluster `mesh-name:service-name:hash(tag + value)` with the routing based on the SNI: `service-name{mesh=mesh-name,hash=hash}`
This cluster has a different configuration of priorities and only subset of zones. Dataplanes interested in this Load Balancing are going to have the same configuration of the cluster which sets SNI: `service-name{mesh=mesh-name,hash=hash}` and based in this SNI egress can make a routing decision to route to the specific cluster.

##### Ingress
Ingress should receive all endpoints, because there is no rules for in zone traffic when the request comes from the outside.

##### Serviceless
On kubernetes you can create a deployment without the service, in this case control-planes creates default inbound so tags should be available.
On universal you cannot create a dataplane without inbound or gateway definition.

#### ClusterLoadBalancing cache
Currently, while retrieving endpoints in the plugin, we are getting a pointer to the object in the cache that is shared between dataplanes. This won't work with LocalityAware load balancing because each dataplane might have a different subset of endpoints with different priorities. In the first iteration, we would like to try using it without CLA cache and measure the impact of the policy in the bigger cluster with many nodes. If the impact of not using the cache is significant, we might have to keep this configuration in the cache if there are more services on the same node. The cache might not improve the performance if there are instances of the same service in different localities (nodes, availability zones).


### Other
#### Cross mesh
We cannot configure this policy cross mesh. We are bound to mesh, when you leave mesh you lose all information needed for locality aware load balancing, and you need to specify new policy for each mesh.

Question what if someone is placing each bounded context in a different mesh for isolation and would still like to benefit from locality aware load balancing? Will cross mesh traffic be routed locally? 
