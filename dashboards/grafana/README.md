# Grafana Dashboards

Grafana dashboard definitions (JSON) for Kuma.

Dashboards placed here are automatically included in release tarballs
under `kuma-VERSION/dashboards/grafana/`.

See [MADR-096](../../docs/madr/decisions/096-observability-dashboards.md)
for the decision record.

## Metric source

Proxy metrics come from the `MeshMetric` Prometheus backend, which kuma-dp exposes on port `5670`. Every sample it emits already carries `mesh`, `zone`, `kuma_workload` and `kuma_proxy_role` (`sidecar`, `gateway`, `zone-ingress` or `zone-egress`), so the dashboards select proxies by those labels and need no relabeling beyond `pod` and `namespace`.

Only the Control Plane dashboard filters by `job`, because `kuma-cp` metrics carry no proxy labels.

| Job                  | Targets                          | Used by                                     |
|----------------------|----------------------------------|---------------------------------------------|
| `kuma-control-plane` | `kuma-cp` `/metrics` (port 5680) | Control Plane dashboard                     |
| any                  | kuma-dp `/metrics` (port 5670)   | Service / Mesh / Zone Ingress / Zone Egress |

Example scrape config for proxies (Kubernetes pod SD). Zone Ingress and Zone Egress are ordinary `Dataplane`s with an injected sidecar, so this one job covers them too:

```yaml
- job_name: kuma-dataplanes
  metrics_path: /metrics
  kubernetes_sd_configs:
    - role: pod
  relabel_configs:
    - source_labels: [__meta_kubernetes_pod_container_name]
      regex: kuma-sidecar
      action: keep
    - source_labels: [__meta_kubernetes_pod_phase]
      regex: Running
      action: keep
    - source_labels: [__address__]
      regex: '(.+?)(:[0-9]+)?'
      target_label: __address__
      replacement: '${1}:5670'
    - source_labels: [__meta_kubernetes_pod_name]
      target_label: pod
    - source_labels: [__meta_kubernetes_namespace]
      target_label: namespace
```
