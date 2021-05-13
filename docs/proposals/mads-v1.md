# Monitoring Assignment Discovery (MADS) v1 API

Date: 2021-02-24

## Context

Currently, the [MADS v1alpha1 API](../../api/observability/v1alpha1/mads.proto) defines an xDS API for discovering
services to scrape metrics from (aka Monitoring Assignments). 

As part of our native integration with Prometheus (see [prometheus#7919](https://github.com/prometheus/prometheus/issues/7919)),
we will need to expose this API externally.

### Current Notes
* This API is used internally in `kuma-prometheus-sd` to fetch these targets and write a `file_sd` JSON file, which
  is run as a sidecar in Prometheus
  * This is hard for users to configure and is not so much fun to maintain (see the vendored dependency on Prometheus)
* The API uses xDS v2 (deprecated)
* The API provides little guidance on which fields should be provided
  * Filling in the bare minimum of the v1alpha1 could produce invalid `Targets` (i.e. no `__address__` or `instance` label)

### What we don't need to solve

* Authorizing/ Authentication of this API
  * Should already be authorized by Kuma itself, if Prometheus is in the mesh :slightly_smiling_face:

### What we do need to solve

* Defining a stable API that can be mapped to Prometheus `Targets` and, potentially, could be implemented by others
  looking for a "generic" SD. 

### Alternatives considered

* There are some interesting discussions happening in the CNCF OpenTelemetry project, which defines standards for telemetry data/ systems
    around whether Service Discovery + scrape targets are within their scope
  * It's looking like they are (see [open-telemetry/opentelemetry-specification#1078](https://github.com/open-telemetry/opentelemetry-specification/issues/1078#issuecomment-780737017)),
    though likely not for some time. If they do ( :crossed_fingers: ), I think we should move to implementing their standard
    and advocate for the same in Prometheus.


## Proposed API

### Design Goals
* Make it clear which fields should be provided
* Make fields explicit, move away from the generic Prometheus labels
  * This will make the SD implementation on the Prometheus-side much more clear
  * Currently, there is implicit implementation details shared between the MADS client and server
  * We cannot have this if we are going to expose this API upstream
* Keep the xDS discovery mechanism, but upgrade it to the latest version v3
* Fully support both xDS transport variants (HTTP + gRPC)
* Keep the ability to provide groups of `Targets` with common labels, but map 1-1 to Dataplanes
  * Currently, `MonitoringAssignments` map 1-1 to Dataplanes, which makes some fields duplicative 
* Keep the flexibility to provide arbitrary labels

### proto

```protobuf
syntax = "proto3";

package kuma.observability.v1;

option go_package = "v1";

import "envoy/service/discovery/v3/discovery.proto";
import "google/api/annotations.proto";
import "validate/validate.proto";

// Monitoring Assignment Discovery Service (MADS).
//
// xDS API that is meant for consumption by monitoring systems, e.g. Prometheus.
service MonitoringAssignmentDiscoveryService {
  // GRPC
  rpc DeltaMonitoringAssignments(stream envoy.service.discovery.v3.DeltaDiscoveryRequest)
          returns (stream envoy.service.discovery.v3.DeltaDiscoveryResponse) {}
  rpc StreamMonitoringAssignments(stream envoy.service.discovery.v3.DiscoveryRequest)
          returns (stream envoy.service.discovery.v3.DiscoveryResponse) {}
  // HTTP
  rpc FetchMonitoringAssignments(envoy.service.discovery.v3.DiscoveryRequest)
          returns (envoy.service.discovery.v3.DiscoveryResponse) {
    option (google.api.http).post = "/v3/discovery:monitoringassignment";
    option (google.api.http).body = "*";
  }
}

// MADS resource type.
//
// Describes a group of targets on a single service that need to be monitored.
message MonitoringAssignment {
  // Mesh of the dataplane.
  //
  // E.g., `default`
  string mesh = 2 [ (validate.rules).string = {min_bytes : 1} ];

  // Identifying service being monitored.
  //
  // E.g., `backend`
  string service = 3 [ (validate.rules).string = {min_bytes : 1} ];

  // List of targets that need to be monitored.
  repeated Target targets = 4;

  // Describes a single target that needs to be monitored.
  message Target {
    // Dataplane name.
    //
    // E.g., `backend-01`
    string name = 1 [ (validate.rules).string = {min_bytes : 1} ];

    // Scheme on which to scrape the target.
    //E.g., `http`
    string scheme = 2 [ (validate.rules).string = {min_bytes : 1} ];

    // Address (preferably IP) for the service
    // E.g., `backend.svc` or `10.1.4.32:9090`
    string address = 3 [ (validate.rules).string = {min_bytes : 1} ];

    // Optional path to append to the address for scraping
    //E.g., `/metrics`
    string metrics_path = 4;

    // Arbitrary labels associated with that particular target.
    //
    // E.g.,
    // `{
    //    "commit_hash" : "620506a88",
    //  }`.
    map<string, string> labels = 5;
  }

  // Arbitrary Labels associated with every target in the assignment.
  //
  // E.g., `{"team": "infra"}`.
  map<string, string> labels = 5;
}
```
