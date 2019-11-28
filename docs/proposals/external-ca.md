# External Certificate Authority Proposal

## Context

To enable mTLS in the mesh, each dataplane has to receive a unique certificate. This certificate has to be signed by the root certificate which is unique for every mesh.
Right now, Kuma auto-generates root certificate when mesh is created. Root certificates are stored in Postgres on Universal and as a Secret on Kubernetes.
In many organisations there is already Certificate Authority, so there should be a way to use it with Kuma.

## Requirements

User wants to generate root certificate and provide it to Kuma.

## Proposed configuration model

```yaml
type: Mesh
name: default
mtls:
  ca:
    external: {}
  enabled: true
```

## How to provide certificates to Kuma

There are several ways we can implement providing root certs for Kuma

### File

```yaml
type: Mesh
name: default
mtls:
  ca:
    external:
      keyPath: /path/to/file
      certPath: /path/to/file
  enabled: true
```

It seems simple, but storing key files on the machines with control plane is not ideal.
The bigger problem though is that you have to synchronize files between CP instances. Once it's not in sync (Puppet fails, any other reason), the mesh will break.
We already have storage to share between CP instances and it's Postgres on Universal and Etcd on K8S.
  
### HTTP Server + kumactl

We can adopt the same pattern as with Dataplane Token Server.
The server is by default exposed only on `localhost`. Optionally user can expose server to public by generating cert+key for server and client certificates.

User can upload a certificate with key in two ways.

1) HTTP API
```bash
$ cat request.json
{
  "key": "-----BEGIN RSA PRIVATE KEY-----\nMLLLEpSDFAGKC24gArqiy1c3pFT3FSk5FE51A4ALAadeR...",
  "cert": "-----BEGIN CERTIFICATE-----\nZZXDFjC3An4gAwFBAFIJ4FFZ66emAA3AZA0GZSqGSGHt...",
  "mesh": "demo"
}
$ curl -XPUT localhost:5691/certificates --data @request.json
```

2) kumactl
```bash
$ cat key.pem
-----BEGIN RSA PRIVATE KEY-----\nMLLLEpSDFAGKC24gArqiy1c3pFT3FSk5FE51A4ALAadeR...
$ cat cert.pem
-----BEGIN CERTIFICATE-----\nZZXDFjC3An4gAwFBAFIJ4FFZ66emAA3AZA0GZSqGSGHt...
$ kumactl upload-certificate --mesh demo --key-file key.pem --cert-file cert.pem
```

This can be either yet another HTTP server in Kuma or we can embed this into Dataplane Token Server and rename the server.

### kuma-cp sub command

Right now, you can run the CP with `kuma-cp run`.
We can introduce new command:

```bash
$ cat key.pem
-----BEGIN RSA PRIVATE KEY-----\nMLLLEpSDFAGKC24gArqiy1c3pFT3FSk5FE51A4ALAadeR...
$ cat cert.pem
-----BEGIN CERTIFICATE-----\nZZXDFjC3An4gAwFBAFIJ4FFZ66emAA3AZA0GZSqGSGHt...
$ kuma-cp upload-certificate --mesh demo --key-file key.pem --cert-file cert.pem
```

## Flow

Either with `HTTP Server + kumactl` or `kuma-cp sub command` we should first upload the certificate.
Then when we apply a mesh with `enabled: true` and `ca: external`, we should validate if certificate is present.

## Certificate modification

Having only builtin CA without ways to modify it, we did not support certificate changes.
Once you change the certificate, current dataplanes will have certs signed by old root cert and new dataplanes will receive certs signed by new root cert, therefore they won't be able to communicate.

Eventually we should introduce the support for certificate rotation, so user can revoke certificate or migrate CA without downtime, but this is a case for another proposal.
For now we can adopt simpler approach:

1. Block update operation
2. Do not block it, but document the behaviour
3. Do not block it, but when user tries to upload certificate for the mesh that already has one we should notify them:
  * in case of HTTP API: "You are trying to update existing certificate. To do that, you should stop dataplanes in the mesh, upload new cert and start them again. Add ?force=true param to the request to upload the new certificate"
  * in case of kumactl: "You are trying to update existing certificate. To do that, you should stop dataplanes in the mesh, upload new cert and start them again. Add --force argument to upload the new certificate"
  
## CA modification

Modifying type of CA (ex. builtin -> external) of the Mesh means that certificates will change which is the same problem as described above.
We can also warn users and introduce `?force=true` to HTTP API and `kumactl apply --force`.

This would work on Universal, but on Kubernetes I don't think that there is a proper equivalent for it.
https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#apply