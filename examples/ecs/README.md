# Running Kuma on ECS

The ECS example templates are divided into three sections: `vpc`, `kuma-cp` and `workload`. The Kuma CP deployment can be `standalone`, `global` or `zone`.

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
The command to deploy the `kuma-cp` stack in the standalone mode is as follows

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp \
    --template-file kuma-cp-standalone.yaml
```

The `kuma-vpc` stack is the default for the `VPCStackName` parameter. Note that `AllowedCidr` parameter and override it accordingly to enable access to Kuma CP ports.

To remove the `kuma-cp` stack use:
```shell
aws cloudformation delete-stack --stack-name kuma-cp
```

### Global

Deploying a global control plane is simple as it does not have many setting to tune.

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp-global \
    --template-file kuma-cp-global.yaml
```

### Zone

Setting up a zone `kuma-cp` is a three-step process. First, deploy the kuma-cp itself:

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp \
    --template-file kuma-cp-zone.yaml
```


#### OPTIONAL: Configure `kumactl` to access the API 
Find the public IP address fo the zone or standalone `kuma-cp` and use it in the command below.

```bash
export PUBLIC_IP=<ip address>
kumactl config control-planes add --name=ecs --address=http://$PUBLIC_IP:5681 --overwrite
```

### Install the Zone Ingress

For cross-zone communication Kuma needs the Ingress DP deployed. As every dataplane (see details in the `workload` chapter below) it needs a dataplane token generated 

```shell
ssh root@<kuma-cp-zone-ip> "wget --header='Content-Type: application/json' --post-data='{\"mesh\": \"default\", \"type\": \"ingress\"}' -qO- http://localhost:5681/tokens"
```

Then simply deploy the ingress itself:

```shell
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name ingress \
    --template-file zone-ingress.yaml \
    --parameter-overrides \
      DPToken="<DP_TOKEN_VALUE>"
```

### Install the Kuma DNS

The services within the Kuma mesh are exposed through their names (as defined in the `kuma.io/service` tag) in the `.mesh` DNS zone. In the default workload example that would be `httpbin.mesh`.
Run the following command to create the necessary Forwarding rules in Route 53 and leverage the integrated DNS server in `kuma-cp`.

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-dns \
    --template-file kuma-dns.yaml \
    --parameter-overrides \
      DNSServer=<kuma-cp-ip>
```

Note: We strongly recommend exposing the Kuma-CP instances behind a load balancer, and use that IP as the `DNSServer` parameter. This will ensure a more robust operation during upgrades, restarts and re-configurations. 

### Install the workload

The `workload` template provides a basic example how `kuma-dp` can be run as a sidecar container alongside an arbitrary, single port service container.
In order to run `kuma-dp` container, we have to issue a token. Token could be generated using Admin API of the Kuma CP.

In this example we'll show the simplest form to generate it by executing this command alongside the `kuma-cp`:
```bash
ssh root@<kuma-cp-ip> "wget --header='Content-Type: application/json' --post-data='{\"mesh\": \"default\"}' -qO- http://localhost:5681/tokens"
```

The passowrd is `root`, as noted in the beginning these are sample deployments and it is not adviseable 

The generated token is valid for all Dataplanes in the `default` mesh. Kuma also allows you to generate tokens based
on Dataplane's Name and Tags.

Note: Kuma allows much more advanced and secure way to expose the `/tokens` endpoint. For this it needs to have `HTTPS` endpoint configured
on port `5682` as well as client ceritificate setup for authentication. The full procedure is available in Kuma Security documentation 
[Data plane proxy authentication](https://kuma.io/docs/1.0.5/documentation/security/#data-plane-proxy-to-control-plane-communication),
[User to control plane communication](https://kuma.io/docs/1.0.5/documentation/security/#user-to-control-plane-communication)

#### Standalone

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name workload \
    --template-file workload.yaml \
    --parameter-overrides \
      DesiredCount=2 \
      DPToken="<DP_TOKEN_VALUE>"
```

#### Zone

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name workload \
    --template-file workload.yaml \
    --parameter-overrides \
      DesiredCount=2 \
      DPToken="<DP_TOKEN_VALUE>" \
      CPAddress="https://zone-1-controlplane.kuma.io:5678"
```

The `workload` template has a lot of parameters, so it can be customized for many scenarios, with different workload images, service name and port etc. Find more information in the template itself.

## A second zone example
Here is an example how to run a second workload with the same SSH server in a second zone:

First create the second zone:

```shell
kumactl generate tls-certificate --type=server --cp-hostname zone-2-controlplane.kuma.io
export KEY=$(cat key.pem)
export CERT=$(cat cert.pem)
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp-zone-2 \
    --template-file kuma-cp-zone.yaml \
    --parameter-overrides \
      ServerCert=$CERT \
      ServerKey=$KEY \
      Zone=zone-2
```

```shell
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name workload-2 \
    --template-file workload.yaml \
    --parameter-overrides \
      WorkloadName=ssh \
      WorkloadImage=sickp/alpine-sshd:latest \
      WorkloadManagementPort=22 \
      CPAddress="https://zone-2-controlplane.kuma.io:5678" \
      DPToken="<DP_TOKEN_VALUE>"
```

Finally log in the new workload container and access the `httpbin.mesh` service:

```shell
ssh roo@<workload-2-ip>
wget -qO- httpbin.mesh
```

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
 * call `kuma-cp migrate up` with all aforementioned environments set
