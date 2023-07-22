local g = import 'g.libsonnet';
local dashboard = g.dashboard;
local annotation = import 'annotation.jsonnet';

{
  new(title):
    dashboard.new(title)
    + dashboard.withAnnotations([annotation])
    + dashboard.timepicker.withRefreshIntervals(['5s', '10s', '30s', '1m', '5m', '15m', '30m', '1h', '2h', '1d'])
    + dashboard.withFiscalYearStartMonth(0)
    + dashboard.withGraphTooltip(0)
    + dashboard.withEditable(true)
    + dashboard.withLiveNow(false)
    + dashboard.withLinks([])
    + dashboard.withRefresh('5s')
    + dashboard.withStyle('dark')
    + dashboard.withTags([])
    + dashboard.withVersion(1),

  withDescription(description):
    dashboard.withDescription(description),

  withPanels(panels):
    dashboard.withPanels(panels),

  withVariables(variables):
    dashboard.withVariables(variables),
}
