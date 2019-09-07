# kumactl

```
Management tool for Kuma.

Usage:
  kumactl [command]

Available Commands:
  apply       Create or modify Kuma resources
  config      Manage kumactl config
  get         Show Kuma resources
  help        Help about any command
  inspect     Inspect Kuma resources
  install     Install Kuma on Kubernetes
  version     Print version

Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
  -h, --help                 help for kumactl
      --mesh string          mesh to use

Use "kumactl [command] --help" for more information about a command.
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
      --debug                enable debug-level logging (default true)
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
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use
```

### kumactl config control-planes

```
Manage known Control Planes.

Usage:
  kumactl config control-planes [command]

Available Commands:
  add         Add a Control Plane
  list        List known Control Planes

Flags:
  -h, --help   help for control-planes

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use

Use "kumactl config control-planes [command] --help" for more information about a command.
```

#### kumactl config control-planes list

```
List known Control Planes.

Usage:
  kumactl config control-planes list [flags]

Flags:
  -h, --help   help for list

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use
```

#### kumactl config control-planes add

```
Add a Control Plane.

Usage:
  kumactl config control-planes add [flags]

Flags:
      --api-server-url string   URL of the Control Plane API Server (required)
  -h, --help                    help for add
      --name string             reference name for the Control Plane (required)

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use
```

## kumactl install

```
Install Kuma on Kubernetes.

Usage:
  kumactl install [command]

Available Commands:
  control-plane Install Kuma Control Plane on Kubernetes

Flags:
  -h, --help   help for install

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use

Use "kumactl install [command] --help" for more information about a command.
```

## kumactl get

```
Show Kuma resources.

Usage:
  kumactl get [command]

Available Commands:
  meshes              Show Meshes
  proxytemplates      Show ProxyTemplates
  traffic-permissions Show TrafficPermissions

Flags:
  -h, --help            help for get
  -o, --output string   Output format: one of table|yaml|json (default "table")

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
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
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use
  -o, --output string        Output format: one of table|yaml|json (default "table")
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
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use
  -o, --output string        Output format: one of table|yaml|json (default "table")
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
  -o, --output string   Output format: one of table|yaml|json (default "table")

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
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
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use
  -o, --output string        Output format: one of table|yaml|json (default "table")
```

