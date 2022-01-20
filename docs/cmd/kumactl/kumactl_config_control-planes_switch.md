## kumactl config control-planes switch

Switch active Control Plane

### Synopsis

Switch active Control Plane.

```
kumactl config control-planes switch [flags]
```

### Options

```
  -h, --help          help for switch
      --name string   reference name for the Control Plane (required)
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

* [kumactl config control-planes](kumactl_config_control-planes.md)	 - Manage known Control Planes

