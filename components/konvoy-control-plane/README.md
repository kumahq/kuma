# Konvoy Control Plane

Universal Control Plane for Envoy-based Service Mesh.

## Building

Run:

```bash
make build
```

## Running locally

Run:

```bash
make run
```

Make a test `Discovery` request to `LDS`:

```bash
make curl/listeners
```

Make a test `Discovery` request to `CDS`:

```bash
make curl/clusters
```

## Pointing Envoy at Control Plane

Start `Control Plane`:

```bash
make run
```

Assuming `envoy` binary is on your `PATH`, run:

```bash
make run/example/envoy
```

Dump effective Envoy config:

```bash
make config_dump/example/envoy
```
