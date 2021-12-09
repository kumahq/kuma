## kuma-prometheus-sd run

Launch Kuma Prometheus SD adapter

### Synopsis

Launch Kuma Prometheus SD adapter.

```
kuma-prometheus-sd run [flags]
```

### Options

```
      --api-version string    MADS API version to request from the Monitoring Assignment server. (default "v1")
      --cp-address string     URL of the Control Plane Monitoring Assignment Discovery Server. Example: grpc://localhost:5676 (default "grpc://localhost:5676")
  -h, --help                  help for run
      --name string           Name to use to identify itself to the Monitoring Assignment server. (default "kuma_sd")
      --output-file file_sd   Path to an output file with a list of scrape targets. The same file path must be used on Prometheus side in a configuration of file_sd discovery mechanism. (default "kuma.file_sd.json")
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

* [kuma-prometheus-sd](kuma-prometheus-sd.md)	 - [DEPRECATED] Prometheus service discovery adapter for native integration with Kuma

