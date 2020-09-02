# Running Kuma on ECS

## Deploy the networking stack

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-vpc \
    --template-file kuma-vpc.yaml
```

## Deploy Kuma Control Plane

### Standalone

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp \
    --template-file kuma-cp-standalone.yaml
```


#### Generating the DP token

`export PUBLIC_IP=<ip address>` is the public IP address of the `kuma-cp` ECS Task.
```bash
ssh root@$PUBLIC_IP 'apk update 2>&1 > /dev/null && apk add curl 2>&1 && curl -s -XPOST -H "Content-Type: application/json" --data \'{"name": "httpbin", "mesh": "default"}\' http://localhost:5679/tokens'
```




### Install the workload

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name workload \
    --template-file workload.yaml \
    --parameter-overrides \
      Image="nickolaev/kuma-dp:latest" \
      DPToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJOYW1lIjoiZHAtZWNoby0xIiwiTWVzaCI6ImRlZmF1bHQifQ.RcVzM8QG3U6A66W3rvw3LzzB8qPfiv4O7CYUyyVs_iU
```


# Example with custom parameters

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp \
    --template-file kuma-cp-standalone.yaml \
    --parameter-overrides \
      Image="nickolaev/kuma-cp:latest" \
      AllowedCidr="82.146.27.0/24"
```

# Future work

### Persistent storage:
 * AWS::RDS::DBInstance
 * Set env as follows:
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


