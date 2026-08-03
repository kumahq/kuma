# External Certificate Authority Proposal

## Context

To enable mTLS in the mesh, each dataplane has to receive a unique certificate. This certificate has to be signed by the CA certificate which is unique for every mesh.
Right now, Kuma auto-generates root CA certificate when mesh is created. Root CA certificates are stored in Postgres on Universal and as a Secret on Kubernetes.
In many organisations there is already Certificate Authority, so there should be a way to use it with Kuma.

## Requirements

A user wants to provide a signing CA certificate to Kuma.

## Proposed configuration model

```yaml
type: Mesh
name: default
mtls:
  ca:
    provided: {}
  enabled: true
```

For the initial implementation there are following constraints:

1. It will not be possible to change provided CA cert while mTLS is enabled on a given Mesh.
Otherwise the current dataplanes will be signed by old CA cert and new dataplanes with new CA cert, therefore they won't be able to communicate.
To be able to change the CA cert, a user must first disable mTLS on that Mesh.

2. It will not be possible to change CA type (e.g., builtin => provided) while mTLS is enabled on a given Mesh.
This is the same problem as with the previous point.
To be able to change CA type, a user must first disable mTLS on that Mesh.
Support for certificate rotation and multiple CA certs will be eventually introduced, but for the first implementation we want to keep it simple.

## How to provide certificates to Kuma

Provided CA certificate has to fulfill requirements described [here](https://github.com/spiffe/spiffe/blob/master/standards/X509-SVID.md#32-signing-certificates)

There are several ways we can implement providing root certs for Kuma

### ~~File~~

```yaml
type: Mesh
name: default
mtls:
  ca:
    provided:
      keyPath: /path/to/file
      certPath: /path/to/file
  enabled: true
```

It seems simple, but storing key files on the machines with control plane is not ideal.
The bigger problem though is that you have to synchronize files between CP instances. Once it's not in sync (Puppet fails, any other reason), the mesh will break.
We already have storage to share between CP instances and it's Postgres on Universal and Etcd on K8S.

Additionally, until we have AuthN & AuthZ on the control plane, it's not safe to let users provide CA cert and key via REST API
(e.g., a malicious user could change somebody else's Mesh to use CA cert and key files of a Mesh he has control over)
  
### HTTP(S) Server + kumactl

We can adopt the same pattern as with Dataplane Token Server. We cannot add it to the API Server on 5681, because otherwise anyone would be able to override certificates. 

The server is by default exposed over HTTP only on `localhost`, so only operators of the Control Plane with access to the machine can upload certificates.

Optionally the server can be exposed to public over HTTPS by generating cert+key for the server and client certificates.
We can reuse certificates from Dataplane Token Server although it's better to maintain separate lists of clients allowed to send a requests to each server.

User can upload a certificate with key in two ways.

1) HTTP API
```bash
$ cat request.json
{
  "key": "-----BEGIN RSA PRIVATE KEY-----\nMLLLEpSDFAGKC24gArqiy1c3pFT3FSk5FE51A4ALAadeR...",
  "cert": "-----BEGIN CERTIFICATE-----\nZZXDFjC3An4gAwFBAFIJ4FFZ66emAA3AZA0GZSqGSGHt...",
}
$ curl -XPOST localhost:5691/meshes/demo/ca/provided/certificates --data @request.json
```

2) kumactl
```bash
$ cat key.pem
-----BEGIN RSA PRIVATE KEY-----\nMLLLEpSDFAGKC24gArqiy1c3pFT3FSk5FE51A4ALAadeR...
$ cat cert.pem
-----BEGIN CERTIFICATE-----\nZZXDFjC3An4gAwFBAFIJ4FFZ66emAA3AZA0GZSqGSGHt...
$ kumactl manage ca provided certificates add --mesh demo --key-file key.pem --cert-file cert.pem
```

This can be either yet another HTTP server in Kuma or we can embed this into Dataplane Token Server and rename the server.
If we embed this into Dataplane Token Server, the list of client will be shared for both use cases.
If we create another server we would need to add `--ca-manager-client-cert` and `--ca-manager-client-key` to `kumactl add control-plane`,
but a user is forced to provide those only when they want to access the manage functionality outside of CP machine. 

Pros:
* More flexible to automate since there is an API next to the CLI and it can be executed from the remote machine
* Consistent behaviour with Dataplane Token Server

Cons:
* We need to create yet another server + port

### ~~kuma-cp sub command~~

Right now, you can run the CP with `kuma-cp run`.
We can introduce new command:

```bash
$ cat key.pem
-----BEGIN RSA PRIVATE KEY-----\nMLLLEpSDFAGKC24gArqiy1c3pFT3FSk5FE51A4ALAadeR...
$ cat cert.pem
-----BEGIN CERTIFICATE-----\nZZXDFjC3An4gAwFBAFIJ4FFZ66emAA3AZA0GZSqGSGHt...
$ kuma-cp manage ca provided certificates add --mesh demo --key-file key.pem --cert-file cert.pem
```

For this command to work, we need to provide a config just like we do with starting the control plane.

Pros:
* We are not polluting kumactl with operations that should be available only for mesh operators
* Easier implementation since we don't have to deal with HTTP(S)

Cons:
* Harder to automate from machine other than the one with CP. In theory a user can take a kuma-cp binary with CP config and execute it elsewhere, but for example will this machine have access to Postgres?

## Flow

Either with `HTTP Server + kumactl` or `kuma-cp sub command` we should first upload the certificate.
Then when we apply a mesh with `enabled: true` and `ca: external`, we should validate if certificate is present.

## Securing Postgres

Additionally, we need to support a connection to Postgres over TLS so the CA cert transfer is secured on Kuma CP <-> Postgres connection.

## Summary

We've chosen providing certificates with HTTP + kumactl as this seems to be the most flexible and consistent option.
We will use the same server as for Dataplane Token for the sake of simplicity for now.