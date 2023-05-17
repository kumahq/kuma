# Upgrade dependencies

Some dependencies upgrades require extra steps

## Go Control Plane

To upgrade Go Control Plane you need to

* Upgrade the version in go.mod
* Regenerate imports `make generate/envoy-imports`
* Fix copy pasted parts of the control plane `pkg/util/xds`

## Envoy

* Find and replace the old version with a new one
* Upload binaries for Universal of every distro to PULP

## Kubernetes client

* Upgrade the version in go.mod
* Upgrade manually `plugins/bootstrap/k8s/cache`
