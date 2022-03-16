## kumactl generate signing-key

Generate signing keys

### Synopsis

Generate a private key for signing tokens.

```
kumactl generate signing-key [flags]
```

### Examples

```

Generate a new signing key to rotate tokens (for example user-token) on Universal.
$ echo "
type: GlobalSecret
name: user-token-signing-key-0002
data: {{ key }}
" | kumactl apply --var key=$(kumactl generate signing-key) -f -

Generate a new signing key to rotate tokens (for example user-token) on Kubernetes.
$ TOKEN="$(kumactl generate signing-key)" && echo "
apiVersion: v1
data:
  value: $TOKEN
kind: Secret
metadata:
  name: user-token-signing-key-0002
  namespace: kong-mesh-system
type: system.kuma.io/global-secret
" | kubectl apply -f - 

```

### Options

```
  -h, --help   help for signing-key
```

### Options inherited from parent commands

```
      --api-timeout duration   the timeout for api calls. It includes connection time, any redirects, and reading the response body. A timeout of zero means no timeout (default 1m0s)
      --config-file string     path to the configuration file to use
      --log-level string       log level: one of off|info|debug (default "off")
  -m, --mesh string            mesh to use (default "default")
      --no-config              if set no config file and config directory will be created
```

### SEE ALSO

* [kumactl generate](kumactl_generate.md)	 - Generate resources, tokens, etc

