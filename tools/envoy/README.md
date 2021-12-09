# Tools for Envoy

The current directory contains tools for building, publishing and fetching Envoy binaries.

There is a new Makefile target `build/envoy` that places an `envoy` binary in `build/artifacts-$GOOS-$GOARCH/envoy` directory.
The default behaviour of that target – fetching binaries from [download.konghq.com](download.konghq.com) since it makes more sense for
overwhelming majority of users. However, there is a variable `BUILD_ENVOY_FROM_SOURCES` that allows to build Envoy from 
source code. 

### Usage

Download the latest supported Envoy binary for your host OS: 
```shell
$ make build/envoy
```

Download the latest supported Envoy binary for specified system:
```shell
$ GOOS=linux make build/envoy # supported OS: linux, centos and darwin
```

Download the specific Envoy tag:
```shell
$ ENVOY_TAG=v1.18.4 make build/envoy
```

Download the specific Envoy commit hash (if it exists in [download.konghq.com](download.konghq.com)):
```shell
$ ENVOY_TAG=bef18019d8fc33a4ed6aca3679aff2100241ac5e make build/envoy
```

If desired commit hash doesn't exist, it could be built from sources:
```shell
$ ENVOY_TAG=bef18019d8fc33a4ed6aca3679aff2100241ac5e BUILD_ENVOY_FROM_SOURCES=true make build/envoy
```

When building from sources its still possible to specify OS:
```shell
$ GOOS=linux ENVOY_TAG=bef18019d8fc33a4ed6aca3679aff2100241ac5e BUILD_ENVOY_FROM_SOURCES=true make build/envoy
```
