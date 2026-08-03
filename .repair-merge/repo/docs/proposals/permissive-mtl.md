# Permissive mTLS mode

## Context 

The goal for this proposal is to provide a more pragmatic way to enable mTLS in a service mesh. When working with existing applications, 
it is not practical to have mTLS enabled for everything or nothing, but we need to provide a more gradual way to do it.
Therefore, a "permissive" mTLS mode (as opposed to the "strict" one we use today) would allow a team to enable mTLS yet allow 
traffic that is not part of the zero-trust infrastructure to still be able to make requests without a strict validation of the 
certificates on incoming requests.

## Configuration

Add a new field `mode` under the `mtls.backends` section. The type of the field is enum with 2 values: `strict` and `permissive`.
To keep backwards compatibility a default value should be `strict`.

```yaml
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
spec:
  mtls:
    enabledBackend: ca-1
    backends:
      - name: ca-1
        type: builtin
        mode: strict|permissive
        dpCert:
          rotation:
            expiration: 1d
        conf:
          caCert:
            RSAbits: 2048
            expiration: 10y
```

## Implementation

Inbound listener should be configured with [TLS Inspector](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listener_filters/tls_inspector) 
to detect type of the traffic (`tls` or `plaintext`). Using [FilterChainMatch](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener_components.proto#config-listener-v3-filterchainmatch)
we can distinguish traffic by `transport_protocol`, it could be either `raw_buffer` or `tls`. If the type is `tls` then FilterChain stays the same, 
if the type is `raw_buffer` then filter chain is the same but without `ServerSideMTLS` section. 