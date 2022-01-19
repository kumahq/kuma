## kumactl generate zone-egress-token

Generate Zone Egress Token

### Synopsis

Generate Zone Egress Token that is used to prove Zone Egress identity.

```
kumactl generate zone-egress-token [flags]
```

### Examples

```

Generate token bound by zone
$ kumactl generate zone-egress-token --zone zone-1 --valid-for 24h

```

### Options

```
  -h, --help                 help for zone-egress-token
      --valid-for duration   how long the token will be valid (for example "24h")
      --zone string          name of the zone where egress resides
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

