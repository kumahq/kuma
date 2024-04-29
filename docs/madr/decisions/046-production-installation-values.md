# Production installation values

- Status: accepted

## Context and Problem Statement

This document is meant to highlight our suggestions for installing Kuma in a production environment.

The default `values.yaml` we are providing now is for "demo" usage only. It is reported by multiple users that the demo usage `values.yaml` is too open and not secure enough for production usage.  

When a user installs Kuma in a production, care must be taken in multiple areas to ensure the system is secure, reliable, and performant.


## Decision Drivers

- Provide a suggested `values.yaml` to represent a production installation
- The production usage `values.yaml` will override the default demo-usage `values.yaml`
- User needs to know the URL location of this `values.yaml` file and pass it to the command line when installing Kuma

## Decision Outcome

- A workable production-usage `values.yaml` overriding the default `values.yaml` that will help install a production-ready Kuma service mesh


## Proposed production installation variants

### Security

Restrict access:
- limit accessing to the Kuma control plane api server by setting CORS domains, so that only trusted website have access to the API server 
- enable global control plane authentication, so that only trusted zone CPs have access to the global CP

### Stability

Improves stability by:
- prevent pod disruptions by enabling the pod disruption budgets for control plane, ingress and egress components 
- ensuring a minimum number of 2 replicas for the control plane, ingress and egress components, and replicas are preferred to be scheduled onto different nodes
- ensuring resource limits identical to requests so that pod resources will not be squeezed by other applications on the cluster
- setting metrics for triggering HPA to 50% CPU and memory usage to make sure pods are not over-utilized when other instances are crashing

### Performance

Improves performance by:
- enabling HPAs for the control plane, ingress and egress components
- setting metrics for triggering HPA to 50% CPU and memory usage to make sure pods are performing well

### Proposed production installation values

The values here will override the default `values.yaml` file, so refer to [the existing file](https://github.com/kumahq/kuma/blob/master/deployments/charts/kuma/values.yaml) if needed. 

```yaml

controlPlane:
  # -- Used in `zone` mode with `kdsGlobalAddress` is not empty
  kdsGlobalAuth:
    type: "token"
    token: ""
    
  # this is a new key to restrict access to the Kuma control plane api server by cores
  apiServer:
    coresAllowedDomains:
      - "https://localhost:5681"
      - "http://localhost:5681"

  replicas: 2

  resources:
    requests:
      cpu: 1000m
      memory: 1024Mi
    limits:
      cpu: 1000m
      memory: 1024Mi

  autoscaling:
    enabled: true

    # -- For clusters that don't support autoscaling/v2, autoscaling/v1 is used
    targetCPUUtilizationPercentage: 50
    # -- For clusters that do support autoscaling/v2, use metrics
    metrics:
      - type: Resource
        resource:
          name: cpu
          target:
            type: Utilization
            averageUtilization: 50
      - type: Resource
        resource:
          name: memory
          target:
            type: Utilization
            averageUtilization: 50


  podDisruptionBudget:
    enabled: true

cni:
  resources:
    requests:
      cpu: 100m
      memory: 100Mi
    limits:
      cpu: 100m
      memory: 100Mi

ingress:
  replicas: 2

  resources:
    requests:
      cpu: 1000m
      memory: 1024Mi
    limits:
      cpu: 1000m
      memory: 1024Mi

  autoscaling:
    enabled: true
    # -- For clusters that don't support autoscaling/v2, autoscaling/v1 is used
    targetCPUUtilizationPercentage: 50
    # -- For clusters that do support autoscaling/v2, use metrics
    metrics:
      - type: Resource
        resource:
          name: cpu
          target:
            type: Utilization
            averageUtilization: 50
      - type: Resource
        resource:
          name: memory
          target:
            type: Utilization
            averageUtilization: 50

  podDisruptionBudget:
    enabled: true
```

### Negative Consequences

Will trouble users who:
- wants to integrate the control plane api server using a custom domain
- wants to install Kuma onto a cluster does not support metrics based HPA
