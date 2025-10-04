# Add REST Endpoint for KRI-based Resource Lookup

* Status: Accepted

Technical Story: [https://github.com/kumahq/kuma/issues/13882](https://github.com/kumahq/kuma/issues/13882)

## Context and Problem Statement

We want to enable clients (GUI, SDKs, CLI) to fetch any resource using its **Kuma Resource Identifier (KRI)**.
Currently, clients construct REST paths using mesh and resource name (`/meshes/{mesh}/{resourceType}/{name}`).
Because of this, it is not possible to simply link between KRI resources: if you have a KRI in one place, you cannot directly navigate to that resource.
This complicates GUI/frontend development.

A dedicated endpoint for retrieving a resource by its KRI would:

* Provide a canonical API surface for resource lookups.
* Simplify frontend logic.
* Ensure KRIs become first-class citizens in the system.

Use cases:

* GUI fetching a resource by clicking a link derived from a KRI.
* CLI retrieving resources directly by KRI without constructing full paths.
* SDKs needing stable and consistent access patterns across resource types.

## Decision Drivers

* Ensure stable, predictable API for GUI and SDKs.
* Keep API design minimal, avoid unnecessary complexity.
* Maintain performance (avoid costly label lookups).
* Avoid breaking existing clients; additive change only.
* Align with OpenAPI/Typescript code generation while balancing developer ergonomics.

## Alternatives Considered

### Backend

#### Option A: Parse KRI and reuse existing store logic (not possible)

**Idea**: Backend parses KRI into `mesh` + `name` and calls existing `store.GetByKey` logic.

* Not possible because the name in KRI is derived using `GetDisplayName` which relies on the `kuma.io/display-name` label.

```go
return Identifier{
    ResourceType: resourceType,
    Mesh:         rm.GetMesh(),
    Zone:         rm.GetLabels()[mesh_proto.ZoneTag],
    Namespace:    rm.GetLabels()[mesh_proto.KubeNamespaceTag],
    Name:         core_model.GetDisplayName(rm),
}
```

[source](https://github.com/kumahq/kuma/blob/86e7375e52abbe933f3f9d2d99dc09f76e0cb1a8/pkg/core/kri/kri.go#L54-L60)

DisplayName is a standard label used to recognize policy names without namespace/hash suffixes.

```go
// prefer display name as it's more predictable
if labels := rm.GetLabels(); labels != nil && labels[mesh_proto.DisplayName] != "" {
    return labels[mesh_proto.DisplayName]
}
return rm.GetName()
```

[source](https://github.com/kumahq/kuma/blob/9e9ad8aadf73f240763c30174cfed7ea7ef416eb/pkg/core/resources/model/resource.go#L502-L510)

#### Option B: Introduce hash/index-based lookup

**Idea**: Backend would fetch resources and filter them by labels extending the label filtering logic [we have now](https://github.com/kumahq/kuma/blob/39cd2e670bbb56a0134fcdfce15c46cdf0a308eb/pkg/api-server/filters/filtering.go#L31). A mapping from KRI to resources would then be stored in a hash map to optimize performance.

**Pros**

* Potential performance gain.
* Avoid repeated expensive label lookups.

**Cons**

* Adds complexity and new indexing logic.

### Frontend & OpenAPI

#### Option A: Single endpoint `/_kri/{kri}` with runtime type guards (chosen)

**Idea**: Provide a single generic endpoint for all resources.
Clients rely on `type: <ResourceType>` inside the response and infer type at runtime using type guards. OpenAPI schema describes the endpoint generically with `oneOf`s.

**Pros**

* Minimal API surface.
* Pure KRI-driven design, consistent with intent of KRIs.
* Flexible for new resource types without adding new endpoints.
* Frontend can infer type using runtime type guards.

**Cons**

* Weaker compile-time safety compared to OpenAPI  typing.
* Requires additional runtime type guard implementation in frontend.
* Slightly more effort for GUI developers to validate resource shape

#### Option B: Typed endpoint with `shortName` segment (not chosen)

**Idea**: Use a separate endpoint per resource type including `shortName` in the path, e.g., `/_kri/{shortName}/{kri}`.
This would allow OpenAPI/Typescript SDKs to generate more precise types for each resource.

**Pros**

* Stronger typing in generated SDKs.
* Frontend can leverage OpenAPI type information directly.

**Cons**

* Increased API surface.
* Redundant endpoints for each resource type.
* More maintenance overhead as new resource types are added..

## Design

We introduce a new endpoint:

```yaml
/_kri/{kri}:
  get:
    operationId: getByKri
    summary: Returns resource by KRI
    tags: [ "KRI" ]
    parameters:
      - in: path
        name: kri
        schema:
          type: string
        required: true
        description: KRI of the resource
    responses:
      '200':
        description: The resource
        content:
          application/json:
            schema:
              oneOf: ... # TODO: add schema
      '400':
        $ref: "/specs/base/specs/common/error_schema.yaml#/components/responses/BadRequest"
      '404':
        $ref: "/specs/base/specs/common/error_schema.yaml#/components/responses/NotFound"
```

## Reliability implications

* Endpoint is additive.
* No changes to existing flows.
* Runtime type guard approach ensures frontend object mapping is correct.

## Implications for Kong Mesh

* None

## Decision

We choose "Option A: Single endpoint `_kri/{kri}`"
This keeps the API minimal.
It avoids redundant typed endpoints.
It lets the frontend enforce typing at runtime.
Backend parses the KRI and uses existing store lookup routines.
SectionName is stripped as needed.

This approach ensures KRIs are first-class in the system.
It avoids adding unnecessary complexity to the API surface.

## Notes

* Typed endpoints were considered but rejected due to added complexity and redundancy.

* Scope is limited to single-resource fetch.

* Frontend will prototype type guard approach to validate developer experience.
