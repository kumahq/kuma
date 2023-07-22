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
  fileName: 'kuma-service-to-service.json',
  definition:
    dashboard.new('Kuma Service to Service')
    + dashboard.withDescription('Statistics of the traffic between services in Kuma Service Mesh')
    + dashboard.withVariables([
      variables.mesh,
      variables.zone,
      variables.new('source_service', 'Source service', 'kuma_io_service', 'envoy_server_live{mesh="$mesh",kuma_io_mesh_gateway=""}', false, false),
      variables.new('destination_cluster', 'Destination service', 'envoy_cluster_name', 'envoy_cluster_upstream_cx_active{kuma_io_service="$source_service",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink"}', false, false),
    ])
    + dashboard.withPanels([
      row.new('Traffic')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Traffic from source service perspective')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_tx_bytes_total{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Bytes sent'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_rx_bytes_total{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Bytes received'),
      ]),

      timeSeries.new('Connection/Requests errors from source service perspective')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_destroy_remote_with_active_rq{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withLegendFormat('Connection destroyed by the client'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_connect_timeout{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Connection timeout'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_destroy_local_with_active_rq{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withLegendFormat('Connection destroyed by local Envoy'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_pending_failure_eject{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('D')
        + g.query.prometheus.withLegendFormat('Pending failure ejection'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_pending_overflow{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('E')
        + g.query.prometheus.withLegendFormat('Pending overflow'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_timeout{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('F')
        + g.query.prometheus.withLegendFormat('Request timeout'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_rx_reset{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('G')
        + g.query.prometheus.withLegendFormat('Response reset'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_tx_reset{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('H')
        + g.query.prometheus.withLegendFormat('Request reset'),
      ]),

      timeSeries.new('Active Connections between services')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(envoy_cluster_upstream_cx_active{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"})')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Connections'),
      ]),

      timeSeries.new('Connection time (P99)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('max(histogram_quantile(0.99, irate(envoy_cluster_upstream_cx_connect_ms_bucket{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m])))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Time'),
      ]),

      timeSeries.new('Connection length (P99)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('max(histogram_quantile(0.99, irate(envoy_cluster_upstream_cx_length_ms_bucket{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m])))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Time'),
      ]),

      row.new('HTTP')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Latency')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('max(histogram_quantile(0.99, rate(envoy_cluster_upstream_rq_time_bucket{kuma_io_service="$source_service",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name="$destination_cluster"}[1m])))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('p99'),

        g.query.prometheus.withExpr('max(histogram_quantile(0.95, rate(envoy_cluster_upstream_rq_time_bucket{kuma_io_service="$source_service",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name="$destination_cluster"}[1m])))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('p95'),

        g.query.prometheus.withExpr('max(histogram_quantile(0.50, rate(envoy_cluster_upstream_rq_time_bucket{kuma_io_service="$source_service",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name="$destination_cluster"}[1m])))')
        + g.query.prometheus.withRefId('D')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('p50'),
      ]),

      timeSeries.new('Traffic')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(envoy_cluster_upstream_rq_total{mesh="$mesh",kuma_io_service="$source_service", envoy_cluster_name="$destination_cluster"}[1m]))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('Requests'),
      ]),

      timeSeries.new('Status codes')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(envoy_cluster_external_upstream_rq_xx{mesh="$mesh",kuma_io_service="$source_service", envoy_cluster_name="$destination_cluster"}[1m])) by (envoy_response_code_class)')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('{{envoy_response_code_class}}xx'),
      ]),

      row.new('Health Checks')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Active Health Checks')
      + timeSeries.withDescription('Data is only available if HealthCheck policy is applied.')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(envoy_cluster_health_check_healthy{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}) / sum(envoy_cluster_membership_total{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"})')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Healthy destinations'),
      ]),

      row.new('Circuit Breakers')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Thresholds Overflow')
      + timeSeries.withDescription('Total times that the cluster’s connection circuit breaker overflowed')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_overflow{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Connection overflow'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_pending_overflow{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('Pending request overflow'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_retry_overflow{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('Retry overflow'),
      ]),

      timeSeries.new('Outlier detection')
      + timeSeries.withDescription('Data is only available if CircuitBreaker policy is applied')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('1 - sum(envoy_cluster_outlier_detection_ejections_active{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"}) / sum(envoy_cluster_membership_total{kuma_io_service="$source_service",envoy_cluster_name="$destination_cluster",kuma_io_zone=~"$zone",mesh="$mesh"})')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Healthy destinations'),
      ]),
    ]),
}
