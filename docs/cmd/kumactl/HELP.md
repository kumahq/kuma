# kumactl

```
Management tool for Kuma.

Usage:
  kumactl [command]

Available Commands:
  apply       Create or modify Kuma resources
  completion  Output shell completion code for bash, fish or zsh
  config      Manage kumactl config
  delete      Delete Kuma resources
  generate    Generate resources, tokens, etc
  get         Show Kuma resources
  help        Help about any command
  inspect     Inspect Kuma resources
  install     Install Kuma on Kubernetes
  version     Print version

Flags:
      --config-file string   path to the configuration file to use
  -h, --help                 help for kumactl
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")

Use "kumactl [command] --help" for more information about a command.
```

## kumactl apply

```
Create or modify Kuma resources.

Usage:
  kumactl apply [flags]

Flags:
      --dry-run              Resolve variable and prints result out without actual applying
  -f, --file string          Path to file to apply
  -h, --help                 help for apply
  -v, --var stringToString   Variable to replace in configuration (default [])

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
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
  -m, --mesh string          mesh to use (default "default")

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
  -m, --mesh string          mesh to use (default "default")
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
  -m, --mesh string          mesh to use (default "default")

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
  -m, --mesh string          mesh to use (default "default")
```

#### kumactl config control-planes add

```
Add a Control Plane.

Usage:
  kumactl config control-planes add [flags]

Flags:
      --address string             URL of the Control Plane API Server (required)
      --admin-client-cert string   Path to certificate of a client that is authorized to use Admin Server
      --admin-client-key string    Path to certificate key of a client that is authorized to use Admin Server
  -h, --help                       help for add
      --name string                reference name for the Control Plane (required)
      --overwrite                  overwrite existing Control Plane with the same reference name

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
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
  -m, --mesh string          mesh to use (default "default")
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
  -m, --mesh string          mesh to use (default "default")
```

## kumactl install

```
Install Kuma on Kubernetes.

Usage:
  kumactl install [command]

Available Commands:
  control-plane Install Kuma Control Plane on Kubernetes
  dns           Install DNS to Kubernetes
  ingress       Install Ingress on Kubernetes
  logging       Install Logging backend in Kubernetes cluster (Loki)
  metrics       Install Metrics backend in Kubernetes cluster (Prometheus + Grafana)
  tracing       Install Tracing backend in Kubernetes cluster (Jaeger)

Flags:
  -h, --help   help for install

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")

Use "kumactl install [command] --help" for more information about a command.
```

### kumactl install control-plane

```
Install Kuma Control Plane on Kubernetes in a 'kuma-system' namespace.

Usage:
  kumactl install control-plane [flags]

Flags:
      --cni-enabled                         install Kuma with CNI instead of proxy init container
      --cni-registry string                 registry for the image of the Kuma CNI component (default "docker.io")
      --cni-repository string               repository for the image of the Kuma CNI component (default "lobkovilya/install-cni")
      --cni-version string                  version of the image of the Kuma CNI component (default "0.0.1")
      --control-plane-registry string       registry for the image of the Kuma Control Plane component (default "kong-docker-kuma-docker.bintray.io")
      --control-plane-repository string     repository for the image of the Kuma Control Plane component (default "kuma-cp")
      --control-plane-service-name string   Service name of the Kuma Control Plane (default "kuma-control-plane")
      --control-plane-version string        version of the image of the Kuma Control Plane component (default "latest")
      --dataplane-init-registry string      registry for the init image of the Kuma DataPlane component (default "kong-docker-kuma-docker.bintray.io")
      --dataplane-init-repository string    repository for the init image of the Kuma DataPlane component (default "kuma-init")
      --dataplane-init-version string       version of the init image of the Kuma DataPlane component (default "latest")
      --dataplane-registry string           registry for the image of the Kuma DataPlane component (default "kong-docker-kuma-docker.bintray.io")
      --dataplane-repository string         repository for the image of the Kuma DataPlane component (default "kuma-dp")
      --dataplane-version string            version of the image of the Kuma DataPlane component (default "latest")
  -h, --help                                help for control-plane
      --image-pull-policy string            image pull policy that applies to all components of the Kuma Control Plane (default "IfNotPresent")
      --injector-failure-policy string      failue policy of the mutating web hook implemented by the Kuma Injector component (default "Ignore")
      --kds-global-address string           URL of Global Kuma CP (example: grpcs://192.168.0.1:5685)
      --mode string                         kuma cp modes: one of standalone|remote|global (default "standalone")
      --namespace string                    namespace to install Kuma Control Plane to (default "kuma-system")
      --tls-cert string                     TLS certificate for Kuma Control Plane servers
      --tls-key string                      TLS key for Kuma Control Plane servers
      --use-node-port                       use NodePort instead of LoadBalancer
      --zone string                         set the Kuma zone name

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
```

### kumactl install metrics

```
Install Metrics backend in Kubernetes cluster (Prometheus + Grafana) in a kuma-metrics namespace

Usage:
  kumactl install metrics [flags]

Flags:
  -h, --help                                help for metrics
      --kuma-cp-address string              the address of Kuma CP (default "http://kuma-control-plane.kuma-system:5681")
      --kuma-prometheus-sd-image string     image name of Kuma Prometheus SD (default "kong-docker-kuma-docker.bintray.io/kuma-prometheus-sd")
      --kuma-prometheus-sd-version string   version of Kuma Prometheus SD (default "latest")
      --namespace string                    namespace to install metrics to (default "kuma-metrics")

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
```

### kumactl install tracing

```
Install Tracing backend in Kubernetes cluster (Jaeger) in a 'kuma-tracing' namespace

Usage:
  kumactl install tracing [flags]

Flags:
  -h, --help               help for tracing
      --namespace string   namespace to install tracing to (default "kuma-tracing")

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
```

### kumactl generate tls-certificate

```
Generate self signed key and certificate pair that can be used for example in Dataplane Token Server setup.

Usage:
  kumactl generate tls-certificate [flags]

Examples:

  # Generate a TLS certificate for use by an HTTPS server, i.e. by the Dataplane Token server
  kumactl generate tls-certificate --type=server

  # Generate a TLS certificate for use by a client of an HTTPS server, i.e. by the 'kumactl generate dataplane-token' command
  kumactl generate tls-certificate --type=client

Flags:
      --cert-file string      path to a file with a generated TLS certificate (default "cert.pem")
      --cp-hostname strings   DNS name of the control plane
  -h, --help                  help for tls-certificate
      --key-file string       path to a file with a generated private key (default "key.pem")
      --type string           type of the certificate: one of client|server

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
```

### kumactl generate dp-token

```
Generate resources, tokens, etc.

Usage:
  kumactl generate [command]

Available Commands:
  dataplane-token Generate Dataplane Token
  tls-certificate Generate a TLS certificate

Flags:
  -h, --help   help for generate

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")

Use "kumactl generate [command] --help" for more information about a command.
```

## kumactl get

```
Show Kuma resources.

Usage:
  kumactl get [command]

Available Commands:
  circuit-breaker     Show a single CircuitBreaker resource
  circuit-breakers    Show CircuitBreakers
  dataplane           Show a single Dataplane resource
  dataplanes          Show Dataplanes
  fault-injection     Show a single Fault-Injection resource
  fault-injections    Show FaultInjections
  healthcheck         Show a single HealthCheck resource
  healthchecks        Show HealthChecks
  mesh                Show a single Mesh resource
  meshes              Show Meshes
  proxytemplate       Show a single Proxytemplate resource
  proxytemplates      Show ProxyTemplates
  secret              Show a single Secret resource
  secrets             Show Secrets
  traffic-log         Show a single TrafficLog resource
  traffic-logs        Show TrafficLogs
  traffic-permission  Show a single TrafficPermission resource
  traffic-permissions Show TrafficPermissions
  traffic-route       Show a single TrafficRoute resource
  traffic-routes      Show TrafficRoutes
  traffic-trace       Show a single TrafficTrace resource
  traffic-traces      Show TrafficTraces
  zone                Show a single Zone resource
  zones               Show Zones

Flags:
  -h, --help            help for get
  -o, --output string   output format: one of table|yaml|json (default "table")

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")

Use "kumactl get [command] --help" for more information about a command.
```

### kumactl get meshes

```
Show Meshes.

Usage:
  kumactl get meshes [flags]

Flags:
  -h, --help            help for meshes
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get dataplanes

```
Show Dataplanes.

Usage:
  kumactl get dataplanes [flags]

Flags:
  -h, --help            help for dataplanes
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get healthchecks

```
Show HealthChecks.

Usage:
  kumactl get healthchecks [flags]

Flags:
  -h, --help            help for healthchecks
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get proxytemplates

```
Show ProxyTemplates.

Usage:
  kumactl get proxytemplates [flags]

Flags:
  -h, --help            help for proxytemplates
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get traffic-logs

```
Show TrafficLog entities.

Usage:
  kumactl get traffic-logs [flags]

Flags:
  -h, --help            help for traffic-logs
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get traffic-permissions

```
Show TrafficPermission entities.

Usage:
  kumactl get traffic-permissions [flags]

Flags:
  -h, --help            help for traffic-permissions
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get traffic-routes

```
Show TrafficRoutes.

Usage:
  kumactl get traffic-routes [flags]

Flags:
  -h, --help            help for traffic-routes
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get traffic-traces

```
Show TrafficTrace entities.

Usage:
  kumactl get traffic-traces [flags]

Flags:
  -h, --help            help for traffic-traces
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get fault-injections

```
Show FaultInjection entities.

Usage:
  kumactl get fault-injections [flags]

Flags:
  -h, --help            help for fault-injections
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get secret

```
Show a single Secret resource.

Usage:
  kumactl get secret NAME [flags]

Flags:
  -h, --help   help for secret

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get secrets

```
Show Secrets.

Usage:
  kumactl get secrets [flags]

Flags:
  -h, --help   help for secrets

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get zones

```
Show Zone entities.

Usage:
  kumactl get zones [flags]

Flags:
  -h, --help            help for zones
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

## kumactl delete

```
Delete Kuma resources.

Usage:
  kumactl delete TYPE NAME [flags]

Flags:
  -h, --help   help for delete

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
```

## kumactl inspect

```
Inspect Kuma resources.

Usage:
  kumactl inspect [command]

Available Commands:
  dataplanes  Inspect Dataplanes
  zones       Inspect Zones

Flags:
  -h, --help            help for inspect
  -o, --output string   output format: one of table|yaml|json (default "table")

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")

Use "kumactl inspect [command] --help" for more information about a command.
```

### kumactl inspect dataplanes

```
Inspect Dataplanes.

Usage:
  kumactl inspect dataplanes [flags]

Flags:
      --gateway              filter gateway dataplanes
  -h, --help                 help for dataplanes
      --tag stringToString   filter by tag in format of key=value. You can provide many tags (default [])

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl inspect zones

```
Inspect Zones.

Usage:
  kumactl inspect zones [flags]

Flags:
  -h, --help   help for zones

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
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
  -m, --mesh string          mesh to use (default "default")
```

