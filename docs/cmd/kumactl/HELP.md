# kumactl

```
Management tool for Kuma.

Usage:
  kumactl [command]

Available Commands:
  apply       Create or modify Kuma resources
  config      Manage kumactl config
  generate    Generate resources
  get         Show Kuma resources
  help        Help about any command
  inspect     Inspect Kuma resources
  install     Install Kuma on Kubernetes
  version     Print version

Flags:
      --config-file string   path to the configuration file to use
  -h, --help                 help for kumactl
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use

Use "kumactl [command] --help" for more information about a command.
```

## kumactl apply

```
Create or modify Kuma resources.

Usage:
  kumactl apply [flags]

Flags:
  -f, --file string          Path to file to apply
  -h, --help                 help for apply
  -v, --var stringToString   Variable to replace in configuration (default [])

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
```

## kumactl config

```
Manage kumactl config.

Usage:
  kumactl config [command]

Available Commands:
  control-planes Manage known Control Planes
  view           Show kumactl config

Flags:
  -h, --help   help for config

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use

Use "kumactl config [command] --help" for more information about a command.
```

### kumactl config view

```
Show kumactl config.

Usage:
  kumactl config view [flags]

Flags:
  -h, --help   help for view

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
```

### kumactl config control-planes

```
Manage known Control Planes.

Usage:
  kumactl config control-planes [command]

Available Commands:
  add         Add a Control Plane
  list        List Control Planes
  remove      Remove a Control Plane
  switch      Switch active Control Plane

Flags:
  -h, --help   help for control-planes

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use

Use "kumactl config control-planes [command] --help" for more information about a command.
```

#### kumactl config control-planes list

```
List Control Planes.

Usage:
  kumactl config control-planes list [flags]

Flags:
  -h, --help   help for list

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
```

#### kumactl config control-planes add

```
Add a Control Plane.

Usage:
  kumactl config control-planes add [flags]

Flags:
      --address string   URL of the Control Plane API Server (required)
  -h, --help             help for add
      --name string      reference name for the Control Plane (required)

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
```

#### kumactl config control-planes remove

```
Remove a Control Plane.

Usage:
  kumactl config control-planes remove [flags]

Flags:
  -h, --help          help for remove
      --name string   reference name for the Control Plane (required)

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
```

#### kumactl config control-planes switch

```
Switch active Control Plane.

Usage:
  kumactl config control-planes switch [flags]

Flags:
  -h, --help          help for switch
      --name string   reference name for the Control Plane (required)

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
```

## kumactl install

```
Install Kuma on Kubernetes.

Usage:
  kumactl install [command]

Available Commands:
  control-plane   Install Kuma Control Plane on Kubernetes
  database-schema Install Kuma schema on DB

Flags:
  -h, --help   help for install

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use

Use "kumactl install [command] --help" for more information about a command.
```

### kumactl install control-plane

```
Install Kuma Control Plane on Kubernetes.

Usage:
  kumactl install control-plane [flags]

Flags:
      --admission-server-tls-cert string    TLS certificate for the admission web hooks implemented by the Kuma Control Plane
      --admission-server-tls-key string     TLS key for the admission web hooks implemented by the Kuma Control Plane
      --control-plane-image string          image of the Kuma Control Plane component (default "kong-docker-kuma-docker.bintray.io/kuma-cp")
      --control-plane-service-name string   Service name of the Kuma Control Plane (default "kuma-control-plane")
      --control-plane-version string        version shared by all components of the Kuma Control Plane (default "latest")
      --dataplane-image string              image of the Kuma Dataplane component (default "kong-docker-kuma-docker.bintray.io/kuma-dp")
      --dataplane-init-image string         init image of the Kuma Dataplane component (default "docker.io/istio/proxy_init")
      --dataplane-init-version string       version of the init image of the Kuma Dataplane component (default "1.1.2")
  -h, --help                                help for control-plane
      --image-pull-policy string            image pull policy that applies to all components of the Kuma Control Plane (default "IfNotPresent")
      --injector-failure-policy string      failue policy of the mutating web hook implemented by the Kuma Injector component (default "Ignore")
      --injector-image string               image of the Kuma Injector component (default "kong-docker-kuma-docker.bintray.io/kuma-injector")
      --injector-service-name string        Service name of the mutating web hook implemented by the Kuma Injector component (default "kuma-injector")
      --injector-tls-cert string            TLS certificate for the mutating web hook implemented by the Kuma Injector component
      --injector-tls-key string             TLS key for the mutating web hook implemented by the Kuma Injector component
      --namespace string                    namespace to install Kuma Control Plane to (default "kuma-system")
      --sds-tls-cert string                 TLS certificate for the SDS server
      --sds-tls-key string                  TLS key for the SDS server

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
```

### kumactl install database-schema

```
Install Kuma schema on DB.

Usage:
  kumactl install database-schema [flags]

Examples:
1. kumactl install database-schema --target=postgres | PGPASSWORD=mysecretpassword psql -h localhost -U postgres
2. sql_file=$(mktemp) ; \ 
kumactl install database-schema --target=postgres >$sql_file ; \
psql --host=localhost --username=postgres --password --file=$sql_file ; \
rm $sql_file

Flags:
  -h, --help            help for database-schema
      --target string   Database type: one of postgres (default "postgres")

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
```

### kumactl generate dp-token

```
Generate Dataplane Token that is used to prove Dataplane identity.

Usage:
  kumactl generate dp-token [flags]

Flags:
  -h, --help          help for dp-token
      --name string   name of the Dataplane

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
```

## kumactl get

```
Show Kuma resources.

Usage:
  kumactl get [command]

Available Commands:
  dataplanes          Show Dataplanes
  meshes              Show Meshes
  proxytemplates      Show ProxyTemplates
  traffic-logs        Show TrafficLogs
  traffic-permissions Show TrafficPermissions

Flags:
  -h, --help            help for get
  -o, --output string   output format: one of table|yaml|json (default "table")

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use

Use "kumactl get [command] --help" for more information about a command.
```

### kumactl get meshes

```
Show Meshes.

Usage:
  kumactl get meshes [flags]

Flags:
  -h, --help   help for meshes

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get dataplanes

```
Show Dataplanes.

Usage:
  kumactl get dataplanes [flags]

Flags:
  -h, --help   help for dataplanes

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get proxytemplates

```
Show ProxyTemplates.

Usage:
  kumactl get proxytemplates [flags]

Flags:
  -h, --help   help for proxytemplates

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get traffic-permissions

```
Show TrafficPermission entities.

Usage:
  kumactl get traffic-permissions [flags]

Flags:
  -h, --help   help for traffic-permissions

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
  -o, --output string        output format: one of table|yaml|json (default "table")
```

## kumactl inspect

```
Inspect Kuma resources.

Usage:
  kumactl inspect [command]

Available Commands:
  dataplanes  Inspect Dataplanes

Flags:
  -h, --help            help for inspect
  -o, --output string   output format: one of table|yaml|json (default "table")

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use

Use "kumactl inspect [command] --help" for more information about a command.
```

### kumactl inspect dataplanes

```
Inspect Dataplanes.

Usage:
  kumactl inspect dataplanes [flags]

Flags:
  -h, --help                 help for dataplanes
      --tag stringToString   filter by tag in format of key=value. You can provide many tags (default [])

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
  -o, --output string        output format: one of table|yaml|json (default "table")
```

## kumactl version

```
Print version.

Usage:
  kumactl version [flags]

Flags:
  -a, --detailed   Print detailed version
  -h, --help       help for version

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
      --mesh string          mesh to use
```

