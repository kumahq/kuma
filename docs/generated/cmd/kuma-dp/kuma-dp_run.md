## kuma-dp run

Launch Dataplane (Envoy)

### Synopsis

Launch Dataplane (Envoy).

```
kuma-dp run [flags]
```

### Options

```
      --binary-path string                        Binary path of Envoy executable (default "envoy")
      --ca-cert-file string                       Path to CA cert by which connection to the Control Plane will be verified if HTTPS is used
      --concurrency uint32                        Number of Envoy worker threads
      --config-dir string                         Directory in which Envoy config will be generated
      --cp-address string                         URL of the Control Plane Dataplane Server. Example: https://localhost:5678 (default "https://localhost:5678")
      --dataplane string                          Dataplane template to apply (YAML or JSON)
  -d, --dataplane-file string                     Path to Dataplane template to apply (YAML or JSON)
      --dataplane-token string                    Dataplane Token
      --dataplane-token-file string               Path to a file with dataplane token (use 'kumactl generate dataplane-token' to get one)
  -v, --dataplane-var stringToString              Variables to replace Dataplane template (default [])
      --dns-coredns-config-template-path string   A path to a CoreDNS config template.
      --dns-coredns-empty-port uint32             A port that always responds with empty NXDOMAIN respond. It is required to implement a fallback to a real DNS. (default 15055)
      --dns-coredns-path string                   A path to CoreDNS binary. (default "coredns")
      --dns-coredns-port uint32                   A port that handles DNS requests. When transparent proxy is enabled then iptables will redirect DNS traffic to this port. (default 15053)
      --dns-enabled                               If true then builtin DNS functionality is enabled and CoreDNS server is started (default true)
      --dns-envoy-port uint32                     A port that handles Virtual IP resolving by Envoy. CoreDNS should be configured that it first tries to use this DNS resolver and then the real one (default 15054)
      --dns-prometheus-port uint32                A port for exposing Prometheus stats (default 19153)
      --dns-server-config-dir string              Directory in which DNS Server config will be generated
  -h, --help                                      help for run
      --mesh string                               Mesh that Dataplane belongs to
      --name string                               Name of the Dataplane
      --proxy-type string                         type of the Dataplane ("dataplane", "ingress") (default "dataplane")
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

