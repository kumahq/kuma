local g = import 'g.libsonnet';
local variable = g.dashboard.variable;


{
  new(name, label, query_label, query_metric, include_all, multi):
    variable.query.new(name)
    + self.prometheusDS
    + variable.query.generalOptions.withLabel(label)
    + variable.query.queryTypes.withLabelValues(query_label, query_metric)
    + variable.query.selectionOptions.withIncludeAll(include_all)
    + variable.query.selectionOptions.withMulti(multi),

  prometheusDS: variable.query.withDatasource(
    type='prometheus',
    uid='${DS_PROMETHEUS}',
  ),

  mesh: self.new('mesh', 'Mesh', 'mesh', 'envoy_server_live', false, false),
  zone: self.new('zone', 'Zone', 'kuma_io_zone', 'envoy_server_live{mesh="$mesh"}', true, true),
}
