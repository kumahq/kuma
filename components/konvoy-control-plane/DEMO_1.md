# Use cases

## k8s environment

### Prerequirements

* dev tools (installed by `make dev/tools`)
* `docker`
* `Minikube`
* `kubectl`

### Build all components (Docker images)

* `konvoy/konvoy-control-plane:latest` - `xDS` and `API` server
* `konvoy/konvoy-injector:latest` - injects `Envoy` side-car into `Pods`
* `konvoy/konvoy-dataplane:latest` - launches `Envoy`
* `konvoyctl` - command-line client for end users

```
make dev/tools
```

```
eval $(minikube docker-env) # start using Docker daemon of Minikube
```

```
make images # notice that images will be built inside Minikube
```

```
make build/konvoyctl # `konvoyctl` is built on host machine
```

```
export PATH=`pwd`/build/artifacts/konvoyctl:$PATH
```

### Install Control Plane

```
kubectl apply -k examples/minikube/konvoy-control-plane/
```

### Deploy Dataplane

```
kubectl apply -f examples/minikube/konvoy-demo/
```

### Get Status

```
mv ~/.konvoyctl/config{,.backup.$(date +%s)}
```

```
konvoyctl get dataplanes
```

```
konvoyctl config view
```

### `kubectl port-forward` route

```
kubectl port-forward -n konvoy-system $( kubectl -n konvoy-system get pods -l app=konvoy-control-plane -o=jsonpath='{.items[0].metadata.name}' ) 15681:5681
```

```
curl localhost:15681/meshes/default/dataplanes
```

```
konvoyctl config control-planes add universal --name minikube-port-forward --api-server-url=http://localhost:15681
```

```
konvoyctl config view
```

```
konvoyctl get dataplanes
```

### k8s-native route (deprecated)

```
konvoyctl config control-planes add k8s --name minikube
```

```
konvoyctl config view
```

```
konvoyctl get dataplanes
```

```
konvoyctl get dataplanes -oyaml
```

```
konvoyctl get dataplanes --mesh pilot
```

## VM environment (memory)

### Prerequirements

* dev tools (installed by `make dev/tools`)
* `docker-compose` (to run `Postgres`)

### Build all components

* `konvoy-control-plane` - `xDS` and `API` server
* `konvoy-dataplane` - launches `Envoy`
* `konvoyctl` - command-line client for end users

```
make dev/tools
```

```
make build
```

```
export PATH=`pwd`/build/artifacts/konvoy-control-plane:$PATH
export PATH=`pwd`/build/artifacts/konvoy-dataplane:$PATH
export PATH=`pwd`/build/artifacts/konvoyctl:$PATH
```

### Install Control Plane

```
mkdir -p /tmp/getkonvoy.io/demo1/vm/memory/control-plane

cat >/tmp/getkonvoy.io/demo1/vm/memory/control-plane/config.yaml <<EOL
environment: universal
store:
  type: memory
xdsServer:
  grpcPort: 5678
  httpPort: 5679
  diagnosticsPort: 5680
apiServer:
  port: 5681
EOL

konvoy-control-plane run --config-file /tmp/getkonvoy.io/demo1/vm/memory/control-plane/config.yaml
```

```
KONVOY_GRPC_PORT=5678 \
KONVOY_HTTP_PORT=5679 \
KONVOY_XDS_SERVER_DIAGNOSTICS_PORT=5680 \
KONVOY_API_SERVER_PORT=5681 \
KONVOY_ENVIRONMENT=universal \
KONVOY_STORE_TYPE=memory \
\
konvoy-control-plane run
```

### Deploy Dataplane

```
KONVOY_CONTROL_PLANE_XDS_SERVER_ADDRESS=localhost \
KONVOY_CONTROL_PLANE_XDS_SERVER_PORT=5678 \
KONVOY_DATAPLANE_ID=gmail-01 \
KONVOY_DATAPLANE_SERVICE=gmail \
KONVOY_DATAPLANE_ADMIN_PORT=9901 \
\
konvoy-dataplane run
```

```
KONVOY_CONTROL_PLANE_XDS_SERVER_ADDRESS=localhost \
KONVOY_CONTROL_PLANE_XDS_SERVER_PORT=5678 \
KONVOY_DATAPLANE_ID=gcalendar-02 \
KONVOY_DATAPLANE_SERVICE=gcalendar \
KONVOY_DATAPLANE_ADMIN_PORT=9902 \
\
konvoy-dataplane run
```

### Get Status

```
mv ~/.konvoyctl/config{,.backup}
```

```
konvoyctl get dataplanes
```

```
konvoyctl config view
```

```
konvoyctl config control-planes add universal --name localhost --api-server-url http://localhost:5681
```

```
konvoyctl config view
```

```
konvoyctl get dataplanes
```

```
konvoyctl get dataplanes -oyaml
```

```
konvoyctl get dataplanes --mesh pilot
```

Notes:
- `NAMESPACE` column

## VM environment (Postgres)

### Install Control Plane

```bash
make start/postgres
```

```
mkdir -p /tmp/getkonvoy.io/demo1/vm/postgres/control-plane

cat >/tmp/getkonvoy.io/demo1/vm/postgres/control-plane/config.yaml <<EOL
environment: universal
store:
  type: postgres
  postgres:
    host: localhost
    port: 15432
    user: konvoy
    password: konvoy
    dbName: konvoy
    connectionTimeout: 10
xdsServer:
  grpcPort: 5678
  httpPort: 5679
  diagnosticsPort: 5680
apiServer:
  port: 5681
EOL

konvoy-control-plane run --config-file /tmp/getkonvoy.io/demo1/vm/postgres/control-plane/config.yaml
```

```
mkdir -p /tmp/getkonvoy.io/demo1/vm/postgres

KONVOY_GRPC_PORT= \
KONVOY_HTTP_PORT= \
KONVOY_XDS_SERVER_DIAGNOSTICS_PORT=5680 \
KONVOY_API_SERVER_PORT=5681 \
KONVOY_ENVIRONMENT=universal \
KONVOY_STORE_TYPE=postgres \
KONVOY_STORE_POSTGRES_HOST=localhost \
KONVOY_STORE_POSTGRES_PORT=15432 \
KONVOY_STORE_POSTGRES_USER=konvoy \
KONVOY_STORE_POSTGRES_PASSWORD=konvoy \
KONVOY_STORE_POSTGRES_DB_NAME=konvoy \
\
konvoy-control-plane run
```

### Deploy Dataplane

```
KONVOY_CONTROL_PLANE_XDS_SERVER_ADDRESS=localhost \
KONVOY_CONTROL_PLANE_XDS_SERVER_PORT=5678 \
KONVOY_DATAPLANE_ID=gmail-01 \
KONVOY_DATAPLANE_SERVICE=gmail \
KONVOY_DATAPLANE_ADMIN_PORT=9901 \
\
konvoy-dataplane run
```

```
KONVOY_CONTROL_PLANE_XDS_SERVER_ADDRESS=localhost \
KONVOY_CONTROL_PLANE_XDS_SERVER_PORT=5678 \
KONVOY_DATAPLANE_ID=gcalendar-02 \
KONVOY_DATAPLANE_SERVICE=gcalendar \
KONVOY_DATAPLANE_ADMIN_PORT=9901 \
\
konvoy-dataplane run
```

### Get Status

```
mv ~/.konvoyctl/config{,.backup}
```

```
konvoyctl get dataplanes
```

```
konvoyctl config view
```

```
konvoyctl config control-planes add universal --name localhost --api-server-url http://localhost:5681
```

```
konvoyctl config view
```

```
konvoyctl get dataplanes
```

```
konvoyctl get dataplanes -oyaml
```

```
konvoyctl get dataplanes --mesh pilot
```
