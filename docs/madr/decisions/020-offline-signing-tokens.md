# Offline signing tokens

* Status: accepted

Technical Story: https://github.com/kumahq/kuma/issues/4031

## Context and Problem Statement

Kuma provides multiple tokens for different lines of security
* User authenticating to Kuma API Server
* Data plane proxy authenticating to Kuma DP Server
* Zone proxy authenticating to Kuma DP Server

Right now, a user has to use Kuma CP to issue the tokens.
While this works, some users may prefer using their existing token issuer.
We can provide an interface of a token (document required fields) and a way to provide public key of a token.
This way a user can use different system to issue token, but let Kuma CP validate the token.
This improves security posture if you already have a safe way of managing signing key for your tokens.

## Considered Options

* CP configuration and kumactl improvements

## Decision Outcome

Chosen option: "CP configuration and kumactl improvements".

## Pros and Cons of the Options

### User token

#### CP configuration

```yaml
apiServer:
  authn:
    type: tokens
    tokens:
      bootstrapAdminToken: true
      enableIssuer: true
      validator:
        useSecrets: true
        publicKeys:
        - kid: 123
          keyFile: /tmp/key.pem # either keyFile or key can be defined
          key: |
            -----BEGIN RSA PUBLIC KEY-----
            MIIEogIBAAKCAQEAy0KtfI7O0TJ00............
            -----END RSA PUBLIC KEY-----
          mapping:
            username: user
            groups: teams
```

*bootstrapAdminToken* - we already have it. Creates the admin token from the default signing key. Cannot be set to true if *enableIssuer* is set to false
*enableIssuer* - if true you can generate tokens using Kuma CP API. It will also create a default signing key if one is missing.
*validator.useSecrets* - if true, Kuma will try to access a Secret if the key is missing on a public keys list.
*validator.publicKeys* - list of public keys. Every token has to have `kid`, so we know which key to use to validate it. Public keys takes precedence over signing keys stored as Secret.
*validator.mapping* - list of fields in claims to map. Users can provide their own tokens with fields that are different from what we require.

#### Use cases

**Use tokens issued by the CP**
```yaml
apiServer:
  authn:
    type: tokens
    tokens:
      enableIssuer: true
      validator:
        useSecrets: true
```

**Use tokens issued by the CP and an external issuer. Tokens still can be issued by the CP**
```yaml
apiServer:
  authn:
    type: tokens
    tokens:
      enableIssuer: true
      validator:
        useSecrets: true
        publicKeys:
        - kid: 123
          key: ...
```

**Use tokens issued by the CP and external issuer. Tokens cannot be issued by the CP**
```yaml
apiServer:
  authn:
    type: tokens
    tokens:
      enableIssuer: false
      validator:
        useSecrets: true
        publicKeys:
          - kid: 123
            key: ...
```

**Use tokens issued by external issuer**
```yaml
apiServer:
  authn:
    type: tokens
    tokens:
      enableIssuer: false
      validator:
        useSecrets: false
        publicKeys:
          - kid: 123
            key: ...
```

#### Claims

* `name (string)` - name of the user.
* `groups ([]string)` - list of groups

### Dataplane Token

#### CP configuration

```yaml
dpServer:
  authn:
    dpp:
      type: tokens # or serviceAccountToken or none
      tokens:
        enableIssuer: true
        validator:
          useSecrets: true
          publicKeys:
            - kid: 123
              key: ...
              mesh: default # <- DPP token is mesh scoped
```

We won't introduce `mapping` yet as it's very unlikely at this moment that a user would like to map it from existing token.

Additionally, we deprecate this config
```yaml
dpServer:
  auth:
    type: "" # ENV: KUMA_DP_SERVER_AUTH_TYPE
```

Because it's common for both data plane proxies and zone proxies. Since both uses different way of auth, it should be separate.

#### Claims

* `name (string)` - name of the DPP.
* `mesh (string)` - name of the mesh.
* `tags (map[string][]string)` - tags of the DPP.

### Zone token

#### CP configuration

```yaml
dpServer:
  authn:
    zoneProxies:
      type: tokens # or serviceAccountToken or none
      tokens:
        enableIssuer: true
        validator:
          useSecrets: true
          publicKeys:
            - kid: 123
              key: ...
```

#### Claims

* `zone (string)` - name of the zone of zone proxy.
* `scope ([]string)` - `ingress`, `egress`. So you can reuse token for all zone proxies if you want.

### Observability

We need to introduce a metric to see which `kid` was used to validate the token.
This will help users to know when they can migrate signing keys.

### kumactl

`kumactl` right now offers token generation, but it uses CP connection. To support offline signing, we need to improve `kumactl`.

#### Generating signing key

Currently, we have a command to generate a signing key that does not require CP connection.
```
$ kumactl generate signing-key
LS0tLS1CRUdJTiBSU0EgUFJ........
```
However, it produces a private key
```
$ kumactl generate signing-key | base64 -d
-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAy0KtfI7O0TJ00............
-----END RSA PRIVATE KEY-----
```

We need to have a way of extracting public key from it.
```
$ kumactl generate signing-key --format=pem > key.pem
$ kumactl generate public-key --private-signing-key-path=key.pem > public-key.pem
$ cat public-key.pem
-----BEGIN RSA PUBLIC KEY-----
MIIEogIBAAKCAQEAy0KtfI7O0TJ00............
-----END RSA PUBLIC KEY-----
```
The content of public-key.pem can be put in `apiServer.authn.tokens.validator.publickeys[].key`

Generating signing keys is the same for every type of the token.

#### Issuing tokens

Then, we can also use generated private key to issue tokens without connection to the CP
```
$ kumactl generate user-token \
    --name=john \
    --valid-for=30h \
    --kid=1
    --signing-key-path=key.pem 
```
Whenever the `--signing-key-path` flag is present, `kumactl` will sign the token offline. 
Additionally, `--kid` is required, it is not embedded in RSA key, so it has to be explicitly specified. 

We will also add `--signing-key-path` and `--kid` arguments to other token generation commands. 
