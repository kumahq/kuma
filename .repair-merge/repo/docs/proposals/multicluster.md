# Multicluster Control plane proposal

## Context

Kuma supports both Kubernetes and Universal mode to account for the two most popular workload scheduling mechanisms. Allowing
both to co-exist alongside each-other and expose a single configuration interface is a feature that is on Kuma's roadmap
for a long time. Even more, this proposal suggests how to be able to run multiple Kuma Control Planes in parallel, regardless
of their type - Kubernetes or Universal.

## Requirements

* Be able to leverage service connectivity for L4-L7 services cross clusters of **any** type, size and geographical location.
* High availability - downtime of one cluster should not break the mesh.
* Scalability - support hundreds of thousands Dataplanes assuming they are segmented into clusters.

## Quick Start

Kuma Multicluster introduces a centralised mesh control of mutli-cluster deployments, unifying the management of Kubernetes
and Universal Kuma Control planes. In this mode, the Kuma control plane is split into a central Global entity, and per-cluster
Remote Control planes, which replace the `kuma-cp` as used in Single Cluster mode.

### Installation

The installation process consists of deploying a single Kuma Global and a Kuma Remote Control Planes per cluster.

We do recommend running the Kuma Global on a dedicated Linux VM (Ubuntu, CentOS or RedHat preferred). The installation procedure is
as simple as:

```
global > curl -L https://kuma.io/global.sh | sh -
```

This will deploy kuma as a systemd service and run it accordingly. It will also deploy `kumactl`.

Enabling Kuma on a Kubernetes cluster and adding it to the global, assuming kubectl is pointing to the cluster:
```
cluster > curl -L https://kuma.io/k8s_remote.sh | sh -

The Kuma Remote CP was installed in cluster my-k8s-cluster with token qt57zu.wuvqh64un13trr7x 

global > kumactl join --remote qt57zu.wuvqh64un13trr7x

Cluster my-k8s-cluster at IP was added to the global

```

Similarly, joining a Universal cluster:
```
cluster > curl -L https://kuma.io/uni_remote.sh | sh -

The Kuma Remote CP was installed with token qt57zu.wuvqh64un13trr7x 

global > kumactl join --remote qt57zu.wuvqh64un13trr7x
Cluster at IP 1.1.1.1 was added to the global.
```

As Kuma Remote exposes the same port and API to have remote Dataplane management in Universal, adding new workloads
follows the standard procedure.

### Consuming services

Let's look at an example where a service `web` is accessing a service `backend` at port `9000`, and look at the different scenarios

1. Both services are in the same namespace `dev-1` in the same Kubernetes cluster `dev-cluster-1`. No changes or additional configurations
needed, `web` can refer to `http://backend:9000`.

2. Service `web` is in a namespace `dev-1` and `backend` in the namespace `dev-2` still on the same Kubernetes cluster `dev-cluster-1`. As Kubernetes
Service names and resolving suggests, `web` should call `http://backend.dev-2:9000`.

3. Both services `web` and `backend` are in the namespace `dev-1` on cluster `dev-cluster-1` and now there is a second replica of `backend` deployed
in namespace `dev-1` in another Kubernetes cluster `dev-cluster-2`. Since now Kuma is in control of all the intra and inter-cluster communications
this second replica is immediately available as if it was deployed locally. To enable the local and remote versions to act interchangeably,
the `web` application has to use `http://backend.kuma:9000`. This notation can actually be used to access any service across the multicluster
deployment once Kuma Multicluster is enabled.

4. Given the setup in (3), adding an instance of `backend` in a Universal cluster `dev-cluster-3` follows the standard deployment procedures for Universal. However
it is highly recommended that an additional tag `k8s/namespace` denoting the target namespace is added in the Dataplane resource.
```yaml
type: Dataplane
mesh: web-app
name: backend
networking:
  address: 192.168.0.1
  inbound:
  - port: 9000
    servicePort: 6379
    tags:
      service: backend
      k8s/namespace: dev-1
```

5. Service `web` and `backend` are deployed in namespace `dev-1` on Kubernetes cluster `dev-cluster-1`, replicas of `backend` are 
available in namespace `dev-1` on Kubernetes cluster `dev-cluster-2` and in Universal cluster `dev-cluster-3`.
In this scenario we add another service in `dev-cluster-1` namespace `dev-2`, which has the same name `backend`, but implements
different API, not targeted to be consumed by the `web` application in `dev-1`. To distinguish between the two services with colliding
names, the user can choose to:
 a) assign the colliding services to different meshes
 b) if (a) is not an option, use a traffic policy like this:
```yaml
type: TrafficPermission
name: web-app-dev-1
mesh: web-app
sources:
  - match:
      service: web
      k8s/namespace: dev-1
destinations:
  - match:
      service: backend
      k8s/namespace: dev-1
```

Note that the tag `k8s/namespace` is automatically generated for the Kubernetes Dataplanes, but if needed, should be manually
set in the Universal deployments, as specified in (4).

In Kuma Multicluster (1) and (2) are supporte for backward compatibiliyt, however the recommendation is to migrate the service
request URLs to use the `.kuma` DNS zone, and ensure smooth cross-cluster experience.  

## Architecture

The model decouples managing dataplanes in a clusters and managing policies for the whole mesh, with a top-level Kuma Global Control Plane that will handle
the user-facing API and GUI requests. Each Cluster (Kubernetes or Universal) runs Kuma Remote. The described solution
leverages the Universal configuration model to the whole multi-cluster deployment, regardless of the actual type of
the clusters managed with the Global Contol Plane.

The high-level architecture is depicted at the picture below.

```
                                                 REST API + GUI
                                                         +
                                                         |
                                                         |
                                                         |
                                                         |
                                            +------------v--------------+
                                            |                           |
                                            | Kuma Global Control Plane |
                                            |                           |
                                            +------------+--------------+
                                                         |
                                                         |
          +----------------------+-----------------------+----------------------+-----------------------+
          |                      |                       |                      |                       |
          |                      |                       |                      |                       |
+---------v----------+ +---------v----------+  +---------v----------+ +---------v----------+  +---------v----------+
|       +----------+ | |       +----------+ |  |       +----------+ | |       +----------+ |  |       +----------+ |
| ⎈ K8s |Kuma Remot| | | ⚙ Uni |Kuma Remot| |  | ⚙ Uni |Kuma Remot| | | ⎈ K8s |Kuma Remot| |  | ⎈ K8s |Kuma Remot| |
|       +----------+ | |       +----------+ |  |       +----------+ | |       +----------+ |  |       +----------+ |
|                    | |                    |  |                    | |                    |  |                    |
| +----------------+ | | +----------------+ |  | +----------------+ | | +----------------+ |  | +----------------+ |
| |     Service    | | | |     Service    | |  | |     Service    | | | |     Service    | |  | |     Service    | |
| +-------+--------+ | | +-------+--------+ |  | +-------+--------+ | | +-------+--------+ |  | +-------+--------+ |
| | Envoy | KumaDP | | | | Envoy | KumaDP | |  | | Envoy | KumaDP | | | | Envoy | KumaDP | |  | | Envoy | KumaDP | |
| +-------+--------+ | | +-------+--------+ |  | +-------+--------+ | | +-------+--------+ |  | +-------+--------+ |
|                    | |                    |  |                    | |                    |  |                    |
| +----------------+ | | +----------------+ |  | +----------------+ | | +----------------+ |  | +----------------+ |
| |     Service    | | | |     Service    | |  | |     Service    | | | |     Service    | |  | |     Service    | |
| +-------+--------+ | | +-------+--------+ |  | +-------+--------+ | | +-------+--------+ |  | +-------+--------+ |
| | Envoy | KumaDP | | | | Envoy | KumaDP | |  | | Envoy | KumaDP | | | | Envoy | KumaDP | |  | | Envoy | KumaDP | |
| +----------------+ | | +----------------+ |  | +----------------+ | | +----------------+ |  | +----------------+ |
+--------------------+ +--------------------+  +--------------------+ +--------------------+  +--------------------+
```

### Kuma Global Control Plane

This is the top-level component that exposes central management point for controlling the unified Kuma deployment. Its functionality can be split into Northbound interface and Southbound operations, as follows: 
 * Northbound interface
    * Exposes the HTTP API configuration endpoints (policies)
    * Exposes the GUI interface
    * Does **not** handle any Dataplane registrations
    * Does **not** accept connections from Dataplanes (XDS/SDS etc.)
 * Southbound operations
    * Periodically push the new resource updates to all registered Remote CPs (policies and ingresses)
    * Poll all registered Remote CPs for their active registered Dataplanes (to have list in GUI)

Communication format between Global and Remote is Kuma Core Resource, therefore Kuma GLobal CP can be deployed either on Universal on Kubernetes.

### Remote CP

The Remote CP is the termination point of the API requests to the Global. Its functionality includes:
 * Register the available dataplanes (manual in Universal and automated in Kubernetes)
 * Receive the resource updates from the upstream (policies and dataplane ingress from other clusters) and stores the copy in their own storage.
 * Respond to polls for dumping the full dataplane available resources to upstream
 * Implements a DNS name resolver to ensure sustainable cross-cluster service name resolving

## Networking

The cross-cluster networking model relies on the ability for each cluster to expose its services at a routable IP address,
which is reachable by the other clusters in the Kuma Multicluster deployment.

```
+--------------------------------------------------------------------------+      +--------------------------------------------+
|                                                                          |      |                                            |
| 1.1.1.1:18080                        1.1.1.1:28080                       |      | 2.2.2.2:18080          2.2.2.2:28080       |
|                                                                          |      |                                            |
| +----------------------------------------------------------------------+ |      | +----------------------------------------+ |
| | Ingress                                                              | |      | | Ingress                                | |
| |                                                                      | |      | |                                        | |
| +----------------------------------------------------------------------+ |      | +----------------------------------------+ |
|                                                                          |      |                                            |
| service-1.default.svc.cluster.local  service-1.example.svc.cluster.local |      | service-1              service-2           |
| 192.168.0.1:8080                     192.168.0.2:8080                    |      | 192.168.0.1:8080       192.168.0.2:8080    |
| +----------------+                   +----------------+                  |      | +----------------+     +----------------+  |
| |   Service 1    |                   |   Service 1    |                  |      | |   Service 1    |     |   Service 2    |  |
| +-------+--------+                   +-------+--------+                  |      | +-------+--------+     +-------+--------+  |
| | Envoy | KumaDP |                   | Envoy | KumaDP |                  |      | | Envoy | KumaDP |     | Envoy | KumaDP |  |
| +-------+--------+                   +-------+--------+                  |      | +-------+--------+     +-------+--------+  |
|                                                                          |      |                                            |
| ⎈ K8s Cluster-1                                                          |      | ⚙ Uni Cluster-2                            |
|                                                                          |      |                                            |
+--------------------------------------------------------------------------+      +--------------------------------------------+

```

### Ingress
The Ingress resource plays a crucial role in the Multicluster Kuma architecture. It will be **automatically** created by the Kuma Remote CP upon the registration of a new Dataplane.
The Ingress resource is not the same as the Gateway, we still want the full fledged Kuma control over the traffic that passes between the clusters.

This Dataplane will have inbounds for all services (not instances) available within the cluster (it has to include every possible tag set).

**Example:** Let's say we've got a cluster `k8s-cluster-1` with `redis` and `elastic` services

```yaml
type: Dataplane
mesh: default
name: ingress-1
networking:
  ingress: true # to decide: we need marker that this is an ingress 
  address: 1.1.1.1
  inbound:
  - port: 18080 # picked automatically 
    tags:
      service: elastic
      cluster: k8s-cluster-1 # cluster name will be provided when starting a Remote CP and used here
  - port: 18081
    tags:
      service: redis
      cluster: k8s-cluster-1
```

The Ingress will be implemented as a dedicated Envoy + KumaDP deployment.
Listeners are bound to an address which will be then used in a load balancer, so the Global and other clusters see the Ingress as one IP address.

**Example:** With 10 clusters, every cluster will have knowledge of their local Dataplanes and 9 Ingresses of other clusters.

**Q:** Why not use Kong?
**A:** Kong is great for exposing mesh to users and managing APIs. For this use case we need more control and tight integration with current Envoy configuration.

### Name resolving

The global service naming in Kuma Multicluster is unified under the `.kuma` zone. These will always resolve to virtual,
non-routable IPs (e.g. 240.0.0.0/4). This an alternative to Kubernetes' ClusterIP, for the purposes of creating outbound
listeners in Envoy's side-car.


#### Kubernetes
In order for the service consumers to have a predictable and controlled IP resolving, the Kuma Remote CP also implements a DNS
resolver. It produces predictable responses, such that will be consistent with the generated oubound listeners for Envoy.
Since Kubernetes 1.11, the de-facto standard DNS resolver is CoreDNS, and it is configured through a ConfigMap. 
The Kuma Remote CP Resolver is added like this:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        forward . 172.16.0.1
        cache 30
        loop
        reload
        loadbalance
    }
    kuma:53 {
        errors
        cache 30
        forward . 10.150.0.1 # the Kuma Remote CP service IP
    }
```

#### Universal
There is no DNS name resolving from the Application side on Universal. It always uses `localhost` in conjunction with the port
of the outbound section that was configured in the Dataplane entity.
In the future, the Kuma Remote CP Resolver can be used to allow for the workloads run in Universal to take advantage of the
`.kuma` DNS zone for globally accessing the named services.

#### Avoiding name clashes

By increasing the number of Namespaces and Clusters in the Kuma Multicluster deployment, the chances are that service name
clashes will be observed. As noted in the QuickStart there are two options:
1) assign the colliding resources to different meshes
2) use traffic permission policies based on the `k8s/namespace` tag

NOTE:
Idea how to migrate from having `service: backend.kuma-demo:1234` in resources to solution with many tags will come in a separate proposal.

## Implementation

### Kuma Remote CP registration
Kuma Global CP will have list of all clusters, their gateways and Remote CPs.

**Example:** Kuma global configuration:
```yaml
clusters:
- remote:
    address: 192.168.0.2 # Load balancer for many Remote CPs in this cluster
  gateway:
    address: 192.168.0.1 # Load balancer for many gateways in this cluster
- remote:
    address: 192.168.2.2 # Load balancer for many Remote CPs in this cluster
  gateway:
    address: 192.168.2.1 # Load balancer for many gateways in this cluster
```

The communication between the upstream API entities and the Kuma Remote CP will be signed with a JWT token, to verify the integrity of its data. Additionally,
it will provide API security and allow only for the Kuma Global Northbound interface to be available for the users of the multicluster Kuma deployment.

### Synchronization
Ideally communication between Global and Remote will be held using XDS.

Given two clusters: `k8s-cluster-1`, `k8s-cluster-2`, the `k8s-cluster-1` needs to know only about Ingress of `k8s-cluster-2`.
This means that XDS can be segmented and does not have to use all the dataplanes from `k8s-cluster-2`.
Even if there are many instances of Ingresses in `k8s-cluster-2`, `k8s-cluster-1` needs to only know about one of them. It will use the gateway load balancer address described in section above.

We need to take into account that both Remote and Global CPs should be deployed with multiple instances.
Synchronization is a job that should be executed only by one instance.
We need leader election either in Globals that will be pushing policies to Remote CPs and polling Dataplanes or in Remote CPs that will be pulling policies from Global and pushing Dataplanes to Remote CPs.

Additionally, we need to block modifying/ignore policies that are created via `kubectl` in local K8S clusters. 

### Secrets
The proposal is that all the secrets get synced from the Global down to the Remote CPs.
This way we ensure that CA is synchronized to Remote CPs and new DPs can receive certificates even when Global is down.

This implies that synchronization process is secured by TLS + Auth.

### Binaries
The Kuma Global, Kuma Remote effectively split the implementation of the existing Kuma-CP. Therefore,
the proposal is to effectively keep `kuma-cp` and add more options/flags to implement a different role. For example `kuma-cp --global` will effectively
disable the newly added Southbound Operations and disable accepting Dataplane Registration on the Northbound Interface. 

### Resource remoteisation
The Kuma Remote CP will leverage the cluster name passed when Remote CP is running to tag all its Dataplanes and Ingress resources as a response to the Global's dataplane
polling requests.

## Migration path
The ability to extend an existing cluster with an additional one should be trivial. The migration path shall be detailed later.
Roughly it includes the following stages:
 * replace the remote Kuma-CP with Kuma Loocal CP
 * deploy the new cluster(s) enabled for Multi-cluster
 * deploy the Kuma Global and register All the Kuma Remote CPs

## Upgrade path
Provide instructions which component should we upgrade first, Globals, Remotes, DPs.

## Disasters

Here are several scenarios what will happen if something goes wrong.

### Global instances are down

* You cannot apply new/modify policies
* You can still apply/start Dataplanes in local clusters and generate certs from the CA.
* All clusters have remote copy of policies - Remote CPs can be restarted.

### Remote CP instances in one of the clusters are down

* You can apply new/modify policies
* You can apply/start Dataplane in other clusters and generate certs from the CA.
* You cannot apply/start Dataplanes in this remote cluster. Also DPs that went down won't be removed from XDS (but they can be excluded with healthchecks)

## Support for flat networking

We could also support flat networking with this architecture. Assuming there is a connection between all Dataplanes from all clusters, we can resign from using Ingress and synchronize full list of dataplanes from other clusters.
However, this won't be scalable, with hundreds of thousands Dataplanes we would have to generate XDS using all of them. 

## MVP
 * Global + Remote CPs
 * `.kuma` zone support for naming services
 * Exchanging policies over XDS
 * Fix `service: backend.kuma-demo:1234` intro more structured clusters
 * Universal
    * Kuma DP + separate
 * K8s
    * kumactl install remote (deploys remote cp and ingress)

### Not in MVP
* Blocking adding policies on remote K8S clusters
* Migration path
* Upgrade path
* Support for flat networking
