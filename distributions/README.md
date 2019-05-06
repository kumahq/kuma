# Konvoy distributions

Build scripts to generate deb/rpm/tar packages and Docker images with `Konvoy` binary.

See [REPO_LAYOUT.md](../REPO_LAYOUT.md) for more details.

## Available commands

To see a list of available commands, run `make help`, e.g.
```bash
$ make help

help                           Display this help screen
clean                          Remove build files
download/binary                Download pre-built Konvoy binary from S3 bucket
build/package                  Build Konvoy package (deb/rpm/tar)
upload/package                 Upload Konvoy package (deb/rpm/tar) into S3 bucket
all/package                    Build and upload Konvoy package (deb/rpm/tar) into S3 bucket
build/image                    Build Konvoy Docker image
upload/image                   Upload Konvoy Docker image into S3 bucket
all/image                      Build and upload Konvoy Docker image into S3 bucket
```

## How to build a deb/rpm/tar package

### Preconditions

[Konvoy binary](../components/konvoy-binary) must have already been built
and uploaded into S3 bucket.

### Parameters

1. `COMMIT` (environment variable): Git SHA of a commit to build from
2. `BASE_IMAGE` (environment variable): one of
    * ubuntu:14.04
    * ubuntu:16.04
    * ubuntu:18.04
    * debian:8
    * debian:9
    * rhel:7
    * centos:7
    * amazonlinux:2
    * alpine:3.9

### Build steps

1. Download a pre-built [Konvoy binary](../components/konvoy-binary) from S3 bucket
2. Pull a [Docker image](images/fpm/Dockerfile.fpm) with [fpm](https://github.com/jordansissel/fpm) tool
3. Build a deb/rpm/tar package depending on the value of `BASE_IMAGE`
    * include [Konvoy binary](../components/konvoy-binary)
    * include [sample configuration](configs/)
4. Upload generated artifact into S3 bucket (requires proper credentials)

### Example

To build a package (no upload):

```bash
$ COMMIT=xxx BASE_IMAGE=centos:7 make local/package  
```

To build and upload a package (requires proper credentials):

```bash
$ COMMIT=xxx BASE_IMAGE=centos:7 make all/package
```

## How to build a Docker image

### Preconditions

[Konvoy binary](../components/konvoy-binary) must have already been built
and uploaded into S3 bucket.

### Parameters

1. `COMMIT` (environment variable): Git SHA of a commit to build from
2. `BASE_IMAGE` (environment variable): one of
    * ubuntu:14.04
    * ubuntu:16.04
    * ubuntu:18.04
    * debian:8
    * debian:9
    * rhel:7
    * centos:7
    * amazonlinux:2
    * alpine:3.9

### Build steps

1. Download a pre-built [Konvoy binary](../components/konvoy-binary) from S3 bucket
2. Build a Docker image on top of `BASE_IMAGE`
    * include [Konvoy binary](../components/konvoy-binary)
    * include [sample configuration](configs/)
3. Upload generated artifact into S3 bucket (requires proper credentials)
    * TODO: upload images into a Docker repository instead of S3 bucket  

### Example

To build a Docker image (no upload):

```bash
$ COMMIT=xxx BASE_IMAGE=centos:7 make local/image  
```

To build and upload a Docker image (requires proper credentials):

```bash
$ COMMIT=xxx BASE_IMAGE=centos:7 make all/image
```
