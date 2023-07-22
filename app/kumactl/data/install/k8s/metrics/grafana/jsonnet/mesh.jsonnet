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
  fileName: 'kuma-mesh.json',
  definition:
    dashboard.new('Kuma Mesh')
    + dashboard.withDescription('Statistics of the single Mesh in Kuma Service Mesh')
    + dashboard.withVariables([
      variables.mesh,
      variables.zone,
    ])
    + dashboard.withPanels([
      row.new('Service Map')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      nodeGraph.new('Service Map')
      + nodeGraph.datasource.withType('Kuma')
      + nodeGraph.queryOptions.withInterval('1m')
      + nodeGraph.queryOptions.withTargets([
        {
          hide: false,
          mesh: '$mesh',
          queryType: 'mesh-graph',
          refId: 'A',
        },
      ]),

      row.new('HTTP')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Latency (99th percentile)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(histogram_quantile(0.99, rate(envoy_cluster_upstream_rq_time_bucket{mesh="$mesh",kuma_io_zone=~"$zone",envoy_cluster_name=~"localhost_.*"}[1m]))) by (kuma_io_service)')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('{{kuma_io_service}}'),
      ]),

      timeSeries.new('Traffic')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(envoy_cluster_upstream_rq_total{mesh="$mesh",kuma_io_zone=~"$zone",envoy_cluster_name=~"localhost_.*"}[1m])) by (kuma_io_service)')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('{{kuma_io_service}}'),
      ]),

      timeSeries.new('Error Status Codes')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(envoy_cluster_external_upstream_rq_xx{mesh="$mesh",kuma_io_zone=~"$zone",envoy_cluster_name=~"localhost_.*", envoy_response_code_class=~"4|5"}[1m])) by (kuma_io_service,envoy_response_code_class)')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('{{ kuma_io_service}} {{ envoy_response_code_class }}xx'),
      ]),

      row.new('Health Checks')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Active Health Checks')
      + timeSeries.withDescription('Data is only available if HealthCheck policy is applied.')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('((sum(rate(envoy_cluster_health_check_success{mesh="$mesh",kuma_io_zone=~"$zone"}[1m])) / sum(rate(envoy_cluster_health_check_attempt{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Success rate'),
      ]),

      row.new('Circuit Breakers')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Thresholds Overflow')
      + timeSeries.withDescription('Total times that the cluster’s connection circuit breaker overflowed')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_overflow{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Connection overflow'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_pending_overflow{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('Pending request overflow'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_retry_overflow{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('Connection overflow'),
      ]),

      timeSeries.new('Outlier detection')
      + timeSeries.withDescription('Data is only available if HealthCheck policy is applied. Note that passive health checks are executed on healthy instances marked by active health checks.')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('1 - sum(envoy_cluster_outlier_detection_ejections_active{mesh="$mesh",kuma_io_zone=~"$zone"}) / sum(envoy_cluster_membership_total{mesh="$mesh",kuma_io_zone=~"$zone"})')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Healthy destinations'),
      ]),

      row.new('Data Plane Proxies')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Dataplanes')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('count(envoy_server_live{mesh="$mesh",kuma_io_zone=~"$zone"}) - sum(envoy_server_live{mesh="$mesh",kuma_io_zone=~"$zone"})')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('Off'),

        g.query.prometheus.withExpr('sum(envoy_server_live{mesh="$mesh",kuma_io_zone=~"$zone"})')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withLegendFormat('Live'),
      ]),

      timeSeries.new('Dataplanes connected to the Control Plane')
      + timeSeries.withDescription('Note that if Control Plane does not sent FIN segment, Dataplanes can still think that connection is up waiting for new update even that Control Plane is down.')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('count(envoy_server_live{mesh="$mesh",kuma_io_zone=~"$zone"}) - sum(envoy_control_plane_connected_state{mesh="$mesh",kuma_io_zone=~"$zone"})')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Disconnected'),

        g.query.prometheus.withExpr('sum(envoy_control_plane_connected_state{mesh="$mesh",kuma_io_zone=~"$zone"})')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Connected'),
      ]),

      timeSeries.new('Bytes flowing through Envoy')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_tx_bytes_total{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Sent'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_rx_bytes_total{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Received'),
      ]),

      timeSeries.new('Connection/Requests errors')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_destroy_remote_with_active_rq{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withLegendFormat('Connection destroyed by the client'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_connect_timeout{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Connection timeout'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_destroy_local_with_active_rq{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withLegendFormat('Connection destroyed by local Envoy'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_pending_failure_eject{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('D')
        + g.query.prometheus.withLegendFormat('Pending failure ejection'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_pending_overflow{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('E')
        + g.query.prometheus.withLegendFormat('Pending overflow'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_timeout{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('F')
        + g.query.prometheus.withLegendFormat('Request timeout'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_rx_reset{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('G')
        + g.query.prometheus.withLegendFormat('Response reset'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_tx_reset{mesh="$mesh",kuma_io_zone=~"$zone"}[1m]))')
        + g.query.prometheus.withRefId('H')
        + g.query.prometheus.withLegendFormat('Request reset'),
      ]),
    ]),
}
