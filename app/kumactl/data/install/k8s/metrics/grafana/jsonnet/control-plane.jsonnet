local dashboard = import 'lib/dashboard.jsonnet';
local g = import 'lib/g.libsonnet';
local variables = import 'lib/variables.jsonnet';
local row = g.panel.row;
local stat = g.panel.stat;
local table = g.panel.table;
local timeSeries = g.panel.timeSeries;

{
  fileName: 'kuma-cp.json',
  definition:
    dashboard.new('Kuma CP')
    + dashboard.withVariables([
      variables.new('zone', 'Zone', 'zone', 'cp_info', true, false),
      variables.new('instance', 'Instance', 'instance', 'cp_info{zone=~"$zone"}', true, false),
    ])
    + dashboard.withPanels([
      row.new('Overview')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      stat.new('Condition')
      + variables.prometheusDS
      + stat.standardOptions.color.withMode('thresholds')
      + stat.standardOptions.thresholds.withMode('absolute')
      + stat.standardOptions.thresholds.withSteps([
        stat.thresholdStep.withColor('red')
        + stat.thresholdStep.withValue(null),

        stat.thresholdStep.withColor('green')
        + stat.thresholdStep.withValue(1),
      ])
      + stat.standardOptions.withMappings([
        stat.valueMapping.RangeMap.options.withFrom(1)
        + stat.valueMapping.RangeMap.options.withTo(9999)
        + stat.valueMapping.RangeMap.options.result.withColor('green')
        + stat.valueMapping.RangeMap.options.result.withIndex(0)
        + stat.valueMapping.RangeMap.options.result.withText('LIVE')
        + stat.valueMapping.RangeMap.withType('range'),

        stat.valueMapping.RangeMap.options.withFrom(0)
        + stat.valueMapping.RangeMap.options.withTo(0.999)
        + stat.valueMapping.RangeMap.options.result.withColor('red')
        + stat.valueMapping.RangeMap.options.result.withIndex(1)
        + stat.valueMapping.RangeMap.options.result.withText('UNAVAILABLE')
        + stat.valueMapping.RangeMap.withType('range'),
      ])
      + stat.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(cp_info{instance=~"$instance"} OR on() vector(0))')
        + g.query.prometheus.withIntervalFactor(1)
        + g.query.prometheus.withInstant(false)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withLegendFormat(''),
      ]),

      table.new('Control Plane information')
      + variables.prometheusDS
      + table.standardOptions.thresholds.withMode('absolute')
      + table.standardOptions.thresholds.withSteps([
        table.thresholdStep.withColor('green')
        + table.thresholdStep.withValue(null),

        table.thresholdStep.withColor('red')
        + table.thresholdStep.withValue(80),
      ])
      + table.queryOptions.withTargets([
        g.query.prometheus.withExpr('cp_info{instance=~"$instance"}')
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

      timeSeries.new('Leader')
      + timeSeries.withDescription('Only one instance should be a leader at a given point of time in every zone')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum by (zone, instance) (leader)')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{zone}} - {{instance}}'),
      ]),

      row.new('Aggregated Discovery Service (XDS): CP-DP communication')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('XDS active connections')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('xds_streams_active{instance=~"$instance"}')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      timeSeries.new('XDS message exchange')
      + timeSeries.withDescription('Number of ADS messages exchanged between Control Plane and Dataplane over XDS')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(xds_responses_sent{instance=~"$instance"}[1m]))')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Configuration sent'),

        g.query.prometheus.withExpr('sum(rate(xds_requests_received{confirmation="ACK",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withHide(false)
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Configuration accepted'),

        g.query.prometheus.withExpr('sum(rate(xds_requests_received{confirmation="NACK",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withLegendFormat('Configuration rejected'),

        g.query.prometheus.withExpr('sum(rate(grpc_server_handled_total{grpc_service="envoy.service.discovery.v3.AggregatedDiscoveryService",grpc_code!="OK",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('D')
        + g.query.prometheus.withLegendFormat('RPC Errors'),

        g.query.prometheus.withExpr('sum(rate(grpc_server_started_total{grpc_service="envoy.service.discovery.v3.AggregatedDiscoveryService",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('E')
        + g.query.prometheus.withLegendFormat('Rate'),
      ]),

      timeSeries.new('XDS config generations')
      + timeSeries.withDescription("Number of Envoy XDS config generations per second. Config is sent only when it's different than previously generated. The value should be the number of connected dataplanes * KUMA_XDS_SERVER_DATAPLANE_CONFIGURATION_REFRESH_INTERVAL")
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(xds_generation_count{instance=~"$instance"}[1m])) - sum(rate(xds_generation_errors{instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Success'),

        g.query.prometheus.withExpr('sum(rate(xds_generation_errors{instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Errors'),
      ]),

      timeSeries.new('Latency of XDS config generation (99th percentile)')
      + timeSeries.withDescription('How much it took to generate Envoy configuration for a single dataplane')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('xds_generation{quantile="0.99",instance=~"$instance"}')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      timeSeries.new('Latency of XDS config delivery (99th percentile)')
      + timeSeries.withDescription('How much it took to deliver Envoy configuration and receive ACK/NACK for a single dataplane')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('xds_delivery{quantile="0.99",instance=~"$instance"}')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      timeSeries.new('Endpoints cache performance')
      + timeSeries.withDescription('Cache protects ClusterLoadAssignments resources by sharing them between many goroutines which reconcile Dataplanes.\n\nhit - request was retrieved from the cache.\n\nhit-wait - request was retrieved from the cache after waiting for a concurrent request to fetch it from the database.\n\nmiss - request was fetched from the database\n\nRefer to https://kuma.io/docs/latest/documentation/fine-tuning/#snapshot-generation')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(cla_cache{result="hit",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Hit'),

        g.query.prometheus.withExpr('sum(rate(cla_cache{result="hit-wait",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withLegendFormat('Hit Wait'),

        g.query.prometheus.withExpr('sum(rate(cla_cache{result="miss",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Miss'),
      ]),

      timeSeries.new('Mesh resources hash cache performance')
      + timeSeries.withDescription('Mesh Cache protects hashes calculated periodically for each Mesh in order to avoid the excessive generation of xDS resources.\n\nhit - request was retrieved from the cache.\n\nhit-wait - request was retrieved from the cache after waiting for a concurrent request to fetch it from the database.\n\nmiss - request was fetched from the database\n\nRefer to https://kuma.io/docs/latest/documentation/fine-tuning/#snapshot-generation')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(mesh_cache{result="hit",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Hit'),

        g.query.prometheus.withExpr('sum(rate(mesh_cache{result="hit-wait",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withLegendFormat('Hit Wait'),

        g.query.prometheus.withExpr('sum(rate(mesh_cache{result="miss",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Miss'),
      ]),

      row.new('Health Discovery Service (HDS): CP-DP communication')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('HDS message exchange')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(hds_responses_sent{instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('endpoint health responses'),

        g.query.prometheus.withExpr('sum(rate(hds_requests_received{instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('health check requests'),
      ]),

      timeSeries.new('HDS active connections')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('hds_streams_active{instance=~"$instance"}')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      timeSeries.new('HDS config generations')
      + timeSeries.withDescription("Number of Envoy HDS config generations per second. Config is sent only when it's different than previously generated. The value should be the number of connected dataplanes * KUMA_DP_SERVER_HDS_REFRESH_INTERVAL")
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(hds_generation_count{instance=~"$instance"}[1m])) - sum(rate(hds_generation_errors{instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Success'),

        g.query.prometheus.withExpr('sum(rate(hds_generation_errors{instance=~"$instance"}[1m]))')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Errors'),
      ]),

      timeSeries.new('Latency of HDS config generation (99th percentile)')
      + timeSeries.withDescription('How much it took to generate Envoy configuration for a single dataplane')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('hds_generation{quantile="0.99",instance=~"$instance"}')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      row.new('API Server - Management of the Control Plane')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Latency of API Server Requests (99th percentile)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('histogram_quantile(0.99, sum by (handler, le) (rate(api_server_http_request_duration_seconds_bucket{instance=~"$instance"}[1m])))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{handler}}'),
      ]),

      timeSeries.new('Response Codes')
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum by (code) (rate(api_server_http_request_duration_seconds_count{instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{code}}'),
      ])
      + variables.prometheusDS,

      row.new('Dataplane Server: CP-DP communication')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Latency of Dataplane Server Requests (99th percentile)')
      + timeSeries.withDescription('XDS and SDS requests are not taken into account because they are long-running requests')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('histogram_quantile(0.99, sum by (handler, le) (rate(dp_server_http_request_duration_seconds_bucket{instance=~"$instance"}[1m])))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{handler}}'),
      ]),

      timeSeries.new('Response Codes')
      + timeSeries.withDescription('XDS and SDS requests are not taken into account because they are long-running requests')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum by (code) (rate(dp_server_http_response_size_bytes_count{instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{code}}'),
      ]),

      row.new('Kuma Discovery Service (KDS) - Mutltizone communication')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Policies count (sync check)')
      + timeSeries.withDescription('This metric presents if policies are properly propagated from Global to Zones. All instances of global and zones in the system should have the same number of policies. This metric have additional latency of 1 minute (it is not computed on the fly when scraping is done)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum by (zone, instance) (resources_count{instance=~"$instance",resource_type!~"Dataplane|DataplaneInsight|Zone|ZoneInsight|ZoneIngress|ZoneIngressInsight|ZoneEgress|ZoneEgressInsight"})')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{zone}} - {{instance}}'),
      ]),

      timeSeries.new('Dataplane count')
      + timeSeries.withDescription('This metric presents if Dataplanes are properly propagated from Zones to Global. Keep in mind that Dataplanes from all zones != Dataplanes from global. All zones receive additional one dataplane of Ingress from other Zones. This metric have additional latency of 1 minute (it is not computed on the fly when scraping is done)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum by (zone, instance) (resources_count{instance=~"$instance",resource_type="Dataplane"})')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{zone}} - {{instance}}'),
      ]),

      timeSeries.new('Global - KDS active connections')
      + timeSeries.withDescription('Number of zone CP connected to Global.')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('kds_streams_active{instance=~"$instance",zone="Global"}')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      timeSeries.new('Global - KDS message exchange')
      + timeSeries.withDescription('Number of KDS messages exchanged between Global Control Plane and Zone Control Plane over KDS. Global sends to Zone policies and Ingress Dataplanes')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(kds_responses_sent{instance=~"$instance",zone="Global"}[1m]))')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Configuration sent'),

        g.query.prometheus.withExpr('sum(rate(kds_requests_received{confirmation="ACK",instance=~"$instance",zone="Global"}[1m]))')
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Configuration accepted'),

        g.query.prometheus.withExpr('sum(rate(kds_requests_received{confirmation="NACK",instance=~"$instance",zone="Global"}[1m]))')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withLegendFormat('Configuration rejected'),

        g.query.prometheus.withExpr('sum(rate(grpc_server_handled_total{grpc_service="kuma.mesh.v1alpha1.MultiplexService",grpc_code!="OK",instance=~"$instance",zone="Global"}[1m]))')
        + g.query.prometheus.withRefId('D')
        + g.query.prometheus.withLegendFormat('RPC Errors'),

        g.query.prometheus.withExpr('sum(rate(grpc_server_started_total{grpc_service="kuma.mesh.v1alpha1.MultiplexService",instance=~"$instance",zone="Global"}[1m]))')
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withRefId('E')
        + g.query.prometheus.withLegendFormat('Rate'),
      ]),

      timeSeries.new('Global - KDS config generations')
      + timeSeries.withDescription("Number of KDS config generations per second. Config is sent only when it's different than previously generated. The value should be the number of connected control planes * KUMA_MULTIZONE_GLOBAL_KDS_REFRESH_INTERVAL")
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(kds_generation_count{instance=~"$instance",zone="Global"}[1m])) - sum(rate(kds_generation_errors{instance=~"$instance",zone="Global"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Success'),

        g.query.prometheus.withExpr('sum(rate(kds_generation_errors{instance=~"$instance",zone="Global"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Errors'),
      ]),

      timeSeries.new('Global - Latency of KDS config generation (99th percentile)')
      + timeSeries.withDescription('How much it took to generate KDS configuration for a single zone control plane')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('kds_generation{quantile="0.99",instance=~"$instance",zone="Global"}')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      timeSeries.new('Global - Latency of KDS config delivery (99th percentile)')
      + timeSeries.withDescription('How much it took to deliver Envoy configuration and receive ACK/NACK for a single dataplane')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('kds_delivery{quantile="0.99",instance=~"$instance",zone="Global"}')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      timeSeries.new('Zone - KDS active connections')
      + timeSeries.withDescription('Number of global CP connected to Zone.')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('kds_streams_active{instance=~"$instance",zone!="Global"}')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      timeSeries.new('Zone - KDS message exchange')
      + timeSeries.withDescription('Number of KDS messages exchanged between Global Control Plane and Zone Control Plane over KDS. Zone sends to Global its Dataplanes and DataplaneInsights')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(kds_responses_sent{instance=~"$instance",zone!="Global"}[1m]))')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Configuration sent'),

        g.query.prometheus.withExpr('sum(rate(kds_requests_received{confirmation="ACK",instance=~"$instance",zone!="Global"}[1m]))')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Configuration accepted'),

        g.query.prometheus.withExpr('sum(rate(kds_requests_received{confirmation="NACK",instance=~"$instance",zone!="Global"}[1m]))')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withLegendFormat('Configuration rejected'),

        g.query.prometheus.withExpr('sum(rate(grpc_server_handled_total{grpc_service="kuma.mesh.v1alpha1.MultiplexService",grpc_code!="OK",instance=~"$instance",zone!="Global"}[1m]))')
        + g.query.prometheus.withRefId('D')
        + g.query.prometheus.withLegendFormat('RPC Errors'),

        g.query.prometheus.withExpr('sum(rate(grpc_server_started_total{grpc_service="kuma.mesh.v1alpha1.MultiplexService",instance=~"$instance",zone!="Global"}[1m]))')
        + g.query.prometheus.withHide(true)
        + g.query.prometheus.withRefId('E')
        + g.query.prometheus.withLegendFormat('Rate'),
      ]),

      timeSeries.new('Zone - KDS config generations')
      + timeSeries.withDescription("Number of KDS config generations per second. Config is sent only when it's different than previously generated. The value should be the number of connected control planes * KUMA_MULTIZONE_ZONE_KDS_REFRESH_INTERVAL")
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(kds_generation_count{instance=~"$instance",zone!="Global"}[1m])) - sum(rate(kds_generation_errors{instance=~"$instance",zone!="Global"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Success'),

        g.query.prometheus.withExpr('sum(rate(kds_generation_errors{instance=~"$instance",zone!="Global"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Errors'),
      ]),

      timeSeries.new('Zone - Latency of KDS config generation (99th percentile)')
      + timeSeries.withDescription('How much it took to generate KDS configuration for a global control plane')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('kds_generation{quantile="0.99",instance=~"$instance",zone!="Global"}')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      timeSeries.new('Global - Latency of KDS config delivery (99th percentile)')
      + timeSeries.withDescription('How much it took to deliver Envoy configuration and receive ACK/NACK for a single dataplane')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('kds_delivery{quantile="0.99",instance=~"$instance",zone="Global"}')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      row.new('Store')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Latency of Store operations (99th percentile)')
      + timeSeries.withDescription('Latency of underlying storage (API Server on K8S, Postgres on Universal etc.)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('histogram_quantile(0.99, sum by (operation, le) (rate(store_bucket{instance=~"$instance"}[1m])))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{operation}}'),
      ]),

      timeSeries.new('Store cache performance')
      + timeSeries.withDescription('Cache protects underlying storage by sharing responses between many goroutines accessing the storage.\n\nhit - request was retrieved from the cache.\n\nhit-wait - request was retrieved from the cache after waiting for a concurrent request to fetch it from the database.\n\nmiss - request was fetched from the database\n\nRefer to https://kuma.io/docs/latest/documentation/fine-tuning/#snapshot-generation')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum(rate(store_cache{result="hit",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('Hit'),

        g.query.prometheus.withExpr('sum(rate(store_cache{result="hit-wait",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('C')
        + g.query.prometheus.withLegendFormat('Hit Wait'),

        g.query.prometheus.withExpr('sum(rate(store_cache{result="miss",instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('B')
        + g.query.prometheus.withLegendFormat('Miss'),
      ]),

      timeSeries.new('Store operations')
      + timeSeries.withDescription('Real requests executed on the underlying storage (Postgres/Kubernetes API Server)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('sum by (operation, resource_type) (rate(store_count{instance=~"$instance"}[1m]))')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{operation}} - {{resource_type}}'),
      ]),

      row.new('Go Runtime')
      + variables.prometheusDS
      + row.withCollapsed(false)
      + row.withPanels([]),

      timeSeries.new('Goroutines')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('go_goroutines{instance=~"$instance"}')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      timeSeries.new('Threads')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('go_threads{instance=~"$instance"}')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      timeSeries.new('Memory allocated')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('go_memstats_alloc_bytes{instance=~"$instance"}')
        + g.query.prometheus.withExemplar(true)
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),

      timeSeries.new('Latency of GC time (75th percentile)')
      + variables.prometheusDS
      + timeSeries.queryOptions.withTargets([
        g.query.prometheus.withExpr('go_gc_duration_seconds{instance=~"$instance", quantile="0.75"}')
        + g.query.prometheus.withRefId('A')
        + g.query.prometheus.withLegendFormat('{{instance}}'),
      ]),
    ]),
}
