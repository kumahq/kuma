# XDS Test Client

Client allows emulating xDS connections without actual running of Envoy proxies. 

## Run
Run Kuma CP without Dataplane tokens, debug endpoint probably also will be useful:

```shell script
KUMA_DP_SERVER_AUTH_TYPE=none KUMA_DIAGNOSTICS_DEBUG_ENDPOINTS=true ./build/artifacts-darwin-amd64/kuma-cp/kuma-cp run
```

Run XDS Test Client:

```shell script
make run -C ./pkg/test/xds/client
```

## Env
- `NUM_OF_DATAPLANES` - total number of Dataplanes to emulate
- `KUMA_CP_ADDRESS` - address of Kuma CP 
