# Introduce Warnings for Resource API

- Status: accepted

## Context and Problem Statement

Resource API implements CRUDL operations for Kuma resources (policies, configs, etc.).
If resource is correct it returns `200 OK` or `201 Created`. 
If there is a validation error it returns `400 Bad Request`.
Today we're missing a way to return the information on deprecated fields or values.

Kubernetes implements similar functionality using the ValidatingWebhook. 
The [AdmissionResponse schema](https://kubernetes.io/docs/reference/config-api/apiserver-admission.v1/#admission-k8s-io-v1-AdmissionResponse) has a `warnings` field:

> warnings is a list of warning messages to return to the requesting API client. 
> Warning messages describe a problem the client making the API request should correct or be aware of. 
> Limit warnings to 120 characters if possible. 
> Warnings over 256 characters and large numbers of warnings may be truncated.

and `kubectl apply` outputs them to stderr:

```bash
$ kubectl apply -f policy.yaml
Warning: the field "X" is deprecated, please consider using "Y"
meshtimeout.kuma.io/timeout-global created
```

## Considered options

- Introduce a `warnings` field to the successful 200 and 201 responses.

## Decision Outcome

- Introduce a `warnings` field to the successful 200 and 201 responses.

## Implementation

The `response` list for Resource API should be updated:

```yaml
responses:
  '200':
    description: Updated
    content:
      application/json:
        schema:
          type: object
          properties:
            warnings:
              type: array
              items:
                type: string
  '201':
    description: Created
    content:
      application/json:
        schema:
          type: object
          properties:
            warnings:
              type: array
              items:
                type: string
```
