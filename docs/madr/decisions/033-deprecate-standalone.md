# Deprecate standalone mode

* Status: accepted

## Context and Problem Statement

We want to simplify our deployments model to only have Zone CP and Global CP.
The goal is not to get rid of standalone altogether, but to change the naming.

## Decision Drivers

* Simplify UX for (new) users. Simpler docs. Fewer things to wrap your head around.
* Provide better migration path from standalone to multizone, as standalone won't exist.

## Considered Options

* Deprecate standalone

## Decision Outcome

Chosen option: "Deprecate standalone".

## Pros and Cons of the Options

### Deprecate standalone

If you want to run the "standalone" you can execute the following command
```
KUMA_MODE=zone kuma-cp run 
```
We will change the default `KUMA_MODE` value to `zone`. So in the end, you would run this the same way as before which is
```
kuma-cp run
```
HELM and kumactl install adds `KUMA_MODE=standalone` by default, we can also change the default to `zone`.

The current way of running Zone CP will not change.

The current difference of running standalone and zone is providing zone name and KDS address.

Current differences between standalone and zone CP:
* Cannot apply policies on Zone CP directly.
* GUI is disabled on Zone CP
* Cluster ID is synced from Global on Zone CP. Cluster ID is used for analytics
* Zone CP connects to Global CP to pull policies
* Default resources (Mesh, policies etc.) are disabled on Zone CP, because those are pulled from Global.
* Standalone does not finalize (gc/finalizer) zone ingresses
* Mesh/Service Insights are not computed on Zone CP
* GatewayAPI is not supported on Zone CP
* On Zone CP, Zone Token public key is synced from Global.

With deprecating standalone we still need to differentiate between zone running on its own and zone connected to the Global CP.
To do this, instead of relying on `KUMA_MODE` `standalone` or `zone`, we will check if KDS address was provided.

Existing standalone and future Zone CP without KDS connection will work the same way.

Changes between existing Zone CP and upcoming Zone CP with KDS connection:
* You can access GUI

#### Visibility changes

Currently, Zone tab is missing in "standalone" GUI.
We create Zone object and ZoneInsights with statistic of KDS only on Global CP. Those objects are not synced to Zone CP.
We want to also create Zone object for Zone CP and create ZoneInsights with KDS statistics from Zone perspective.
Then we can show on this screen whether the Zone is connected to Global or not.

#### Transition

Transition is described with the assumption that MADR #032 (applying policies on Zone CP) is implemented. 

##### Federated Zone CP to non-federated Zone CP

I tested it and it seems to just work. Nothing extraordinary happens there.

##### Non-federated Zone CP to federated Zone CP 

Let's assume we first start using Kuma with Zone CP without KDS connection.

Assuming that applying policies on Zone CP and syncing them on Global (see 032 MADR) is done, we only need to sync the following resources:
* Secrets
* Global Secrets
* Meshes

When we switch the zone to connect to the Global CP. The following will happen:
* Policies are synced from Global CP. Policies applied on Zone will be synced to Global CP.
* Secrets that are synced from Global CP are replaced on Zone CP
* You can no longer issue zone tokens on Zone CP.

**Zone on Kubernetes. Global on Kubernetes**

I managed to successfully federate Zone CP with the following commands.
```sh
# change context to zone
kubectl get mesh -oyaml > /tmp/meshes.yaml
yq eval 'del(.items[].metadata.resourceVersion, .items[].metadata.ownerReferences, .items[].metadata.uid, .items[].metadata.annotations, .items[].metadata.generation)' /tmp/meshes.yaml > /tmp/meshes-clean.yaml
kubectl get secrets -n kuma-system --field-selector='type=system.kuma.io/global-secret' -oyaml > /tmp/dump-gsecrets.yaml
yq eval 'del(.items[].metadata.resourceVersion, .items[].metadata.ownerReferences, .items[].metadata.uid, .items[].metadata.annotations, .items[].metadata.generation)' /tmp/dump-gsecrets.yaml > /tmp/dump-gsecrets-clean.yaml
kubectl get secrets -n kuma-system --field-selector='type=system.kuma.io/secret' -oyaml > /tmp/dump-secrets.yaml
yq eval 'del(.items[].metadata.resourceVersion, .items[].metadata.ownerReferences, .items[].metadata.uid, .items[].metadata.annotations, .items[].metadata.generation)' /tmp/dump-secrets.yaml > /tmp/dump-secrets-clean.yaml

# change context to global
kubectl delete meshes --all && sleep 5 # or deploy with controlPlane.defaults.skipMeshCreation=true
kubectl apply -f /tmp/dump-gsecrets-clean.yaml
kubectl apply -f /tmp/dump-secrets-clean.yaml
kubectl apply -f /tmp/meshes-clean.yaml
```

We could encapsulate this using `kumactl` with the following command
```sh
kumactl export \
  --source-kube-context zone \ # read config from KUBECONFIG just like install control-plane. Instead of using Kuma API, it will use kubectl to extract policies
  --profile federation \ # "all" to dump everything
  --format k8s > policies.yaml
  
kubectl --context=global apply -f policies.yaml
```

The advantage of two-step command instead of something like `kumactl federate` is that we can see what will be applied on the cluster.
Because we use regular `kubectl apply -f`, a user will understand that we simply copy a bunch of resources.

The advantage of `profile` is that `federation` is a minimal profile for transition to happen.
However, a user can pick `all` to export all policies so after the federation is done, other zone that connects to global will pick up all policies.

**Zone on Kube/Universal. Global on Universal**
```sh
# export policies
export ZONE_ADMIN_TOKEN=$(kubectl get secrets -n kuma-system admin-user-token -ojson | jq -r .data.value | base64 -d)
kumactl config control-planes add \
  --address http://localhost:25681 \
  --headers "authorization=Bearer $ZONE_ADMIN_TOKEN" \
  --name "zone" \
  --overwrite
kumactl get global-secrets -oyaml | yq '.items[] | split_doc' > /tmp/global_secrets.yaml
kumactl get secrets -oyaml | yq '.items[] | split_doc' > /tmp/secrets.yaml
kumactl get meshes -oyaml | yq '.items[] | split_doc' > /tmp/meshes.yaml

# apply policies
export GLOBAL_ADMIN_TOKEN=$(docker exec kuma-global-cp wget -O- -q localhost:5681/global-secrets/admin-user-token | jq -r '.data' | base64 -d)
kumactl config control-planes add \
  --address http://localhost:15681 \
  --name "global" \
  --headers "authorization=Bearer $GLOBAL_ADMIN_TOKEN" \
  --overwrite
kumactl apply -f /tmp/global_secrets.yaml
kumactl apply -f /tmp/secrets.yaml
kumactl apply -f /tmp/meshes.yaml
```
Note: I had to run all commands after `# apply policies` twice because we replace user token signing key, so auth is broken.
To avoid broken auth, `kumactl export` could export auth policy as the last resource on the list.

We could encapsulate this using `kumactl` with the following command
```
kumactl export \
  --context zone \ # implement https://github.com/kumahq/kuma/issues/7309
  --profile federation \
  --format universal > policies.yaml

kumactl --context=global apply -f policies.yaml
```

**Zone on Universal. Global on Kubernetes**
It's possible to generate K8S responses on Universal, but it's broken for secrets.
When we fix it, we can extract K8S jsons using API and then apply them using kubectl on global.
We also need to support this for lists, currently we only support the format for a single resource.

We can encapsulate this with the following command

```
kumactl export \
  --context zone \ # implement https://github.com/kumahq/kuma/issues/7309
  --profile federation \
  --format k8s > policies.yaml

kubectl --context=global apply -f policies.yaml
```

##### Multiple non-federated Zone CP to federated Zone CP

We first need to make sure that all separate non-federated Zone CP have the same secrets. To be more specific:
* Same CAs for the same meshes
* Same signing keys for all keys

Rotation guides should be described in docs already.

Then we need to export YAMLs in the expected format on global and apply all files on Global CP.

##### Default Mesh policies

Default mesh policies are a bit problematic during the transition.
Let say we created Mesh on Zone CP and deleted default policies. If we simply reapply Mesh on Global CP, it will create default policies which then will be synced to Zone CP.
In this case we should recommend for users to add `skipCreatingInitialPolicies` to meshes before starting the transition.

The other case is that a user modified default mesh policies. In this case if we reapply Mesh on Global CP, it will create default policies which then will be synced to Zone CP.
However, default policies on Zone will be more specific so everything will stay as is.
