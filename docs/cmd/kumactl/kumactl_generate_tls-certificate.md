## kumactl generate tls-certificate

Generate a TLS certificate

### Synopsis

Generate self signed key and certificate pair that can be used for example in Dataplane Token Server setup.

```
kumactl generate tls-certificate --type=server|client --hostname=HOST1[,HOST2...] [flags]
```

### Examples

```

  # Generate a TLS certificate for use by an HTTPS server, i.e. by the Dataplane Token server
  kumactl generate tls-certificate --type=server --hostname=localhost

  # Generate a TLS certificate for use by a client of an HTTPS server, i.e. by the 'kumactl generate dataplane-token' command
  kumactl generate tls-certificate --type=client --hostname=dataplane-1
```

### Options

```
      --cert-file string   path to a file with a generated TLS certificate ('-' for stdout) (default "cert.pem")
  -h, --help               help for tls-certificate
      --hostname strings   DNS hostname(s) to issue the certificate for
      --key-file string    path to a file with a generated private key ('-' for stdout) (default "key.pem")
      --key-type string    type of the private key: one of rsa|ecdsa
      --type string        type of the certificate: one of client|server
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

