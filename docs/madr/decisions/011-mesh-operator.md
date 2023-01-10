# Kubernetes Operator

* Status: proposed

## Context and Problem Statement

Today, installing Kuma on Kubernetes is a matter of using [kumactl][kumactl] or
[helm][chart] to install the control-plane and any other bootstrapping
componenents needed. One of the disadvantages of this style of management is
is imperative style operations which cannot take full advantage of what
Kubernetes offers, e.g. declarative operations, richer lifecycle
management, and Kubernetes' consistent and standardized API specification. Other
disadvantages for imperative operations is that these commonly require an
external human operator that manually interacts with the tooling to trigger
workflows, which is more costly for organizations and introduces more
opportunities for human error.

The purpose of this document is to propose that we create an operator for Kuma
which will enable full lifecycle management of components according to the
[Kubernetes Operator Pattern][opp] and the [Operator Framework Capability
Model][caps]. Organizations are striving to move more cluster operations to
"autopilot mode" because it helps reduce costs, increase efficiency, and reduce
human error. We can meet the demands of today's Kubernetes cluster operators by
providing declarative APIs for them that follow the Kubernetes standards they
are already familiar with and also build [Kubernetes Controllers][ctrl] that can
enable automating workflows which have been historically manual such as
upgrades, re-balancing of control-plane to sidecar traffic. Building a
Kubernetes operator for Kuma also creates a new surface for automating workflows
which require deep insights into the running cluster, e.g. tracking and
reporting changes in service latency and sidecar resource utilization which may
have otherwise led to failure.

[kumactl]:https://kuma.io/docs/2.0.x/installation/kubernetes/#download-kumactl
[chart]:https://kuma.io/docs/2.0.x/installation/helm/
[caps]:https://operatorframework.io/what/
[opp]:https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
[ctrl]:https://kubernetes.io/docs/concepts/architecture/controller/

## Decision Drivers

* upgrades have historically been one of the greatest operational pain points
  for Kuma mesh, and automation of upgrades is widely desired.
* managing the lifecycle of a multi-zone mesh, particularly across multiple
  clusters, will enable much more painless multi-cluster operations.

## Considered Options

* *Option 1*: Build a Kuma operator to deploy and manage the lifecycle of the
  Kuma control plane and all required subcomponents within a single Kubernetes
  cluster.
* *Option 2*: Build a Kuma operator with multi-cluster/multi-zone lifecycle
  management built in.
