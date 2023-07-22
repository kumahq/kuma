local g = import 'g.libsonnet';
local dashboard = g.dashboard;
local annotation = g.dashboard.annotation;

annotation.datasource.withType('datasource')
+ annotation.datasource.withUid('grafana')
+ annotation.withEnable(true)
+ annotation.withHide(true)
+ annotation.withIconColor('rgba(0, 211, 255, 1)')
+ annotation.withName('Annotations & Alerts')
+ annotation.withType('dashboard')
+ annotation.target.withLimit(100)
+ annotation.target.withMatchAny(false)
+ annotation.target.withTags([])
+ annotation.target.withType('dashboard')
