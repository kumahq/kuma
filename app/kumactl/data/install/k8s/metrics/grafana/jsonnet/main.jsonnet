local control_plane = import 'control-plane.jsonnet';
local dataplane = import 'dataplane.jsonnet';
local gateway = import 'gateway.jsonnet';
local mesh = import 'mesh.jsonnet';
local service_to_service = import 'service-to-service.jsonnet';
local service = import 'service.jsonnet';

local dashboards = [control_plane, dataplane, gateway, mesh, service_to_service, service];

{
  [dashboard.fileName]: dashboard.definition
  for dashboard in dashboards
}
