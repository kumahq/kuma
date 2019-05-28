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
