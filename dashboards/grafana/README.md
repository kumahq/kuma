# Grafana Dashboards

Grafana dashboard definitions (JSON) for Kuma.

Dashboards placed here are automatically included in release tarballs
under `kuma-VERSION/dashboards/grafana/`.

See [MADR-096](../../docs/madr/decisions/096-observability-dashboards.md)
for the decision record.

## Prometheus scrape jobs

The dashboards filter metrics by `job` label. Users must scrape Kuma
components into matching jobs:

| Job                 | Targets                                       | Used by                                  |
|---------------------|-----------------------------------------------|------------------------------------------|
| `kuma-control-plane`| `kuma-cp` `/metrics` (port 5680)              | Control Plane dashboard                  |
| `kuma-dataplanes`   | Workload sidecar `/stats/prometheus` (9902)   | Service / Mesh / Workload dashboards     |
| `kuma-zone-proxies` | Zone Ingress + Zone Egress `/stats/prometheus`| Zone Ingress / Zone Egress dashboards    |

Example scrape config for `kuma-zone-proxies` (Kubernetes pod SD):

```yaml
- job_name: kuma-zone-proxies
  metrics_path: /stats/prometheus
  kubernetes_sd_configs:
    - role: pod
      namespaces:
        names: [kuma-system]
  relabel_configs:
    - source_labels: [__meta_kubernetes_pod_label_app]
      regex: 'kuma-.*-(ingress|egress)'
      action: keep
    - source_labels: [__meta_kubernetes_pod_label_app]
      regex: 'kuma-.*-(ingress|egress)'
      target_label: proxy
      replacement: '$1'
    - source_labels: [__address__]
      regex: '(.+?)(:[0-9]+)?'
      target_label: __address__
      replacement: '${1}:9902'
    - source_labels: [__meta_kubernetes_pod_name]
      target_label: pod
    - target_label: namespace
      replacement: kuma-system
    - target_label: zone
      replacement: <your-zone-name>
```

The zone proxy dashboards expect the `proxy`, `pod`, `namespace`, and
`zone` labels above to be present on every scraped sample.
