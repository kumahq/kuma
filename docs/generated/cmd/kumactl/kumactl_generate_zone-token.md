## kumactl generate zone-token

Generate Zone Token

### Synopsis

Generate Zone Token that is used to prove identity of zone components (Zone Ingress, Zone Egress).

```
kumactl generate zone-token [flags]
```

### Examples

```
Generate token bound by zone
$ kumactl generate zone-token --zone zone-1 --valid-for 24h
$ kumactl generate zone-token --zone zone-1 --valid-for 24h --scope egress
$ kumactl generate zone-token --zone zone-1 --valid-for 24h --scope ingress
$ kumactl generate zone-token --zone zone-1 --valid-for 24h --scope ingress --scope egress
```

### Options

```
  -h, --help                 help for zone-token
      --scope strings        scope of resources which the token will be able to identify (can be: [ingress egress]) (default [ingress,egress])
      --valid-for duration   how long the token will be valid (for example "24h")
      --zone string          name of the zone where resides
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
      --no-config              if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl generate](kumactl_generate.md)	 - Generate resources, tokens, etc

