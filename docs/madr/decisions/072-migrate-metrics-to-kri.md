# Migration of metrics to KRI format

* Status: accepted

## Context and problem statement

https://raw.githubusercontent.com/kumahq/kuma/refs/heads/master/app/kuma-dp/pkg/dataplane/metrics/merge.go


## Problems


## Solution

#### Pros and Cons of the Option 1

**Pros:**
- 

**Cons:**
- 

## Downsides

1. Each component now has to parse annotations and handle config merging on its own. They'll also need to mount any custom ConfigMap the user specifies. But we can reuse the same code across components.
2. When using custom ConfigMaps on workloads, it won’t be possible to specify settings needed during sidecar injection, such as `kumaDPUser` or `ebpf`. This isn’t a major limitation because these values rarely change. If needed, they can still be set globally in the main ConfigMap or in the control plane configuration. The need to override them per workload is extremely rare. 
