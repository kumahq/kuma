# Kubernetes Deployment Guide

## Setup Environment

### 1. Start a Kubernetes cluster with at least 4GB of memory. We've tested Kuma on Kubernetes v1.13.0 - v1.16.x, so use anything older than v1.13.0 with caution. In this demo, we'll be using v1.15.4. 

```
$ minikube start --cpus 2 --memory 4096 --kubernetes-version v1.15.4
😄  minikube v1.4.0 on Darwin 10.14.6
🔥  Creating virtualbox VM (CPUs=2, Memory=4096MB, Disk=20000MB) ...
🐳  Preparing Kubernetes v1.15.4 on Docker 18.09.9 ...
🚜  Pulling images ...
🚀  Launching Kubernetes ...
⌛  Waiting for: apiserver proxy etcd scheduler controller dns
🏄  Done! kubectl is now configured to use "minikube"
```

### 2. Navigate into the directory where all the kuma-demo YAML files are:

```
$ cd examples/kubernetes/kuma-demo/
```

### 3. Deploy Kuma's sample marketplace application

```
$ kubectl apply -f kuma-demo-aio.yaml
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

```
$ wget https://kong.bintray.com/kuma/kuma-0.3.0-rc2-darwin-amd64.tar.gz
--2019-10-13 05:53:46--  https://kong.bintray.com/kuma/kuma-0.3.0-rc2-darwin-amd64.tar.gz
Resolving kong.bintray.com (kong.bintray.com)... 52.88.33.18, 54.200.232.13
Connecting to kong.bintray.com (kong.bintray.com)|52.88.33.18|:443... connected.
HTTP request sent, awaiting response... 302
Location: https://akamai.bintray.com/69/694567d6d0d64f5eb5a5841aea3b4c3d60c8f2a6e6c3ff79cd5d580edf22e12b?__gda__=exp=1570917947~hmac=68f26ab23b95f97acebfc4b33a1bc1e88aeca46a44b1bc349af851019c941d0a&response-content-disposition=attachment%3Bfilename%3D%22kuma-0.3.0-rc2-darwin-amd64.tar.gz%22&response-content-type=application%2Fgzip&requestInfo=U2FsdGVkX1_SREBFG76q54ykX416x4BKSbGVrX5A-GfV55I-FdyX_0L9WI3EaLJdsXfRQ4V2pY3vP9viaRvtUxQEjLKVz_AEytCDaz5VW3oTvdhio0sq10KPgW3Z3hFN&response-X-Checksum-Sha1=01c56caae58a6d14a1ad24545ee0b25421c6d48e&response-X-Checksum-Sha2=694567d6d0d64f5eb5a5841aea3b4c3d60c8f2a6e6c3ff79cd5d580edf22e12b [following]
--2019-10-13 05:53:47--  https://akamai.bintray.com/69/694567d6d0d64f5eb5a5841aea3b4c3d60c8f2a6e6c3ff79cd5d580edf22e12b?__gda__=exp=1570917947~hmac=68f26ab23b95f97acebfc4b33a1bc1e88aeca46a44b1bc349af851019c941d0a&response-content-disposition=attachment%3Bfilename%3D%22kuma-0.3.0-rc2-darwin-amd64.tar.gz%22&response-content-type=application%2Fgzip&requestInfo=U2FsdGVkX1_SREBFG76q54ykX416x4BKSbGVrX5A-GfV55I-FdyX_0L9WI3EaLJdsXfRQ4V2pY3vP9viaRvtUxQEjLKVz_AEytCDaz5VW3oTvdhio0sq10KPgW3Z3hFN&response-X-Checksum-Sha1=01c56caae58a6d14a1ad24545ee0b25421c6d48e&response-X-Checksum-Sha2=694567d6d0d64f5eb5a5841aea3b4c3d60c8f2a6e6c3ff79cd5d580edf22e12b
Resolving akamai.bintray.com (akamai.bintray.com)... 104.93.1.149
Connecting to akamai.bintray.com (akamai.bintray.com)|104.93.1.149|:443... connected.
HTTP request sent, awaiting response... 200 OK
Length: 42892462 (41M) [application/gzip]
Saving to: ‘kuma-0.3.0-rc2-darwin-amd64.tar.gz’

kuma-0.3.0-rc2-darwin-amd64.tar.g 100%[===============================================>]  40.91M  2.61MB/s    in 20s

2019-10-13 05:54:08 (2.09 MB/s) - ‘kuma-0.3.0-rc2-darwin-amd64.tar.gz’ saved [42892462/42892462]
```

### 6. Unbundle the files to get the following components:

```
$ tar xvzf kuma-0.3.0-rc2-darwin-amd64.tar.gz
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
envoy   kuma-cp   kuma-dp   kuma-tcp-echo kumactl
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

```
$ kubectl apply -f kuma-demo-log.yaml
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

### 20. If we wanted to enable the Redis service again in the future, just change the traffic-permission back like this:
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