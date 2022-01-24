## kumactl generate zone-token

Generate Zone Token

### Synopsis

Generate Zone Token that is used to prove identity of Zone dataplanes, ingresses and egresses.

```
kumactl generate zone-token [flags]
```

### Examples

```
Generate token bound by zone
$ kumactl generate zone-token --zone zone-1 --valid-for 24h

Generate token which can be used to prove identity of both zone ingress and egress
$ kumactl generate zone-token --zone zone-1 --valid-for 24h --scope ingress,egress
$ kumactl generate zone-token --zone zone-1 --valid-for 24h --scope ingress --scope egress

Generate token which can be used to prove identity of dataplane
$ kumactl generate zone-token --zone zone-1 --valid-for 24h --scope dataplane
```

### Options

```
  -h, --help                 help for zone-token
      --scope scope...       scope of the token; can be any combination of 'dataplane', 'ingress', 'egress' (default [egress])
      --valid-for duration   how long the token will be valid (for example "24h")
      --zone string          name of the zone where resides
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

