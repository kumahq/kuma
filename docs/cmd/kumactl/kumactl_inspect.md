## kumactl inspect

Inspect Kuma resources

### Synopsis

Inspect Kuma resources.

### Options

```
  -h, --help            help for inspect
  -o, --output string   output format: one of table|yaml|json (default "table")
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
  -m, --mesh string            mesh to use (default "default")
      --no-config              if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl](kumactl.md)	 - Management tool for Kuma
* [kumactl inspect dataplanes](kumactl_inspect_dataplanes.md)	 - Inspect Dataplanes
* [kumactl inspect meshes](kumactl_inspect_meshes.md)	 - Inspect Meshes
* [kumactl inspect services](kumactl_inspect_services.md)	 - Inspect Services
* [kumactl inspect zone-ingresses](kumactl_inspect_zone-ingresses.md)	 - Inspect Zone Ingresses
* [kumactl inspect zoneegresses](kumactl_inspect_zoneegresses.md)	 - Inspect Zone Egresses
* [kumactl inspect zones](kumactl_inspect_zones.md)	 - Inspect Zones

