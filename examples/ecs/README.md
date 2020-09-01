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
ssh root@$PUBLIC_IP 'apk update 2>&1 > /dev/null && apk add curl 2>&1 && curl -s -XPOST -H "Content-Type: application/json" --data \'{"name": "dp-echo-1", "mesh": "default"}\' http://localhost:5679/tokens'
```



```


# Example with custom parameters

```bash
aws cloudformation deploy \
    --capabilities CAPABILITY_IAM \
    --stack-name kuma-cp \
    --template-file kuma-cp-standalone.yaml \
    --parameter-overrides Image="nickolaev/kuma-cp:latest" \
    --parameter-overrides AllowedCidr="82.146.27.0/24"
```


