# Support projected service account tokens

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4121

## Context and Problem Statement

Kuma uses Kubernetes service account tokens (SAT) to identify data plane proxies to the control plane during bootstrapping. These tokens have no expiry. If the user wants to rotate the token, Kuma won't reload the newer token for authentication and will use the old one to reconnect. Currently, `kuma-dp` loads the token only once at the startup and does not detect the token's change. That is the problem and we want to support the case where the token is rotated and later used for authentication with control-plane.

## Considered Options

* Use GoogleGRPC configuration in Envoy
* Implement support for reading token from files in EnvoyGRPC

## Decision Outcome

Chosen option: "Use GoogleGRPC configuration in Envoy", because it works and does not require additional implementation in Envoy which would do the same. At the beginning we weren't sure if it works because of [issue](https://github.com/envoyproxy/envoy/issues/15380), but after validation, it seems that the problem described there doesn't occur anymore.

### Doubts
* should we add grpc connection timeout ? (if we are keeping the connection `forever/to the failure` we won't use token to reauthenticate)

### Positive Consequences

* support for rotation
* token won't be visible in config_dump, only file path

### Negative Consequences

* if we add connection timeout we might observe some logs about breaking connection

## Pros and Cons of the Options

### Use GoogleGRPC configuration in Envoy

`kuma-dp` allows providing a dataplane token or path to a file that has the token. When the user provides a path to the token, `kuma-dp` reads that token and adds it to the bootstrap request. To support rotation we are going to send both the token and path to the control plane in the bootstrap request. By default, we are going to use the current solution which uses `envoyGrpc` and `initial_metdata` for authentication. Token rotation needs to be enabled in the control-plane by setting `dpServer.auth.useTokenPath`. When the property is enabled, `kuma-cp` adds a path to the token to the Envoy configuration. In case there is no path, the `Envoy` will receive a configuration that doesn't support reloading.

Envoy configuration:
```json
  "ads_config": {
    "api_type": "GRPC",
    "grpc_services": [
     {
      "google_grpc": {
       "target_uri": "<kuma_cp_hostname>:5678",
       "channel_credentials": {
        "ssl_credentials": {
         "root_certs": {
          ...
         }
        }
       },
       "call_credentials": [
        {
         "from_plugin": {
          "name": "envoy.grpc_credentials.file_based_metadata",
          "typed_config": {
           "@type": "type.googleapis.com/envoy.config.grpc_credential.v3.FileBasedMetadataConfig",
           "secret_data": {
            "filename": "<token_file_path>"
           }
          }
         }
        }
       ],
       ...
      }
     }
    ],
    ...
   }
```

#### Advantages

* Does not require implementation in Envoy
* Safe to switch from one to another model, in case of an error we can disable

#### Disadvantages

* The flag adds some complexity (we can eventually deprecate and remove the flag)

### Implement support for reading tokens from files in EnvoyGRPC

We would have to implement support for reading file values in EnvoyGRPC, which is already supported by GoogleGRPC.

#### Disadvantages

* Already supported by GoogleGRPC

## Links

* https://github.com/kumahq/kuma/issues/4121
* https://github.com/envoyproxy/envoy/issues/15380
