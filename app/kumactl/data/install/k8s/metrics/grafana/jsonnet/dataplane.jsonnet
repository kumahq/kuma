local dashboard = import 'lib/dashboard.jsonnet';
local g = import 'lib/g.libsonnet';
local variables = import 'lib/variables.jsonnet';
local row = g.panel.row;
local stat = g.panel.stat;
local table = g.panel.table;
local timeSeries = g.panel.timeSeries;
local gauge = g.panel.gauge;

{
  fileName: 'kuma-dataplane.json',
  definition:
    dashboard.new('Kuma Dataplane')
    + dashboard.withDescription('Statistics of a single Dataplane in Kuma Service Mesh')
    + dashboard.withVariables([
      variables.mesh,
      variables.zone,
      variables.new('dataplane', 'Dataplane', 'dataplane', 'envoy_server_live{mesh="$mesh",kuma_io_zone=~"$zone"}', false, false),
    ])
    + dashboard.withPanels([
      row.new('Dataplane')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      gauge.new('Status')
      + gauge.standardOptions.color.withMode('thresholds')
      + variables.prometheusDS
      + gauge.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(envoy_server_live{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"} OR on() vector(0))')
        + g.query.prometheus.withIntervalFactor(1)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat(''),
      ]),

      stat.new('Uptime')
      + stat.standardOptions.color.withMode('thresholds')
      + variables.prometheusDS
      + stat.queryOptions.withTargets([
        g.query.prometheus.withExpr('envoy_server_uptime{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}')
        + g.query.prometheus.withRefId('A'),
      ]),

      stat.new('Heap size')
      + stat.standardOptions.color.withMode('thresholds')
      + variables.prometheusDS
      + stat.queryOptions.withTargets([
        g.query.prometheus.withExpr('envoy_server_memory_heap_size{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}')
        + g.query.prometheus.withRefId('A'),
      ]),

      stat.new('Memory allocated')
      + stat.standardOptions.color.withMode('thresholds')
      + variables.prometheusDS
      + stat.queryOptions.withTargets([
        g.query.prometheus.withExpr('envoy_server_memory_allocated{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}')
        + g.query.prometheus.withRefId('A'),
      ]),

      timeSeries.new('Connection to the Control Plane')
      + timeSeries.withDescription('Note that if Control Plane does not sent FIN segment, Dataplanes can still think that connection is up waiting for new update even that Control Plane is down.')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('envoy_control_plane_connected_state{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Connected'),
      ]),

      timeSeries.new('Control Plane connected to')
      + timeSeries.standardOptions.color.withMode('palette-classic')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('envoy_control_plane_identifier{dataplane="$dataplane", kuma_io_zone=~"$zone", mesh="$mesh"}')
        + g.query.prometheus.withEditorMode('code')
        + g.query.prometheus.withExemplar(false)
        + g.query.prometheus.withFormat('time_series')
        + g.query.prometheus.withInstant(false)
        + g.query.prometheus.withRange(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{text_value}}'),
      ]),

      row.new('Incoming traffic')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Active Connections to this Dataplane')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('envoy_cluster_upstream_cx_active{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost.*"}')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      timeSeries.new('Connection time (P99)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('histogram_quantile(0.99, irate(envoy_cluster_upstream_cx_connect_ms_bucket{dataplane="$dataplane",envoy_cluster_name=~"localhost.*",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      timeSeries.new('Connection length (P99)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('histogram_quantile(0.99, irate(envoy_cluster_upstream_cx_length_ms_bucket{dataplane="$dataplane",envoy_cluster_name=~"localhost.*",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      timeSeries.new('Bytes received from requests')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_cx_tx_bytes_total{dataplane="$dataplane",envoy_cluster_name=~"localhost.*",kuma_io_zone=~"$zone",mesh="$mesh"}[1m])')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      timeSeries.new('Bytes sent in responses')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_cx_rx_bytes_total{dataplane="$dataplane",envoy_cluster_name=~"localhost.*",kuma_io_zone=~"$zone",mesh="$mesh"}[1m])')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      timeSeries.new('Traffic permissions')
      + timeSeries.withDescription('Data is only available if TrafficPermission policy is applied.')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('irate(envoy_rbac_allowed{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}[1m])')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Allowed - {{listener}}'),

        g.query.prometheus.withExpr('irate(envoy_rbac_shadow_allowed{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}[1m])')
        + g.query.prometheus.withRefId('D')
        + g.query.prometheus.withLegendFormat('Shadow allowed - {{listener}}'),

        g.query.prometheus.withExpr('irate(envoy_rbac_denied{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}[1m])')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Denied - {{listener}}'),

        g.query.prometheus.withExpr('irate(envoy_rbac_shadow_denied{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}[1m])')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withLegendFormat('Shadow denied - {{listener}}'),
      ]),

      timeSeries.new('Connection/Requests errors')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_cx_destroy_remote_with_active_rq{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost.*"}[1m])')
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('connection destroyed by the client - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_cx_connect_timeout{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost.*"}[1m])')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('connection timeout - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_cx_destroy_local_with_active_rq{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost.*"}[1m])')
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withLegendFormat('connection destroyed by local Envoy - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_rq_pending_failure_eject{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost.*"}[1m])')
        + g.query.prometheus.withRefId('D')
        + g.query.prometheus.withLegendFormat('pending failure ejection - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_rq_pending_overflow{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost.*"}[1m])')
        + g.query.prometheus.withRefId('E')
        + g.query.prometheus.withLegendFormat('pending overflow - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_rq_timeout{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost.*"}[1m])')
        + g.query.prometheus.withRefId('F')
        + g.query.prometheus.withLegendFormat('request timeout - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_rq_rx_reset{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost.*"}[1m])')
        + g.query.prometheus.withRefId('G')
        + g.query.prometheus.withLegendFormat('response reset by the client - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_rq_tx_reset{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost.*"}[1m])')
        + g.query.prometheus.withRefId('H')
        + g.query.prometheus.withLegendFormat('request reset by local Envoy - {{envoy_cluster_name}}'),
      ]),

      row.new('Outgoing traffic')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Active Connections to other Dataplanes')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('envoy_cluster_upstream_cx_active{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink"}')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      timeSeries.new('Connection time (P99)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('histogram_quantile(0.99, irate(envoy_cluster_upstream_cx_connect_ms_bucket{dataplane="$dataplane",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      timeSeries.new('Connection length (P99)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('histogram_quantile(0.99, irate(envoy_cluster_upstream_cx_length_ms_bucket{dataplane="$dataplane",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      timeSeries.new('Bytes received from other Dataplanes')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_cx_rx_bytes_total{dataplane="$dataplane",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink",kuma_io_zone=~"$zone",mesh="$mesh"}[1m])')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      timeSeries.new('Bytes sent to other Dataplanes')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_cx_tx_bytes_total{dataplane="$dataplane",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink",kuma_io_zone=~"$zone",mesh="$mesh"}[1m])')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      table.new('Secured destinations by mTLS')
      + table.withDescription('Data is only available if TrafficPermission policy is applied.')
      + variables.prometheusDS
      + table.queryOptions.withTargets([
        g.query.prometheus.withExpr('envoy_cluster_ssl_handshake{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"} > 0')
        + g.query.prometheus.withFormat('time_series')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      timeSeries.new('Connection/Requests errors')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_cx_destroy_remote_with_active_rq{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink"}[1m])')
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('destroyed by remote Envoy - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_cx_connect_timeout{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink"}[1m])')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('connection timeout - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_cx_destroy_local_with_active_rq{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink"}[1m])')
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withLegendFormat('destroyed by local Envoy - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_rq_pending_failure_eject{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink"}[1m])')
        + g.query.prometheus.withRefId('D')
        + g.query.prometheus.withLegendFormat('pending failure ejection - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_rq_pending_overflow{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink"}[1m])')
        + g.query.prometheus.withRefId('E')
        + g.query.prometheus.withLegendFormat('pending overflow - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_rq_timeout{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink"}[1m])')
        + g.query.prometheus.withRefId('F')
        + g.query.prometheus.withLegendFormat('request timeout - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_rq_rx_reset{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink"}[1m])')
        + g.query.prometheus.withRefId('G')
        + g.query.prometheus.withLegendFormat('request reset by other Envoy - {{envoy_cluster_name}}'),

        g.query.prometheus.withExpr('irate(envoy_cluster_upstream_rq_tx_reset{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name!~"(localhost.*)|ads_cluster|kuma_envoy_admin|access_log_sink"}[1m])')
        + g.query.prometheus.withRefId('H')
        + g.query.prometheus.withLegendFormat('response reset by local Envoy - {{envoy_cluster_name}}'),
      ]),

      row.new('HTTP')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Latency')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(histogram_quantile(0.99, rate(envoy_cluster_upstream_rq_time_bucket{kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost_.*",dataplane="$dataplane"}[1m])))')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('p99'),

        g.query.prometheus.withExpr('sum(histogram_quantile(0.95, rate(envoy_cluster_upstream_rq_time_bucket{kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost_.*",dataplane="$dataplane"}[1m])))')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('p95'),

        g.query.prometheus.withExpr('sum(histogram_quantile(0.50, rate(envoy_cluster_upstream_rq_time_bucket{kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost_.*",dataplane="$dataplane"}[1m])))')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withLegendFormat('p50'),
      ]),

      timeSeries.new('Traffic')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(envoy_cluster_upstream_rq_total{kuma_io_zone=~"$zone",mesh="$mesh",envoy_cluster_name=~"localhost_.*",dataplane="$dataplane"}[1m]))')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Incoming'),

        g.query.prometheus.withExpr('sum(rate(envoy_cluster_upstream_rq_total{kuma_io_zone=~"$zone",mesh="$mesh",dataplane="$dataplane",envoy_cluster_name!~"localhost_.*", envoy_cluster_name!="kuma_envoy_admin", envoy_cluster_name!="kuma_metrics_hijacker"}[1m]))')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Outgoing'),
      ]),

      timeSeries.new('Status codes')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(envoy_cluster_external_upstream_rq_xx{kuma_io_zone=~"$zone",mesh="$mesh",dataplane="$dataplane", envoy_cluster_name=~"localhost_.*"}[1m])) by (envoy_response_code_class)')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Incoming {{envoy_response_code_class}}xx'),

        g.query.prometheus.withExpr('sum(rate(envoy_cluster_external_upstream_rq_xx{kuma_io_zone=~"$zone",mesh="$mesh",dataplane="$dataplane", envoy_cluster_name!~"localhost_.*", envoy_cluster_name!="kuma_envoy_admin", envoy_cluster_name!="kuma_metrics_hijacker"}[1m])) by (envoy_response_code_class)')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Outgoing {{envoy_response_code_class}}xx'),
      ]),

      row.new('Health Checks')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Active Health Checks - healthy service instances')
      + timeSeries.withDescription('Data is only available if HealthCheck policy is applied.')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('envoy_cluster_health_check_healthy{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      timeSeries.new('Active Health Checks - failing service instances')
      + timeSeries.withDescription('Data is only available if HealthCheck policy is applied.')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('envoy_cluster_membership_total{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"} - envoy_cluster_health_check_healthy{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('{{envoy_cluster_name}}'),
      ]),

      row.new('Circuit Breakers')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Thresholds Overflow')
      + timeSeries.withDescription('Total times that the cluster’s connection circuit breaker overflowed')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_cx_overflow{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Connection overflow'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_pending_overflow{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Pending request overflow'),

        g.query.prometheus.withExpr('sum(irate(envoy_cluster_upstream_rq_retry_overflow{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}[1m]))')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withLegendFormat('Retry overflow'),
      ]),

      timeSeries.new('Outlier detection')
      + timeSeries.withDescription('Data is only available if HealthCheck policy is applied. Note that passive health checks are executed on healthy instances marked by active health checks.')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('1 - sum(envoy_cluster_outlier_detection_ejections_active{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"}) / sum(envoy_cluster_membership_total{dataplane="$dataplane",kuma_io_zone=~"$zone",mesh="$mesh"})')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Healthy destinations'),
      ]),
    ]),
}
