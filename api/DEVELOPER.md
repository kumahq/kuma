# Developer documentation

## Pre-requirements

- `curl`
- `git`
- `go`
- `make`
- `zip`

For a quick start, use the official `golang` Docker image (which has most of these tools pre-installed), e.g.

```bash
docker run --rm -ti \
  --volume `pwd`:/go/src/github.com/kumahq/kuma/api \
  --workdir /go/src/github.com/kumahq/kuma/api \
  --env HOME=/tmp/home \
  --env GO111MODULE=on \
  golang:1.16 bash
export PATH=$HOME/bin:$PATH
apt update && apt install unzip
```

## Helper commands

```bash
make help
```

## Installing build tools

Run:

```bash
make dev/tools
```

## Processing .proto files 

Run:

```bash
make generate
```

## Building Golang code

Run:

```bash
make build
```
