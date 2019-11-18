# Kubernetes Deployment Guide

## Setup Environment

### 1. Start a Kubernetes cluster with at least 4GB of memory. We've tested Kuma on Kubernetes v1.13.0 - v1.16.x, so use anything older than v1.13.0 with caution. In this demo, we'll be using v1.15.4. 

```
$ minikube start --cpus 2 --memory 4096 --kubernetes-version v1.15.4 -p kuma-demo
😄  [kuma-demo] minikube v1.5.2 on Darwin 10.15.1
✨  Automatically selected the 'hyperkit' driver (alternates: [virtualbox])
🔥  Creating hyperkit VM (CPUs=2, Memory=4096MB, Disk=20000MB) ...
🐳  Preparing Kubernetes v1.15.4 on Docker '18.09.9' ...
🚜  Pulling images ...
🚀  Launching Kubernetes ... 
⌛  Waiting for: apiserver
🏄  Done! kubectl is now configured to use "kuma-demo"
```

### 2. Navigate into the directory where all the kuma-demo YAML files are:

```
$ cd examples/kubernetes/kuma-demo/
```

### 3. Deploy Kuma's sample marketplace application in minikube
You can deploy the sample marketplace application via the [bit.ly](http://bit.ly/kuma1116) link as shown below or via the `kuma-demo-aio.yaml` file in this directory.
```
$ kubectl apply -f http://bit.ly/kuma1116
namespace/kuma-demo created
serviceaccount/elasticsearch created
service/elasticsearch created
replicationcontroller/es created
deployment.apps/redis-master created
service/redis created
service/backend created
deployment.apps/kuma-demo-backend-v0 created
deployment.apps/kuma-demo-backend-v1 created
deployment.apps/kuma-demo-backend-v2 created
configmap/demo-app-config created
service/frontend created
deployment.apps/kuma-demo-app created
```

This will deploy our demo marketplace application split across multiple pods:
1. The first pod is an Elasticsearch service that stores all the items in our marketplace
2. The second pod is a Redis service that stores reviews for each item
3. The third pod is a Node application that represents a backend
4. The remaining pods represent multiple versions of our Node/Vue application that allows you to visually query the Elastic and Redis endpoints

Check the pods are up and running by checking the `kuma-demo` namespace

```
$ kubectl get pods -n kuma-demo
NAME                                    READY   STATUS    RESTARTS   AGE
es-v6g88                                1/1     Running   0          32s
kuma-demo-app-7bb5d85c8c-8kl2z          2/2     Running   0          30s
kuma-demo-backend-v0-7dcb8dc8fd-rq798   1/1     Running   0          31s
redis-master-5b5978b77f-pmhnz           1/1     Running   0          32s
```

In the following steps, we will be using the pod name of the `kuma-demo-app-*************` pod. Please replace any `${KUMA_DEMO_APP_POD_NAME}` variables with your pod name.

### 4. Port-forward the sample application to access the front-end UI at http://localhost:8080

<pre><code>$ kubectl port-forward <b>${KUMA_DEMO_APP_POD_NAME}</b> -n kuma-demo 8080:80
Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80
</code></pre>

Now you can access the marketplace application through your web browser at http://localhost:8080.

The items on the front page are pulled from the Elasticsearch service. While the reviews for each item sit within the Redis service. You can query for individual items and look at their reviews.


### 5. Download the latest version of Kuma
The following command will download the Mac compatible version of Kuma. To find the correct version for your operating system, please check out [Kuma's official installation page](https://kuma.io/install).

```
$ wget https://kong.bintray.com/kuma/kuma-0.3.0-darwin-amd64.tar.gz
--2019-11-18 07:46:55--  https://kong.bintray.com/kuma/kuma-0.3.0-darwin-amd64.tar.gz
Resolving kong.bintray.com (kong.bintray.com)... 52.36.38.54, 54.149.74.157
Connecting to kong.bintray.com (kong.bintray.com)|52.36.38.54|:443... connected.
HTTP request sent, awaiting response... 302 
Location: https://akamai.bintray.com/3a/3afc187b8e3daa912648fcbe16f0aa9c2eb90b4b0df4f0a5d47d74ae426371b1?__gda__=exp=1574092735~hmac=69b07d97c61a32e3f09e9072f740b3472f86bf663a84f3a808142bcf7541da72&response-content-disposition=attachment%3Bfilename%3D%22kuma-0.3.0-darwin-amd64.tar.gz%22&response-content-type=application%2Fgzip&requestInfo=U2FsdGVkX1-wOkJsEzHavzEbyKAyRNRIaEgd96BSSg_Fa7UU3OhI_p-1NSKjEepZrhEAl7IRPiU5LqI6KDH4rX7QxYihgWtBtGY2rlIY51TCbTYnklZZvXx4xQo-mDE2&response-X-Checksum-Sha1=6df196169311c66a544eccfdd73931b6f3b83593&response-X-Checksum-Sha2=3afc187b8e3daa912648fcbe16f0aa9c2eb90b4b0df4f0a5d47d74ae426371b1 [following]
--2019-11-18 07:46:55--  https://akamai.bintray.com/3a/3afc187b8e3daa912648fcbe16f0aa9c2eb90b4b0df4f0a5d47d74ae426371b1?__gda__=exp=1574092735~hmac=69b07d97c61a32e3f09e9072f740b3472f86bf663a84f3a808142bcf7541da72&response-content-disposition=attachment%3Bfilename%3D%22kuma-0.3.0-darwin-amd64.tar.gz%22&response-content-type=application%2Fgzip&requestInfo=U2FsdGVkX1-wOkJsEzHavzEbyKAyRNRIaEgd96BSSg_Fa7UU3OhI_p-1NSKjEepZrhEAl7IRPiU5LqI6KDH4rX7QxYihgWtBtGY2rlIY51TCbTYnklZZvXx4xQo-mDE2&response-X-Checksum-Sha1=6df196169311c66a544eccfdd73931b6f3b83593&response-X-Checksum-Sha2=3afc187b8e3daa912648fcbe16f0aa9c2eb90b4b0df4f0a5d47d74ae426371b1
Resolving akamai.bintray.com (akamai.bintray.com)... 23.35.181.234
Connecting to akamai.bintray.com (akamai.bintray.com)|23.35.181.234|:443... connected.
HTTP request sent, awaiting response... 200 OK
Length: 38017379 (36M) [application/gzip]
Saving to: ‘kuma-0.3.0-darwin-amd64.tar.gz’

kuma-0.3.0-darwin-amd64.tar.gz            100%[====================================================================================>]  36.26M  5.02MB/s    in 8.0s    

2019-11-18 07:47:04 (4.52 MB/s) - ‘kuma-0.3.0-darwin-amd64.tar.gz’ saved [38017379/38017379]
```

### 6. Unbundle the files to get the following components:

```
$ tar xvzf kuma-0.3.0-darwin-amd64.tar.gz
x ./
x ./conf/
x ./conf/kuma-cp.conf
x ./bin/
x ./bin/kuma-dp
x ./bin/envoy
x ./bin/kuma-tcp-echo
x ./bin/kumactl
x ./bin/kuma-cp
x ./README
x ./LICENSE
```

### 7. Go into the ./bin directory where the kuma components will be:

```
$ cd bin && ls
envoy   kuma-cp   kuma-dp   kuma-tcp-echo   kumactl
```

### 8. Install the control plane using `kumactl`

```
$ ./kumactl install control-plane | kubectl apply -f -
namespace/kuma-system created
secret/kuma-admission-server-tls-cert created
secret/kuma-injector-tls-cert created
secret/kuma-sds-tls-cert created
configmap/kuma-control-plane-config created
configmap/kuma-injector-config created
serviceaccount/kuma-control-plane created
customresourcedefinition.apiextensions.k8s.io/dataplaneinsights.kuma.io configured
customresourcedefinition.apiextensions.k8s.io/dataplanes.kuma.io configured
customresourcedefinition.apiextensions.k8s.io/meshes.kuma.io configured
customresourcedefinition.apiextensions.k8s.io/proxytemplates.kuma.io configured
customresourcedefinition.apiextensions.k8s.io/trafficlogs.kuma.io configured
customresourcedefinition.apiextensions.k8s.io/trafficpermissions.kuma.io configured
customresourcedefinition.apiextensions.k8s.io/trafficroutes.kuma.io configured
clusterrole.rbac.authorization.k8s.io/kuma:control-plane unchanged
clusterrolebinding.rbac.authorization.k8s.io/kuma:control-plane unchanged
role.rbac.authorization.k8s.io/kuma:control-plane created
rolebinding.rbac.authorization.k8s.io/kuma:control-plane created
service/kuma-injector created
service/kuma-control-plane created
deployment.apps/kuma-control-plane created
deployment.apps/kuma-injector created
mutatingwebhookconfiguration.admissionregistration.k8s.io/kuma-admission-mutating-webhook-configuration configured
mutatingwebhookconfiguration.admissionregistration.k8s.io/kuma-injector-webhook-configuration configured
validatingwebhookconfiguration.admissionregistration.k8s.io/kuma-validating-webhook-configuration configured
```

You can check the pods are up and running by checking the `kuma-system` namespace

```
$ kubectl get pods -n kuma-system
NAME                                  READY   STATUS    RESTARTS   AGE
kuma-control-plane-7bcc56c869-lzw9t   1/1     Running   0          70s
kuma-injector-9c96cddc8-745r7         1/1     Running   0          70s
```

In the following steps, we will be using the pod name of the `kuma-control-plane-*************` pod. Please replace any `${KUMA_CP_POD_NAME}` with your pod name.

### 9. Delete the existing kuma-demo pods so they restart:

```
$ kubectl delete pods --all -n kuma-demo
pod "es-v6g88" deleted
pod "kuma-demo-app-7bb5d85c8c-8kl2z" deleted
pod "kuma-demo-backend-v0-7dcb8dc8fd-rq798" deleted
pod "redis-master-5b5978b77f-pmhnz" deleted
```

And check the pods are up and running again with an additional container. The additional container is the Envoy sidecar proxy that Kuma is injecting into each pod.

```
$ kubectl get pods -n kuma-demo
NAME                                    READY   STATUS    RESTARTS   AGE
es-5snv2                                2/2     Running   0          37s
kuma-demo-app-7bb5d85c8c-5sqxl          3/3     Running   0          37s
kuma-demo-backend-v0-7dcb8dc8fd-7ttjm   2/2     Running   0          37s
redis-master-5b5978b77f-hwjvd           2/2     Running   0          37s
```

### 10. Port-forward the sample application again to access the front-end UI at http://localhost:8080

<pre><code>$ kubectl port-forward <b>${KUMA_DEMO_APP_POD_NAME}</b> -n kuma-demo 8080:80
Forwarding from 127.0.0.1:8080 -> 80
Forwarding from [::1]:8080 -> 80
</code></pre>

Now you can access the marketplace application through your web browser at http://localhost:8080 with Envoy handling all the traffic between the services. Happy shopping!

### 11.  Now we will port forward the kuma-control-plane so we can access it with `kumactl`

<pre><code>$ kubectl -n kuma-system port-forward <b>${KUMA_CP_POD_NAME}</b> 5681
Forwarding from 127.0.0.1:5681 -> 5681
Forwarding from [::1]:5681 -> 5681
</code></pre>

Please refer to step 7 to copy the correct `${KUMA_CP_POD_NAME}`.

### 12.  Now configure `kumactl` to point towards the control plane address

```
$ ./kumactl config control-planes add --name=minikube --address=http://localhost:5681
added Control Plane "minikube"
switched active Control Plane to "minikube"
```

### 13. You can use `kumactl` to look at the dataplanes in the mesh. You should see three dataplanes that correlates with our pods in Kubernetes:

```
$ ./kumactl inspect dataplanes
MESH      NAME                                    TAGS                                                                                               STATUS   LAST CONNECTED AGO   LAST UPDATED AGO   TOTAL UPDATES   TOTAL ERRORS
default   redis-master-5b5978b77f-hwjvd           app=redis pod-template-hash=5b5978b77f role=master service=redis.kuma-demo.svc:6379 tier=backend   Online   2m7s                 2m3s               8               0
default   es-5snv2                                component=elasticsearch service=elasticsearch.kuma-demo.svc:80                                     Online   1m49s                1m48s              3               0
default   kuma-demo-app-7bb5d85c8c-5sqxl          app=kuma-demo-frontend pod-template-hash=7bb5d85c8c service=frontend.kuma-demo.svc:80              Online   1m49s                1m48s              3               0
default   kuma-demo-backend-v0-7dcb8dc8fd-7ttjm   app=kuma-demo-backend pod-template-hash=7dcb8dc8fd service=backend.kuma-demo.svc:3001 version=v0   Online   1m47s                1m46s              3               0
```

### 14. You can also use `kumactl` to look at the mesh. As shown below, our default mesh does not have mTLS enabled.

```
$ ./kumactl get meshes
NAME      mTLS
default   off
```

### 15.  Let's enable mTLS.

```
$ cat <<EOF | kubectl apply -f - 
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
  namespace: kuma-system
spec:
  mtls:
    ca:
      builtin: {}
    enabled: true
EOF
```

Using `kumactl`, inspect the mesh again to see if mTLS is enabled:

```
$ ./kumactl get meshes
NAME      mTLS
default   on
```

### 16.  Now let's enable traffic-permission for all services so our application will work like it use to:

```
$ cat <<EOF | kubectl apply -f - 
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  namespace: kuma-demo
  name: everything
spec:
  sources:
  - match:
      service: '*'
  destinations:
  - match:
      service: '*'
EOF
```

Using `kumactl`, you can check the traffic permissions like this:
```
$ ./kumactl get traffic-permissions
MESH      NAME
default   everything
```

Now that we have traffic permission that allows any source to talk to any destination, our application should work like it use to. 

### 17. Deploy the logstash service.
You can deploy the logtash service via the [bit.ly](http://bit.ly/kumalog) link as shown below or via the `kuma-demo-log.yaml` file in this directory.
```
$ kubectl apply -f http://bit.ly/kumalog
namespace/logging created
service/logstash created
configmap/logstash-config created
deployment.apps/logstash created
```

### 18. Let's add logging for traffic between all services and send them to logstash: 
```
$ cat <<EOF | kubectl apply -f - 
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
  namespace: kuma-system
spec:
  mtls:
    ca:
      builtin: {}
    enabled: true
  logging:
    backends:
    - name: logstash
      format: |
        {
            "destination": "%KUMA_DESTINATION_SERVICE%",
            "destinationAddress": "%UPSTREAM_HOST%",
            "source": "%KUMA_SOURCE_SERVICE%",
            "sourceAddress": "%KUMA_SOURCE_ADDRESS%",
            "bytesReceived": "%BYTES_RECEIVED%",
            "bytesSent": "%BYTES_SENT%"
        }
      tcp:
        address: logstash.logging:5000
---
apiVersion: kuma.io/v1alpha1
kind: TrafficLog
mesh: default
metadata:
  namespace: kuma-demo
  name: everything
spec:
  sources:
  - match:
      service: '*'
  destinations:
  - match:
      service: '*'
  conf:
    backend: logstash
EOF
```
Logs will be sent to https://kumademo.loggly.com/

### 19. Now let's take down our Redis service because someone is spamming fake reviews. We can easily accomplish that by changing our traffic-permissions:

```
$ kubectl delete trafficpermission -n kuma-demo --all
```

```
$ cat <<EOF | kubectl apply -f - 
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  namespace: kuma-demo
  name: frontend-to-backend
spec:
  sources:
  - match:
      service: frontend.kuma-demo.svc:80
  destinations:
  - match:
      service: backend.kuma-demo.svc:3001
---
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  namespace: kuma-demo
  name: backend-to-elasticsearch
spec:
  sources:
  - match:
      service: backend.kuma-demo.svc:3001
  destinations:
  - match:
      service: elasticsearch.kuma-demo.svc:80
EOF
```

This traffic-permission will only allow traffic from the kuma-demo-api service to the Elasticsearch service. Now try to access the reviews on each item. They will not load because of the traffic-permissions you described in the the policy above.

### 19.5. If we wanted to enable the Redis service again in the future, just change the traffic-permission back like this:
```
$ cat <<EOF | kubectl apply -f - 
apiVersion: kuma.io/v1alpha1
kind: TrafficPermission
mesh: default
metadata:
  namespace: kuma-demo
  name: backend-to-redis
spec:
  sources:
  - match:
      service: backend.kuma-demo.svc:3001
  destinations:
  - match:
      service: redis.kuma-demo.svc:6379
EOF
```

### 20. Let's explore adding traffic routing to our service mesh. But before we do, we need to scale up the v1 and v2 deployment of our sample application:
```
$ kubectl scale deployment kuma-demo-backend-v1 -n kuma-demo --replicas=1
deployment.extensions/kuma-demo-backend-v1 scaled
```
```
$ kubectl scale deployment kuma-demo-backend-v2 -n kuma-demo --replicas=1
deployment.extensions/kuma-demo-backend-v2 scaled
```
and check all the pods are running like this:
```
kubectl get pods -n kuma-demo
NAME                                    READY   STATUS    RESTARTS   AGE
es-v6t5t                                2/2     Running   0          5h56m
kuma-demo-app-85bb496b68-ccv2f          3/3     Running   0          5h56m
kuma-demo-backend-v0-bd9984f8f-d9tl7    2/2     Running   0          5h56m
kuma-demo-backend-v1-554c4d85c4-trt67   2/2     Running   0          16m
kuma-demo-backend-v2-6b6bc8f585-4qtjw   2/2     Running   0          16m
redis-master-b688d4f4-jjvvt             2/2     Running   0          5h56m
```
`v0` is set to have 0 sales, while `v1` has 1 special offer item, and lastly `v2` has 2 special offer. Here is a visual representation of how it looks:
```           
                        ----> backend-v0  :  service=backend, version=v0, env=prod
                      /
(browser) -> frontend   ----> backend-v1  :  service=backend, version=v1, env=intg
                      \
                        ----> backend-v2  :  service=backend, version=v2, env=dev
```

### 21. Define a handy alias that will can help show the power of Kuma's traffic routing:
```
$ alias benchmark='echo "NUM_REQ NUM_SPECIAL_OFFERS"; kubectl -n kuma-demo exec $( kubectl -n kuma-demo get pods -l app=kuma-demo-frontend -o=jsonpath="{.items[0].metadata.name}" ) -c kuma-fe -- sh -c '"'"'for i in `seq 1 100`; do curl -s http://backend:3001/items?q | jq -c ".[] | select(._source.specialOffer == true)" | wc -l ; done | sort | uniq -c | sort -k2n'"'"''
```
This alias will help send 100 request from `front-end` to `backend` and count the number of special offers in the response. Then it will group the request by the number of special offers. Here is an example of the output before we start configuring our traffic-routing.
```
$ benchmark
NUM_REQ    NUM_SPECIAL_OFFERS
34         0
33         1
33         2
```
The traffic is equally distributed because have not set any traffic-routing. Let's change that!

### 22. Traffic routing to limit amount of special offers on Kuma marketplace:
To avoid going broke, let's limit the amount of special offers that appear on our marketplace. To do so, apply this TrafficRoute policy:

```bash
$ cat <<EOF | kubectl apply -f -
apiVersion: kuma.io/v1alpha1
kind: TrafficRoute
metadata:
  name: frontend-to-backend
  namespace: kuma-demo
mesh: default
spec:
  sources:
  - match:
      service: frontend.kuma-demo.svc:80
  destinations:
  - match:
      service: backend.kuma-demo.svc:3001
  conf:
  # it is NOT a percentage. just a positive weight
  - weight: 80
    destination:
      service: backend.kuma-demo.svc:3001
      version: v0
  # we're NOT checking if total of all weights is 100  
  - weight: 20
    destination:
      service: backend.kuma-demo.svc:3001
      version: v1
  # 0 means no traffic will be sent there
  - weight: 0
    destination:
      service: backend.kuma-demo.svc:3001
      version: v2
EOF
```
Run our benchmark to make sure no one is getting two special offers on the webpage:
```bash
$ benchmark
NUM_REQ    NUM_SPECIAL_OFFERS
84         0
16         1
```
And clean the traffic route before we try more things:
```bash
kubectl delete trafficroute -n kuma-demo --all
```

### 23. Resolving Collisions - Identical Selectors

Let's dive deeper into certain Kuma's traffic routing behaviors. If 2 routes have identical selectors but different destinations, how would Kuma handle it? Let's add start by creating this situation with the following traffic route policies:
```bash
$ cat <<EOF | kubectl apply -f -
apiVersion: kuma.io/v1alpha1
kind: TrafficRoute
metadata:
  name: route-2                                         # notice the choice of a name
  namespace: kuma-demo
mesh: default
spec:
  sources:
  - match:
      service: frontend.kuma-demo.svc:80      # <<< same selector
  destinations:
  - match:
      service: backend.kuma-demo.svc:3001
  conf:
  - weight: 100
    destination:
      service: backend.kuma-demo.svc:3001
      version: v1       # <<< subset 1
EOF
```
and 

```bash
cat <<EOF | kubectl apply -f -
apiVersion: kuma.io/v1alpha1
kind: TrafficRoute
metadata:
  name: route-1
  namespace: kuma-demo
mesh: default
spec:
  sources:
  - match:
      service: frontend.kuma-demo.svc:80      # <<< same selector
  destinations:
  - match:
      service: backend.kuma-demo.svc:3001
  conf:
  - weight: 100
    destination:
      service: backend.kuma-demo.svc:3001
      version: v2       # <<< subset 2
EOF
```

With two routes set up with identical selectors, let's try our `benchmark` alias again.
```bash
$ benchmark
NUM_REQ    NUM_SPECIAL_OFFERS
100        2
```
Due to ordering by name, the `TrafficRoute` with the name of `route-1` takes priority and all the traffic is routed to our `v2` application with 2 special offers.
Let's clean the traffic route before we try more things:
```bash
kubectl delete trafficroute -n kuma-demo --all
```

### 24. Resolving Collisions - Extra Tags

In the scenario where one route has more tag, what would happen? Apply these two routes and find out:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: kuma.io/v1alpha1
kind: TrafficRoute
metadata:
  name: route-1
  namespace: kuma-demo
mesh: default
spec:
  sources:
  - match:
      service: frontend.kuma-demo.svc:80      # <<< match by 1 tag
  destinations:
  - match:
      service: backend.kuma-demo.svc:3001
  conf:
  - weight: 100
    destination:
      service: backend.kuma-demo.svc:3001
      version: v0 # <<< subset 1
EOF
```
and
```bash
cat <<EOF | kubectl apply -f -
apiVersion: kuma.io/v1alpha1
kind: TrafficRoute
metadata:
  name: route-2
  namespace: kuma-demo
mesh: default
spec:
  sources:
  - match:
      service: frontend.kuma-demo.svc:80      # <<< match by 2 tags
      env: prod                               
  destinations:
  - match:
      service: backend.kuma-demo.svc:3001
  conf:
  - weight: 100
    destination:
      service: backend.kuma-demo.svc:3001
      version: v2       # <<< subset 2
EOF
```
Now run the `benchmark` alias again:
```bash
$ benchmark
NUM_REQ    NUM_SPECIAL_OFFERS
100        2
```
Once again, our `route-2` traffic routing policy triumphs. In the scenario where one route has more tags, Kuma will prioritize that route.
