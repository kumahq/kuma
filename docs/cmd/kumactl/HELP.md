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
  install     Install various Kuma components.
  uninstall   Uninstall various Kuma components.
  version     Print version

Flags:
      --config-file string   path to the configuration file to use
  -h, --help                 help for kumactl
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created

Use "kumactl [command] --help" for more information about a command.
```

## kumactl apply

```
Create or modify Kuma resources.

Usage:
  kumactl apply [flags]

Examples:

Apply a resource from file
$ kumactl apply -f resource.yaml

Apply a resource from stdin
$ echo "
type: Mesh
name: demo
" | kumactl apply -f -

Apply a resource from external URL
$ kumactl apply -f https://example.com/resource.yaml


Flags:
      --dry-run              Resolve variable and prints result out without actual applying
  -f, --file -               Path to file to apply. Pass - to read from stdin
  -h, --help                 help for apply
  -v, --var stringToString   Variable to replace in configuration (default [])

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created

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
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created

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
      --no-config            if set no config file and config directory will be created
```

#### kumactl config control-planes add

```
Add a Control Plane.

Usage:
  kumactl config control-planes add [flags]

Flags:
      --address string            URL of the Control Plane API Server (required). Example: http://localhost:5681 or https://localhost:5682)
      --ca-cert-file string       path to the certificate authority which will be used to verify the Control Plane certificate (kumactl stores only a reference to this file)
      --client-cert-file string   path to the certificate of a client that is authorized to use the Admin operations of the Control Plane (kumactl stores only a reference to this file)
      --client-key-file string    path to the certificate key of a client that is authorized to use the Admin operations of the Control Plane (kumactl stores only a reference to this file)
      --headers stringToString    add these headers while communicating to control plane, format key=value (default [])
  -h, --help                      help for add
      --name string               reference name for the Control Plane (required)
      --overwrite                 overwrite existing Control Plane with the same reference name
      --skip-verify               skip CA verification

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
```

## kumactl install

```
Install various Kuma components.

Usage:
  kumactl install [command]

Available Commands:
  control-plane     Install Kuma Control Plane on Kubernetes
  crds              Install Kuma Custom Resource Definitions on Kubernetes
  demo              Install Kuma demo on Kubernetes
  dns               Install DNS to Kubernetes
  gateway           Install ingress gateway on Kubernetes
  logging           Install Logging backend in Kubernetes cluster (Loki)
  metrics           Install Metrics backend in Kubernetes cluster (Prometheus + Grafana)
  tracing           Install Tracing backend in Kubernetes cluster (Jaeger)
  transparent-proxy Install Transparent Proxy pre-requisites on the host

Flags:
  -h, --help   help for install

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created

Use "kumactl install [command] --help" for more information about a command.
```

### kumactl install control-plane

```
Install Kuma Control Plane on Kubernetes in a 'kuma-system' namespace.
This command requires that the KUBECONFIG environment is set

Usage:
  kumactl install control-plane [flags]

Flags:
      --cni-bin-dir string                          set the CNI binary directory (default "/var/lib/cni/bin")
      --cni-chained                                 enable chained CNI installation
      --cni-conf-name string                        set the CNI configuration name (default "kuma-cni.conf")
      --cni-enabled                                 install Kuma with CNI instead of proxy init container
      --cni-net-dir string                          set the CNI install directory (default "/etc/cni/multus/net.d")
      --cni-registry string                         registry for the image of the Kuma CNI component (default "docker.io/lobkovilya")
      --cni-repository string                       repository for the image of the Kuma CNI component (default "install-cni")
      --cni-version string                          version of the image of the Kuma CNI component (default "0.0.8")
      --control-plane-registry string               registry for the image of the Kuma Control Plane component (default "docker.io/kumahq")
      --control-plane-repository string             repository for the image of the Kuma Control Plane component (default "kuma-cp")
      --control-plane-service-name string           Service name of the Kuma Control Plane (default "kuma-control-plane")
      --control-plane-version string                version of the image of the Kuma Control Plane component (default "latest")
      --dataplane-init-registry string              registry for the init image of the Kuma DataPlane component (default "docker.io/kumahq")
      --dataplane-init-repository string            repository for the init image of the Kuma DataPlane component (default "kuma-init")
      --dataplane-init-version string               version of the init image of the Kuma DataPlane component (default "latest")
      --dataplane-registry string                   registry for the image of the Kuma DataPlane component (default "docker.io/kumahq")
      --dataplane-repository string                 repository for the image of the Kuma DataPlane component (default "kuma-dp")
      --dataplane-version string                    version of the image of the Kuma DataPlane component (default "latest")
      --env-var stringToString                      environment variables that will be passed to the control plane (default [])
  -h, --help                                        help for control-plane
      --image-pull-policy string                    image pull policy that applies to all components of the Kuma Control Plane (default "IfNotPresent")
      --ingress-drain-time string                   drain time for Envoy proxy (default "30s")
      --ingress-enabled                             install Kuma with an Ingress deployment, using the Data Plane image
      --ingress-use-node-port                       use NodePort instead of LoadBalancer for the Ingress Service
      --injector-failure-policy string              failue policy of the mutating web hook implemented by the Kuma Injector component (default "Ignore")
      --kds-global-address string                   URL of Global Kuma CP (example: grpcs://192.168.0.1:5685)
      --mode string                                 kuma cp modes: one of standalone|zone|global (default "standalone")
      --namespace string                            namespace to install Kuma Control Plane to (default "kuma-system")
      --tls-api-server-client-certs-secret string   Secret that contains list of .pem certificates that can access admin endpoints of Kuma API on HTTPS
      --tls-api-server-secret string                Secret that contains tls.crt, key.crt for protecting Kuma API on HTTPS
      --tls-general-ca-bundle string                Base64 encoded CA certificate (the same as in controlPlane.tls.general.secret#ca.crt)
      --tls-general-secret string                   Secret that contains tls.crt, key.crt and ca.crt for protecting Kuma in-cluster communication
      --tls-kds-global-server-secret string         Secret that contains tls.crt, key.crt for protecting cross cluster communication
      --tls-kds-zone-client-secret string           Secret that contains ca.crt which was used to sign KDS Global server. Used for CP verification
      --use-node-port                               use NodePort instead of LoadBalancer
      --without-kubernetes-connection               install without connection to Kubernetes cluster. This can be used for initial Kuma installation, but not for upgrades
      --zone string                                 set the Kuma zone name

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created
```

### kumactl install metrics

```
Install Metrics backend in Kubernetes cluster (Prometheus + Grafana) in a kuma-metrics namespace

Usage:
  kumactl install metrics [flags]

Flags:
  -h, --help                                help for metrics
      --kuma-cp-address string              the address of Kuma CP (default "grpc://kuma-control-plane.kuma-system:5676")
      --kuma-prometheus-sd-image string     image name of Kuma Prometheus SD (default "docker.io/kumahq/kuma-prometheus-sd")
      --kuma-prometheus-sd-version string   version of Kuma Prometheus SD (default "latest")
      --namespace string                    namespace to install metrics to (default "kuma-metrics")
      --without-grafana                     disable Grafana resources generation
      --without-prometheus                  disable Prometheus resources generation

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
```

### kumactl generate dataplane-token

```
Generate Dataplane Token that is used to prove Dataplane identity.

Usage:
  kumactl generate dataplane-token [flags]

Examples:

Generate token bound by name and mesh
$ kumactl generate dataplane-token --mesh demo --name demo-01

Generate token bound by mesh
$ kumactl generate dataplane-token --mesh demo

Generate Ingress token
$ kumactl generate dataplane-token --type ingress

Generate token bound by tag
$ kumactl generate dataplane-token --mesh demo --tag kuma.io/service=web,web-api


Flags:
  -h, --help                 help for dataplane-token
      --name string          name of the Dataplane
      --tag stringToString   required tag values for dataplane (split values by comma to provide multiple values) (default [])
      --type string          type of the Dataplane ("dataplane", "ingress")

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created
```

## kumactl get

```
Show Kuma resources.

Usage:
  kumactl get [command]

Available Commands:
  circuit-breaker     Show a single CircuitBreaker resource
  circuit-breakers    Show CircuitBreaker
  dataplane           Show a single Dataplane resource
  dataplanes          Show Dataplane
  external-service    Show a single ExternalService resource
  external-services   Show ExternalService
  fault-injection     Show a single FaultInjection resource
  fault-injections    Show FaultInjection
  global-secret       Show a single GlobalSecret resource
  global-secrets      Show GlobalSecret
  healthcheck         Show a single HealthCheck resource
  healthchecks        Show HealthCheck
  mesh                Show a single Mesh resource
  meshes              Show Mesh
  proxytemplate       Show a single ProxyTemplate resource
  proxytemplates      Show ProxyTemplate
  rate-limit          Show a single RateLimit resource
  rate-limits         Show RateLimit
  retries             Show Retry
  retry               Show a single Retry resource
  secret              Show a single Secret resource
  secrets             Show Secret
  timeout             Show a single Timeout resource
  timeouts            Show Timeout
  traffic-log         Show a single TrafficLog resource
  traffic-logs        Show TrafficLog
  traffic-permission  Show a single TrafficPermission resource
  traffic-permissions Show TrafficPermission
  traffic-route       Show a single TrafficRoute resource
  traffic-routes      Show TrafficRoute
  traffic-trace       Show a single TrafficTrace resource
  traffic-traces      Show TrafficTrace
  zone                Show a single Retry resource
  zones               Show Zone

Flags:
  -h, --help            help for get
  -o, --output string   output format: one of table|yaml|json (default "table")

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created

Use "kumactl get [command] --help" for more information about a command.
```

### kumactl get meshes

```
Show Mesh entities.

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
      --no-config            if set no config file and config directory will be created
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get dataplanes

```
Show Dataplane entities.

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
      --no-config            if set no config file and config directory will be created
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get healthchecks

```
Show HealthCheck entities.

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
      --no-config            if set no config file and config directory will be created
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get rate-limits

```
Show RateLimit entities.

Usage:
  kumactl get rate-limits [flags]

Flags:
  -h, --help            help for rate-limits
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get retries

```
Show Retry entities.

Usage:
  kumactl get retries [flags]

Flags:
  -h, --help            help for retries
      --offset string   the offset that indicates starting element of the resources list to retrieve
      --size int        maximum number of elements to return

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get proxytemplates

```
Show ProxyTemplate entities.

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
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get traffic-routes

```
Show TrafficRoute entities.

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
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
  -o, --output string        output format: one of table|yaml|json (default "table")
```

### kumactl get secrets

```
Show Secret entities.

Usage:
  kumactl get secrets [flags]

Flags:
  -h, --help   help for secrets

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
```

## kumactl inspect

```
Inspect Kuma resources.

Usage:
  kumactl inspect [command]

Available Commands:
  dataplanes  Inspect Dataplanes
  meshes      Inspect Meshes
  services    Inspect Services
  zones       Inspect Zones

Flags:
  -h, --help            help for inspect
  -o, --output string   output format: one of table|yaml|json (default "table")

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created

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
      --ingress              filter ingress dataplanes
      --tag stringToString   filter by tag in format of key=value. You can provide many tags (default [])

Global Flags:
      --config-file string   path to the configuration file to use
      --log-level string     log level: one of off|info|debug (default "off")
  -m, --mesh string          mesh to use (default "default")
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
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
      --no-config            if set no config file and config directory will be created
```

