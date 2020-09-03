# Running Kuma on ECS

The ECS example templates are divided into three sections: `vpc`, `kuma-cp` and `workload`. The Kuma CP deployment can be `standalone`, `global` or `remote`. 

## Deploy the VPC stack

This stack is not parametrised. It provides the basic VPC setup that is the base for deploying Kuma-CP and the workloads.

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-vpc \
    --template-file kuma-vpc.yaml
```

NOTE: It is recommended that the stack name is `kuma-vpc`, as that is the default further in the examples.

## Deploy Kuma Control Plane

The examples provide separate templates for each Kuma CP mode. Follow the instructions below depending on the needed usecase.

### Standalone
The command to deploy the `kuma-cp` stack in snadalone mode is as follows

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp \
    --template-file kuma-cp-standalone.yaml
```

The `kuma-vpc` stack is the default for the `VPCStackName` parameter. Note that `AllowedCidr` parameter and override it accordingly to enable access to Kuma CP ports accordingly.

### Global

Deplyng a global control plane is simple as it does not have many setting to tune.

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp-global \
    --template-file kuma-cp-global.yaml
```


### Remote

Setting up a remote `kuma-cp` is a two step process. First, deploy the kuma-cp itself:

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp \
    --template-file kuma-cp-remote.yaml
```

Then add a resource in the global (see how to configure `kumactl` in the next session)

#### ECS/Universal
```bash
echo "type: Zone
name: zone-1
ingress:
  address: <zone-ingress-address>" | kumactl apply -f -
```

#### Kubernetes

```bash
echo "apiVersion: kuma.io/v1alpha1
kind: Zone
mesh: default
metadata:
  name: zone-1
spec:
  ingress:
    address:  <zone-ingress-address>" | kubectl apply -f -
```

Where `<zone-ingress-address>` is composed of the public address of the remote kuma-cp and the port assigned for the Ingress.


## OPTIONAL: Configure `kumactl` to access the API 
Find the public IP address fo the remote or standalone `kuma-cp` and use it in the command below.

```bash
export PUBLIC_IP=<ip address>
kumactl config control-planes add --name=ecs --address=http://$PUBLIC_IP:5681 --overwrite
```

### Install the workload


The `workload` template provides a basic example how `kuma-dp` can be run as a sidecar container alongside an arbitrary, single port service container.

#### Standalone
```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name workload \
    --template-file workload.yaml \
    --parameter-overrides \
      DesiredCount=2
```

#### Remote
```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name workload \
    --template-file workload.yaml \
    --parameter-overrides \
      DesiredCount=2 \
      CPAddress="http://zone-1-controlplane.kuma.io:5681"
```

The `workload` template has a lot of parameters so it can be customized for many scenarios, with different workload images, service name and port etc. Find more information in the template itself.


# Future work

### Persistent storage:

The default mode for deploying `kuma-cp` in these examples is to use the ephemeral, in-memory storage for resources. This imposes 2 limitations in the examples: a) no `kuma-cp` replicase are possibe, as they would rely on sharing some common state over the shared persisitne storage; b) all resources (policies, zones, dataplanes) shall be lost upon `kuma-cp` restart.

 * AWS::RDS::DBInstance
 * Set-up the relevant env variables in `TaskDefinitionKumaCP` as follows:
    	KUMA_STORE_TYPE=postgres
    	KUMA_STORE_POSTGRES_HOST=localhost
    	KUMA_STORE_POSTGRES_PORT=5432
    	KUMA_STORE_POSTGRES_USER=kuma
    	KUMA_STORE_POSTGRES_PASSWORD=kuma
    	KUMA_STORE_POSTGRES_DB_NAME=kuma 
    	KUMA_STORE_POSTGRES_TLS_MODE= disable | verifyNone | verifyCa | verifyFull
    	KUMA_STORE_POSTGRES_TLS_CERT_PATH=
    	KUMA_STORE_POSTGRES_TLS_KEY_PATH=
    	KUMA_STORE_POSTGRES_TLS_CA_PATH=


