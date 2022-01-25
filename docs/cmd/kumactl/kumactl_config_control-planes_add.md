## kumactl config control-planes add

Add a Control Plane

### Synopsis

Add a Control Plane.

```
kumactl config control-planes add [flags]
```

### Options

```
      --address string             URL of the Control Plane API Server (required). Example: http://localhost:5681 or https://localhost:5682)
      --auth-conf stringToString   authentication configuration for defined authentication type format key=value (default [])
      --auth-type string           authentication type (for example: "tokens")
      --ca-cert-file string        path to the certificate authority which will be used to verify the Control Plane certificate (kumactl stores only a reference to this file)
      --client-cert-file string    path to the certificate of a client that is authorized to use the Admin operations of the Control Plane (kumactl stores only a reference to this file)
      --client-key-file string     path to the certificate key of a client that is authorized to use the Admin operations of the Control Plane (kumactl stores only a reference to this file)
      --headers stringToString     add these headers while communicating to control plane, format key=value (default [])
  -h, --help                       help for add
      --name string                reference name for the Control Plane (required)
      --overwrite                  overwrite existing Control Plane with the same reference name
      --skip-verify                skip CA verification
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

