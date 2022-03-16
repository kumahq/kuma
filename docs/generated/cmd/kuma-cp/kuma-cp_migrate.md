## kuma-cp migrate

Migrate database to which Control Plane is connected

### Synopsis

Migrate database to which Control Plane is connected. The database contains all policies, dataplanes and secrets. The schema has to be in sync with version of Kuma CP to properly work. Make sure to run "kuma-cp migrate up" before running new version of Kuma.

### Options

```
  -h, --help   help for migrate
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

* [kuma-cp](kuma-cp.md)	 - Universal Control Plane for Envoy-based Service Mesh
* [kuma-cp migrate up](kuma-cp_migrate_up.md)	 - Apply the newest schema changes to the database.

