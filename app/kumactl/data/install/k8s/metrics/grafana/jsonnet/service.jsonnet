local dashboard = import 'lib/dashboard.jsonnet';
local g = import 'lib/g.libsonnet';
local variables = import 'lib/variables.jsonnet';
local row = g.panel.row;
local stat = g.panel.stat;
local table = g.panel.table;
local timeSeries = g.panel.timeSeries;
local nodeGraph = g.panel.nodeGraph;
local gauge = g.panel.gauge;

{
  fileName: 'kuma-service.json',
  definition:
    dashboard.new('Kuma Service')
    + dashboard.withVariables([
      variables.mesh,
      variables.zone,
      variables.new('service', 'Service', 'kuma_io_service', 'envoy_server_live{mesh="$mesh",kuma_io_zone=~"$zone",kuma_io_mesh_gateway=""}', false, false),
    ])
    + dashboard.withPanels([
      stat.new('Dataplanes')
      + stat.standardOptions.color.withMode('fixed')
      + stat.standardOptions.thresholds.withMode('absolute')
      + variables.prometheusDS
      + stat.queryOptions.withTargets([
        g.query.prometheus.withExpr('count(envoy_server_live{kuma_io_services=~".*$service.*"})')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat(''),
      ]),

      table.new('Dataplanes')
      + table.standardOptions.color.withMode('thresholds')
      + variables.prometheusDS
      + table.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(envoy_server_live{kuma_io_services=~".*$service.*"}) by (dataplane)')
        + g.query.prometheus.withInstant(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat(''),
      ])
      + table.queryOptions.withTransformations([
        table.transformations.withId('labelsToFields')
        + table.transformations.withOptions({}),

        table.transformations.withId('merge')
        + table.transformations.withOptions({}),

        table.transformations.withId('organize')
        + table.transformations.withOptions({}),

        table.transformations.withId('merge')
        + table.transformations.withOptions({}),
      ]),

      stat.new('')
      + stat.standardOptions.color.withMode('thresholds')
      + variables.prometheusDS
      + stat.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(envoy_cluster_upstream_rq_total{mesh="$mesh",kuma_io_zone=~"$zone",kuma_io_services=~".*$service.*",envoy_cluster_name=~"localhost_.*"}[1m]))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat(''),
      ]),

      stat.new('')
      + stat.standardOptions.color.withMode('thresholds')
      + variables.prometheusDS
      + stat.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(envoy_cluster_upstream_rq_total{mesh="$mesh",kuma_io_zone=~"$zone",kuma_io_services=~".*$service.*",envoy_cluster_name!~"localhost_.*",envoy_cluster_name!="kuma_envoy_admin",envoy_cluster_name!="kuma_metrics_hijacker"}[1m]))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat(''),
      ]),

      row.new('HTTP')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Latency')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('max(histogram_quantile(0.99, rate(envoy_cluster_upstream_rq_time_bucket{kuma_io_services=~".*$service.*",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost_.*"}[1m])))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('p99'),

        g.query.prometheus.withExpr('max(histogram_quantile(0.95, rate(envoy_cluster_upstream_rq_time_bucket{kuma_io_services=~".*$service.*",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost_.*"}[1m])))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('p95'),

        g.query.prometheus.withExpr('max(histogram_quantile(0.50, rate(envoy_cluster_upstream_rq_time_bucket{kuma_io_services=~".*$service.*",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost_.*"}[1m])))')
        + g.query.prometheus.withRefId('D')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('p50'),
      ]),

      timeSeries.new('Traffic')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(envoy_cluster_upstream_rq_total{mesh="$mesh",kuma_io_zone=~"$zone",kuma_io_services=~".*$service.*", envoy_cluster_name=~"localhost_.*"}[1m]))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('Incoming'),

        g.query.prometheus.withExpr('sum(rate(envoy_cluster_upstream_rq_total{mesh="$mesh",kuma_io_zone=~"$zone",kuma_io_services=~".*$service.*", envoy_cluster_name!~"localhost_.*", envoy_cluster_name!="kuma_envoy_admin",envoy_cluster_name!="kuma_metrics_hijacker"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('Outgoing'),
      ]),

      timeSeries.new('Status codes')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(envoy_cluster_external_upstream_rq_xx{mesh="$mesh",kuma_io_zone=~"$zone",kuma_io_services=~".*$service.*", envoy_cluster_name=~"localhost_.*"}[1m])) by (envoy_response_code_class)')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('Incoming {{envoy_response_code_class}}xx'),

        g.query.prometheus.withExpr('sum(rate(envoy_cluster_external_upstream_rq_xx{mesh="$mesh",kuma_io_zone=~"$zone",kuma_io_services=~".*$service.*", envoy_cluster_name!~"localhost_.*", envoy_cluster_name!="kuma_envoy_admin", envoy_cluster_name!="kuma_metrics_hijacker"}[1m])) by (envoy_response_code_class)')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('Outgoing {{envoy_response_code_class}}xx'),
      ]),

      row.new('Kubernetes')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('CPU')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('max(sum(rate(container_cpu_usage_seconds_total[1m])) by (namespace, pod) * on (namespace, pod) group_right(kuma_io_service) envoy_server_live{kuma_io_services=~".*$service.*"}) by (dataplane) /\nmax(sum(kube_pod_container_resource_limits{resource="cpu",unit="core"}) by (namespace, pod) * on (namespace, pod) group_right(kuma_io_service) envoy_server_live{kuma_io_services=~".*$service.*"}) by (dataplane)')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('{{dataplane}}'),
      ]),

      timeSeries.new('Memory Utilization')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('max(sum(container_memory_working_set_bytes{image!=""}) by (namespace, pod) * on (namespace, pod) group_right(kuma_io_service) envoy_server_live{kuma_io_services=~".*$service.*"}) by (dataplane)')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('{{dataplane}}'),
      ]),

      timeSeries.new('Memory Saturation')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('max(sum(container_memory_working_set_bytes) by (namespace, pod) * on (namespace, pod) group_right(kuma_io_service) envoy_server_live{kuma_io_services=~".*$service.*"}) by (dataplane) / max(sum(kube_pod_container_resource_limits{resource="memory",unit="byte"}) by (namespace, pod) * on (namespace, pod) group_right(kuma_io_service) envoy_server_live{kuma_io_services=~".*$service.*"}) by (dataplane)')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('{{dataplane}}'),
      ]),
    ]),
}
