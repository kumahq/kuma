## kuma-dp wait

Waits for Dataplane to be ready

### Synopsis

Waits for Dataplane (Envoy) to be ready.

```
kuma-dp wait [flags]
```

### Options

```
      --check-frequency duration   frequency of checking if the Dataplane is ready (default 1s)
  -h, --help                       help for wait
      --request-timeout duration   requestTimeout defines timeout for the request to the Dataplane (default 500ms)
      --timeout duration           timeout defines how long waits for the Dataplane (default 1m0s)
      --url string                 url at which admin is exposed (default "http://localhost:9901/ready")
```

### Options inherited from parent commands

```
      --log-level string             log level: one of off|info|debug (default "info")
      --log-max-age int              maximum number of days to retain old log files based on the timestamp encoded in their filename (default 30)
      --log-max-retained-files int   maximum number of the old log files to retain (default 1000)
      --log-max-size int             maximum size in megabytes of a log file before it gets rotated (default 100)
      --log-output-path string       path to the file that will be filled with logs. Example: if we set it to /tmp/kuma.log then after the file is rotated we will have /tmp/kuma-2021-06-07T09-15-18.265.log
```

### SEE ALSO

* [kuma-dp](kuma-dp.md)	 - Dataplane manager for Envoy-based Service Mesh

