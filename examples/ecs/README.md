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

The `kuma-vpc` stack is the default for the `StackName` parameter. Note that `AllowedCidr` parameter and override it accordingly to enable access to Kuma CP ports accordingly.

### Global

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp-global \
    --template-file kuma-cp-global.yaml
```




## Generating the DP token

Before starting a workload, One needs

Find the public IP address of the `kuma-cp` ECS Task in either `standalone` or `remote` mode and replace `<ip address>` in the followiing instructions.
```bash
`export PUBLIC_IP=<ip address>`
ssh root@$PUBLIC_IP 'apk update 2>&1 > /dev/null && apk add curl 2>&1 && curl -s -XPOST -H "Content-Type: application/json" --data \'{"name": "httpbin", "mesh": "default"}\' http://localhost:5679/tokens'
```

Save the generated token and use it when installing a new workload.


### Install the workload


The `workload` template provides a basic example how `kuma-dp` can be run as a sidecar container alongside an arbitrary, single port service container. As a prerequisite one needs to obtain a token as already explained, then deploy the default workload with `httpbin` container.

```bash
export DPTOKEN=<token>
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name workload \
    --template-file workload.yaml \
    --parameter-overrides \
      Image="nickolaev/kuma-dp:latest" \
      DPToken=$DPTOKEN
```

The `workload` template has a lot fo parameters so it cna be customized for many scenarios, with different workload images, service name and port etc. Find more information in the template itself.

# Examples with custom parameters

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp \
    --template-file kuma-cp-standalone.yaml \
    --parameter-overrides \
      Image="nickolaev/kuma-cp:latest" \
      AllowedCidr="82.146.27.0/24"
```

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp-global \
    --template-file kuma-cp-global.yaml \
    --parameter-overrides \
      Image="nickolaev/kuma-cp:latest" \
      AllowedCidr="82.146.27.0/24"
```

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp \
    --template-file kuma-cp-remote.yaml \
    --parameter-overrides \
      Image="nickolaev/kuma-cp:latest" \
      AllowedCidr="82.146.27.0/24"
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


