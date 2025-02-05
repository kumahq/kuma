# Resource Identifier

* Status: accepted

## Context and Problem Statement

Kuma multizone mode allows synchronizing state between multiple clusters. Zone Control Planes connect to the Global Control Plane, forming a tree with the Global CP as the root.

There are several resource synchronization patterns:
- A resource is created on the Global CP and synced down to all Zone CPs.
- A resource is created on a Zone CP and synced to the Global CP.
- A resource is created on a Zone CP and synced to other Zone CPs.

For example, the Zone CP in `zone-1` can possess the following MeshTimeouts:
- MeshTimeouts created in `zone-1`
- MeshTimeouts created on the Global CP and synced to `zone-1`
- MeshTimeouts created in `zone-2` and synced to `zone-1`

To uniquely identify a resource regardless of its place of origin, Kuma uses the following Go structure:

```go
type ResourceIdentifier struct {
    Name      string
    Mesh      string
    Namespace string
    Zone      string
}
```

When mixing identifiers of different resource types, an extended version is used:

```go
type TypedResourceIdentifier struct {
    ResourceIdentifier

    ResourceType ResourceType
    SectionName  string
}
```

Currently, the resource identifier lacks a string representation and is not exposed to the Kuma public API. The goal of this MADR is to propose a format.

## Decision Drivers

- Possible to use the identifier in a URL path as `:5681/_rules/<identifier>`
- Human-readable: users should be able to type the identifier manually if needed

## Considered Options

* Option 1 - Order-based, no field names in the identifier
* Option 2 - Field names in the identifier

## Decision Outcome

* Option 1 - Order-based, no field names in the identifier

## Pros and Cons of the Options

### Option 1 - Order-based, no field names in the identifier

We need to pick a delimiter that's allowed in the URL path and not allowed in resource names, meshes, namespaces, or zones. Good candidates are `:` and `_`.

There is an identifier format from Amazon called [ARN](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference-arns.html). We can adopt a similar approach:

```
kri:<mesh>:<zone>:<namespace>:<resource-type>:<resource-name>:<section-name>
```
For example:
```
kri:mesh-1:us-east-2:kuma-demo:meshservice:backend
kri:mesh-1:us-east-2:kuma-demo:meshservice:backend:http-port
kri:mesh-1:::meshtimeout:global-timeouts
```

Having a prefix like `kri` (Kuma Resource Identifier) is useful for two reasons:
* It visually clarifies the format for users, who can then search for the format description in the documentation.
* It acts as an implicit version. If we need to update the format, we can use a different prefix (e.g., `uri` or `ri`).

**Pros:**
- Shorter
- Resembles existing formats from Amazon

**Cons:**
- Hard to read when names are poorly chosen, e.g., `kri:default:default:default:meshservice:backend`

### Option 2 - Field names in the identifier

We need to pick two delimiters: one to separate keys from values and another to separate key-value pairs.

```yaml
kri:meshservice:mesh=mesh-1:zone:zone-1:namespace=kuma-demo:name=backend:section=http-port
```

**Pros:**
- Better handling of gaps; no need for `::` when a value is not defined

**Cons:**
- Longer
- The order of fields still matters if we want to compare identifiers with `==`
