# K8s

## Different meshes

### Components
 
* `konvoy-control-plane`
* `service `backend`, mesh `default`
* `service `web`, mesh `default`
* `service `mobile`, mesh `pilot`

### Installing

TODO: control-plane

```
kubectl apply -f demos/demo2/k8s/meshes/backend.yaml
kubectl wait --timeout=60s --for=condition=Available -n demo2-meshes-backend deployment/backend
kubectl wait --timeout=60s --for=condition=Ready -n demo2-meshes-backend pods -l app=backend
kubectl get pods -n demo2-meshes-backend

kubectl port-forward -n konvoy-system $(kubectl get pods -n konvoy-system -l app=konvoy-control-plane -o=jsonpath='{.items[0].metadata.name}') 25681:5681 &
konvoyctl config control-planes add --name demo2-k8s-meshes --api-server-url http://localhost:25681
konvoyctl config view
konvoyctl get dataplanes

kubectl get dataplanes -n demo2-meshes-backend
kubectl get dataplanes -n demo2-meshes-backend -oyaml
konvoyctl get dataplanes

#

kubectl apply -f demos/demo2/k8s/meshes/web.yaml
kubectl wait --timeout=60s --for=condition=Available -n demo2-meshes-web deployment/web
kubectl wait --timeout=60s --for=condition=Ready -n demo2-meshes-web pods -l app=web
kubectl get pods -n demo2-meshes-web

konvoyctl get dataplanes
kubectl get dataplanes -n demo2-meshes-web

#

kubectl apply -f demos/demo2/k8s/meshes/mobile.yaml
kubectl wait --timeout=60s --for=condition=Available -n demo2-meshes-mobile deployment/mobile
kubectl wait --timeout=60s --for=condition=Ready -n demo2-meshes-mobile pods -l app=mobile
kubectl get pods -n demo2-meshes-mobile

konvoyctl get dataplanes
konvoyctl get dataplanes --mesh=pilot
kubectl get dataplanes -n demo2-meshes-mobile

```

### Missing

- delete `DataplaneInsights` on Pod/Dataplane removal

# Universal

## Different meshes

### Components
 
* `konvoy-control-plane`
* `konvoy-dataplane` for service  `backend`, mesh `default`
* `konvoy-dataplane` for service  `web`, mesh `default`
* `konvoy-dataplane` for service  `mobile`, mesh `pilot`

### Deploying

Start Control Plane (in-memory) and bring up Dataplanes (Envoys):
```
docker-compose  -f demos/demo2/universal/docker-compose.yaml up
```

Showcase that there are no `Mesh` definitions yet:
```
curl -s localhost:15681/meshes | jq
```

Showcase that there are no `Dataplane` definitions yet:
```
curl -s localhost:15681/meshes/default/dataplanes | jq
curl -s localhost:15681/meshes/pilot/dataplanes | jq
```

Showcase that Envoy nodes are already connected to the Control Plane (although Control Plane cannot generate configuration yet):
```
curl -s localhost:15681/meshes/default/dataplane-insights | jq
curl -s localhost:15681/meshes/pilot/dataplane-insights | jq
```

Prelpare `konvoyctl`
```
make build/konvoyctl
export PATH=`pwd`/build/artifacts/konvoyctl:$PATH

rm $HOME/.konvoyctl/config
konvoyctl config control-planes add --name demo2-universal-meshes --api-server-url http://localhost:15681
konvoyctl config view
```

Showcase creation of `Meshes`: `default` and `pilot`
```
konvoyctl get meshes
konvoyctl apply -f demos/demo2/universal/meshes/default/mesh.yaml
konvoyctl get meshes
konvoyctl apply -f demos/demo2/universal/meshes/pilot/mesh.yaml
konvoyctl get meshes
```

Deploy `Dataplane` definitions:
```
# Notice 'LAST UPDATED AGO' is 'never'
konvoyctl get dataplanes --mesh default

konvoyctl apply -f demos/demo2/universal/meshes/default/dataplanes/backend-01.dataplane.yaml
curl -s localhost:15681/meshes/default/dataplanes | jq
# Notice 'LAST UPDATED AGO'
konvoyctl get dataplanes --mesh default

konvoyctl apply -f demos/demo2/universal/meshes/default/dataplanes/backend-02.dataplane.yaml
konvoyctl get dataplanes --mesh default

konvoyctl apply -f demos/demo2/universal/meshes/default/dataplanes/web-01.dataplane.yaml
konvoyctl get dataplanes --mesh default

konvoyctl apply -f demos/demo2/universal/meshes/pilot/dataplanes/mobile-01.dataplane.yaml
konvoyctl get dataplanes --mesh pilot
```

Check Envoy configurations:
```
docker-compose -f demos/demo2/universal/docker-compose.yaml exec backend-01 wget -qO- localhost:9901/config_dump | jq -c . | jq .
docker-compose -f demos/demo2/universal/docker-compose.yaml exec mobile-01 wget -qO- localhost:9901/config_dump | jq -c . | jq .
```

### Showcase ProxyTemplate

List `ProxyTemaplate`s
```
konvoyctl get proxytemplates
konvoyctl get proxytemplates --mesh=pilot
```

Override default `ProxyTemplate` within `pilot` mesh
```
konvoyctl get dataplanes --mesh default
konvoyctl apply -f demos/demo2/universal/meshes/pilot/proxytemplates/empty.yaml
konvoyctl get proxytemplates --mesh=pilot
konvoyctl get proxytemplates --mesh=pilot -oyaml

konvoyctl get dataplanes --mesh pilot
curl -XDELETE http://localhost:15681/meshes/pilot/dataplanes/mobile-01

konvoyctl get dataplanes --mesh pilot
curl http://localhost:15681/meshes/pilot/dataplanes/

docker-compose -f demos/demo2/universal/docker-compose.yaml exec mobile-01 wget -qO- localhost:9901/config_dump | jq -c . | jq .

konvoyctl get dataplanes --mesh pilot
konvoyctl apply -f demos/demo2/universal/meshes/pilot/dataplanes/mobile-01.dataplane.yaml
konvoyctl get dataplanes --mesh pilot

docker-compose -f demos/demo2/universal/docker-compose.yaml exec mobile-01 wget -qO- localhost:9901/config_dump | jq -c . | jq .

curl -XDELETE http://localhost:15681/meshes/default/proxytemplates/empty
```

Trigger regeneration of Envoy config:
```

```

### Missing

- no `default` mesh
- syntax `mesh` and `name`
```
type: Mesh
mesh: default
name: default
```
- no selectore in `ProxyTemplate` means that iut applies to everytging (k8s conventions)
- `konvoyctl delete`
- there is no validation of `Mesh`, `ProxyTemplate`, etc during create/update
- validation if such a Mesh already exists at Dataplane create/update time
- validation if such a Mesh already exists at DataplaneInsight create/update time
