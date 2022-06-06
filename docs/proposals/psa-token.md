# Support projected service account tokens
 
## Context
 
Kuma uses Kubernetes service account tokens (SAT) to identify dataplanes to the control-plane during bootstrapping. These tokens have no expiry. In case the user wants to rotate the token kuma won't reload the newer token for authentication and use the old (invalid) to reconnect.
 
## Requirements
 
* Envoy can read rotated token and use it to reauthenticate
 
## Design
 
`kuma-dp` allows providing a dataplane token or path to a file that has the token. When the user provides a path to the token, `kuma-dp` reads that token and adds it to the bootstrap request. To support rotation we are going to send both information to the `kuma-cp` in the bootstrap request (token and path). By default, we are going to use the current solution which uses `envoyGrpc` and `initial_metdata` for authentication. Configuration to enable using token rotation needs to be enabled in the control-plane by enabling `dpServer.auth.useTokenPath`.
 
### Envoy configuration
 
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
 
### kuma-dp behavior
 
`kuma-dp` at the start reads a token from a provided path and adds two pieces of information to the bootstrap request: `path` and `token`. In case there is no path we are taking only tokens.
 
### kuma-cp behavior
 
We are going to expose the configuration property under `dpServer.auth.useTokenPath` in `kuma-cp` configuration to enable specified behavior. When the property is enabled, `kuma-cp` adds a path to the token to the configuration of an `Envoy`. In case there is no path the `Envoy` will receive a configuration that doesn't support reloading.
 
## Doubts
 
* should we add grpc connection timeout ? (if we are keeping the connection `forever/to the failure` we won't use token to reauthenticate)
