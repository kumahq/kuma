## kumactl generate dataplane-token

Generate Dataplane Token

### Synopsis

Generate Dataplane Token that is used to prove Dataplane identity.

```
kumactl generate dataplane-token [flags]
```

### Examples

```

Generate token bound by name and mesh
$ kumactl generate dataplane-token --mesh demo --name demo-01

Generate token bound by mesh
$ kumactl generate dataplane-token --mesh demo

Generate Ingress token
$ kumactl generate dataplane-token --type ingress

Generate token bound by tag
$ kumactl generate dataplane-token --mesh demo --tag kuma.io/service=web,web-api

```

### Options

```
  -h, --help                 help for dataplane-token
      --name string          name of the Dataplane
      --proxy-type string    type of the Dataplane ("dataplane", "ingress")
      --tag stringToString   required tag values for dataplane (split values by comma to provide multiple values) (default [])
```

### Options inherited from parent commands

```
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl generate](kumactl_generate.md)	 - Generate resources, tokens, etc

