# konvoyctl

```
Management tool for Konvoy Service Mesh.

Usage:
  konvoyctl [command]

Available Commands:
  apply       Create or modify Konvoy resources
  config      Manage konvoyctl config
  get         Show Konvoy resources
  help        Help about any command
  install     Install Konvoy on Kubernetes

Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
  -h, --help                 help for konvoyctl
      --mesh string          mesh to use

Use "konvoyctl [command] --help" for more information about a command.
```

## konvoyctl config

```
Manage konvoyctl config.

Usage:
  konvoyctl config [command]

Available Commands:
  control-planes Manage known Control Planes
  view           Show konvoyctl config

Flags:
  -h, --help   help for config

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use

Use "konvoyctl config [command] --help" for more information about a command.
```

### konvoyctl config view

```
Show konvoyctl config.

Usage:
  konvoyctl config view [flags]

Flags:
  -h, --help   help for view

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use
```

### konvoyctl config control-planes

```
Manage known Control Planes.

Usage:
  konvoyctl config control-planes [command]

Available Commands:
  add         Add a Control Plane
  list        List known Control Planes

Flags:
  -h, --help   help for control-planes

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use

Use "konvoyctl config control-planes [command] --help" for more information about a command.
```

#### konvoyctl config control-planes list

```
List known Control Planes.

Usage:
  konvoyctl config control-planes list [flags]

Flags:
  -h, --help   help for list

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use
```

#### konvoyctl config control-planes add

```
Add a Control Plane.

Usage:
  konvoyctl config control-planes add [flags]

Flags:
      --api-server-url string   URL of the Control Plane API Server (required)
  -h, --help                    help for add
      --name string             reference name for the Control Plane (required)

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use
```

## konvoyctl install

```
Install Konvoy on Kubernetes.

Usage:
  konvoyctl install [command]

Available Commands:
  control-plane Install Konvoy Control Plane on Kubernetes

Flags:
  -h, --help   help for install

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use

Use "konvoyctl install [command] --help" for more information about a command.
```

## konvoyctl get

```
Show Konvoy resources.

Usage:
  konvoyctl get [command]

Available Commands:
  dataplanes     Show running Dataplanes
  meshes         Show Meshes
  proxytemplates Show ProxyTemplates

Flags:
  -h, --help            help for get
  -o, --output string   Output format: one of table|yaml|json (default "table")

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use

Use "konvoyctl get [command] --help" for more information about a command.
```

### konvoyctl get meshes

```
Show Meshes.

Usage:
  konvoyctl get meshes [flags]

Flags:
  -h, --help   help for meshes

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use
  -o, --output string        Output format: one of table|yaml|json (default "table")
```

### konvoyctl get proxytemplates

```
Show ProxyTemplates.

Usage:
  konvoyctl get proxytemplates [flags]

Flags:
  -h, --help   help for proxytemplates

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use
  -o, --output string        Output format: one of table|yaml|json (default "table")
```

### konvoyctl get dataplanes

```
Show running Dataplanes.

Usage:
  konvoyctl get dataplanes [flags]

Flags:
  -h, --help                 help for dataplanes
      --tag stringToString   filter by tag in format of key=value. You can provide many tags (default [])

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --mesh string          mesh to use
  -o, --output string        Output format: one of table|yaml|json (default "table")
```

