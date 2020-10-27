Certificates generated with following commands.

## Client
```bash
kumactl generate tls-certificate --type=client --key-file=client.key --cert-file=client.pem
```

## Server
```bash
kumactl generate tls-certificate --type=server --cp-hostname=kuma-control-plane
```
