# kubectl

Kubectl image based on alpine image. Contains default tools: sh, wget, ...

## Build

```bash
docker build --build-arg ARCH=amd64 --build-arg BASE_IMAGE=amd64/alpine:latest --tag kumahq/kubectl .
```

```bash
docker build --build-arg ARCH=arm64 --build-arg BASE_IMAGE=arm64v8/alpine:latest --tag kumahq/kubectl .
```

## How to push?

Run `./docker-build-and-publish.sh` with parameter which is equal to kubernetes version that you want to release for.
