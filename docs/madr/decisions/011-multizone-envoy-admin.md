# Multizone Envoy Admin Operations

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/5080

## Context and Problem Statement

Some time ago we introduced functionality that lets user execute Config Dump / Stats / Clusters from the CP.
This is also available through Global CP in multi zone deployment.
The problem is that Zone CP from given zone connects to only one instance of Global CP and admin operations are piped through KDS.
In HA deployment, when you execute the Admin Operation request you don't know which instance you will use.

## Considered Options

* Discovery
  * Config
* Communication
  * API Server
  * Inter-CP Server
  * KDS

## Pros and Cons of the Options

We need to pass the request to a proper instance, for this we need to consider how we want to communicate between CP instances and how they will discover each other.

## Discovery

### Config

We could gather all instances and IPs in one Config object (ConfigMap on Kubernetes).
```yaml
type: Config
name: cp-instances
config: |
  {
    "instances": [
      {
        "address": "192.168.1.100",
        "interCpPort": 5682,
        "id": "<instance_id>",
        "leader": true
      },
      {
        "address": "192.168.1.101",
        "interCpPort": 5682,
        "id": "<instance_id>",
        "leader": false
      },
    ]
  }
```

Every CP instance connects to a leader to inform they are alive.
A leader is the only writer of this config and periodically flushes info of all instances they received the ping from.
Every follower sends the ping message every 10s. If after 30s the leader did not receive the ping, it can remove the instance from the list. 

On Kubernetes, we could have an alternative implementation. Leader can start a component that fetches Pods in system namespace and dumps this to a Config.
However, I'd like to avoid separate implementations for Kubernetes and Universal.

**How to retrieve IP?**
Each CP needs to know its IP address. We can go over network interfaces on the host and take the first which is not localhost.
This comes with a cost that we can pick incorrect interface, therefore we also need a config to override this explicitly.

#### Advantages
* We have a catalog of instances. This is quite useful information that we use later as a building block.
  For example, there was a feature request to see all Zone CP instances in Global CP GUI. We could compute this also on Zone CP and sync to Global CP.

## Communication

### API Server

Global CP instance could pass around request to another CP instance. This comes with a couple of challenges.

#### Authentication
We need to authenticate requests to the API Server. We can propagate authentication header that the original instance received.
This will only work with header based authentication methods, but admin certs are deprecated anyway.

We could generate a separate user token that will only allow this one operation (to reduce potential blast radius of compromised auth data), but issuing user tokens might not be available (see https://github.com/kumahq/kuma/issues/4031).
We could create separate internal token, but this comes with a cost of extra complexity (docs, explaining, managing).

#### TLS
Because of authentication, we need to use TLSed version of the API server.
We most likely cannot verify SAN of other instance cert, because certs usually contains only hostnames and not IPs.
To verify cert against the CA, a user needs to add CA bundle that was used to sign API Server TLS cert.
This further complicates security configuration of Kuma.

#### Advantages
* Less new parts

#### Disadvantages
* Trickier to configure

### Inter-CP Server

We could create a separate server for inter-cp communication.

The separation might be useful for introducing coordination APIs like:
* Better load balancing of clients (e.g. drop some Envoy connections on new CP instance. Same with Zone -> Global CP)
* Instead of computing all jobs on leader, offload some of them to the followers.

#### TLS & Authentication

We can follow the same pattern as we did for CP <-> Envoy Admin.
Generate internal CA, store it as a secret and generate server and client certs from it.
We can verify SANs of certs because separate instance will have separate cert.

#### Schema

```grpc
service InterCpService {
  rpc Ping(PingRequest) returns (PingResponse);

  // ideally reuse message types from kds.proto
  rpc EnvoyXDSConfig(XDSConfigRequest) returns (XDSConfigResponse);
  rpc EnvoyStats(StatsRequest) returns (StatsResponse);
  rpc EnvoyClusters(ClustersRequest) returns (ClustersResponse);
}

message PingRequest {
  string instance_id = 1;
  string address = 2;
  uint32 inter_cp_port = 3;
}

message PingResponse {
  bool is_leader = 1;
}
```

#### Advantages
* Secure configuration is simpler.
* We can pick gRPC as a protocol which is more efficient and have streaming (might be useful in a future)
* Separation between user facing APIs and internal CP APIs.
* Security separation. We don't use user facing tokens.

#### Disadvantages
* An extra port
* More to explain to users? But can be hidden as "advanced topic".

### KDS

Reuse KDS server for inter cp communications

#### Advantages
* Less new ports

#### Disadvantages
* Same problems as API server with TLS & Authentication
* Might a bit awkward to explain that this is sometimes use to Global CP <-> Zone CP, and sometimes it's used for inter cp communciations.

## Decision Outcome

Chosen option: "Config" with "Inter-CP Server", because it seems more secure, easier to use and more future-proof.

## Putting everything together

When executing Envoy Admin operation on Global CP we would:
* Fetch ZoneInsight of the DP's zone we want to inspect and retrieve Global CP instance ID (we have this in ZoneInsight already) of active subscription.
* If we are on the same instance, we can just execute the existing logic.
* If not, we retrieve catalog of CP instances and check the address of other CP instance. Then we execute request to the server of other CP instance, which would execute existing logic.
