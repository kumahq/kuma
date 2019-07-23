# konvoyctl

```
Management tool for Konvoy.

Usage:
  konvoyctl [command]

Available Commands:
  config      Manage konvoyctl config
  get         Show Konvoy resources
  help        Help about any command

Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
  -h, --help                 help for konvoyctl

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
```

#### konvoyctl config control-planes add

```
Add a Control Plane.

Usage:
  konvoyctl config control-planes add [command]

Available Commands:
  k8s         Add a Control Plane installed on Kubernetes

Flags:
  -h, --help          help for add
      --name string   reference name for a Control Plane

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)

Use "konvoyctl config control-planes add [command] --help" for more information about a command.
```

##### konvoyctl config control-planes add k8s

```
Add a Control Plane installed on Kubernetes.

Usage:
  konvoyctl config control-planes add k8s [flags]

Flags:
  -h, --help   help for k8s

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
      --name string          reference name for a Control Plane
```

## konvoyctl get

```
Show Konvoy resources.

Usage:
  konvoyctl get [command]

Available Commands:
  dataplanes  Show running Dataplanes

Flags:
  -h, --help            help for get
  -o, --output string   Output format: one of table|yaml|json (default "table")

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)

Use "konvoyctl get [command] --help" for more information about a command.
```

### konvoyctl get dataplanes

```
Show running Dataplanes.

Usage:
  konvoyctl get dataplanes [flags]

Flags:
  -h, --help   help for dataplanes

Global Flags:
      --config-file string   path to the configuration file to use
      --debug                enable debug-level logging (default true)
  -o, --output string        Output format: one of table|yaml|json (default "table")
```

