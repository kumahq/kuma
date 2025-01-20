# XDS Test Client

Client allows emulating xDS connections without actual running of Envoy proxies. 

## Run
Run Kuma CP without Dataplane tokens, debug endpoint probably also will be useful:

```shell script
KUMA_DP_SERVER_AUTHN_DP_PROXY_TYPE=none KUMA_DIAGNOSTICS_DEBUG_ENDPOINTS=true ./build/artifacts-darwin-amd64/kuma-cp/kuma-cp run
```

Run XDS Test Client:

```shell script
make run/xds-client
```

## Env
- `NUM_OF_DATAPLANES` - total number of Dataplanes to emulate
- `NUM_OF_SERVICES` - total number of services to emulate
- `KUMA_CP_ADDRESS` - address of Kuma CP 
