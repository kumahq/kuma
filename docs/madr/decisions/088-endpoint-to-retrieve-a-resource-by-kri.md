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

#### Option A: Parse KRI and reuse existing store logic

**Idea**: Backend parses KRI into `mesh` and `name`, computes the name for the store and uses existing store logic to fetch the resource.

**Pros**

* Reuses existing store logic.
* Minimal perf overhead (no range scans).

**Cons**

* Adds complexity and new indexing logic.

#### Option B: Introduce hash/index-based lookup (not chosen)

**Idea**: Backend would fetch resources and filter them by labels extending the label filtering logic [we have now](https://github.com/kumahq/kuma/blob/39cd2e670bbb56a0134fcdfce15c46cdf0a308eb/pkg/api-server/filters/filtering.go#L31). A mapping from KRI to resources would then be stored in a hash map to optimize performance.

**Pros**

* None.

**Cons**

* Adds complexity and new indexing logic.

### Frontend & OpenAPI

#### Option A: Single endpoint `/_kri/{kri}` only with manual casting and/or runtime type guards (chosen)

**Idea**: Provide a single generic endpoint for all resources.

At runtime, clients can rely on the requested KRI `shortName` or the response `type: <ResourceType>` to understand the shape of the response.
At build-time/compile-type, clients will receive a very wide type generated via OpenAPIs `oneOf` (i.e. a large union) to understand the shape of the runtime response. Clients _may_ decide to manually implement additional static approaches to provide their own guarantees as to the shape/types of the response via casting and/or type guards.

Following usage we _may_ decide at a future date to add a further set of typed endpoints. If so, a separate MADR will be written to document that decision.

**Pros**

* Minimal API surface for consumers.
* Pure KRI-driven design, consistent with intent of KRIs.
* Flexible for new resource types without adding new endpoints.

**Cons**

* Weaker compile-time safety compared to OpenAPI typing.
* Requires additional effort in clients of runtime type guard/casting implementation to validate the shape of the requested resource.
* Moves API specification responsibilities from OpenAPI to Typescript against existing boundaries.

## Design

For the generic `/_kri/{kri}` endpoint we would have:

```yaml
openapi: 3.0.0
info:
  title: KRI API
  version: 1.0.0
paths:
  /_kri/{kri}:
    get:
      operationId: getByKri
      summary: Returns a resource by KRI
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
                oneOf:
                  - $ref: '#/components/schemas/MeshAccessLog'
                  - $ref: '#/components/schemas/MeshTrafficPermission'
                  - $ref: '#/components/schemas/MeshTrace'
              discriminator:
                propertyName: type
              mapping: # maybe not needed because name of type == name of schema
                MeshTrafficPermission: '#/components/schemas/MeshTrafficPermission'
                  # ... all other resource types
        '400': # only invalid KRI triggers a 400
          $ref: "/specs/base/specs/common/error_schema.yaml#/components/responses/BadRequest"
        '404': # this is triggered if the KRI is valid but the resource is not found
          $ref: "/specs/base/specs/common/error_schema.yaml#/components/responses/NotFound"
```

### Resource fetching

#### Name computation

To compute the name we implement a function that takes a KRI and returns a ResourceKey:
1. Figure out if resource originated on this CP or not. This can be done via a function [IsLocallyOriginated](https://github.com/kumahq/kuma/blob/9e9ad8aadf73f240763c30174cfed7ea7ef416eb/pkg/core/resources/model/resource.go#L475-L486).
2. If it's locally originated, adjust the name for the k8s store because on k8s we have CoreName (which is ${name}.${namespace}). If it's universal that's not the case.
3. If it's not locally originated, you need to compute the hash from the data from KRI using [HashSuffixMapper](https://github.com/kumahq/kuma/blob/c989d3d842850aa468248e76298b9729af217a3b/pkg/kds/context/context.go#L232).

#### Endpoint handling

We will add the new endpoint next to the existing endpoints in [addFindEndpoint](https://github.com/kumahq/kuma/blob/4004a7231090fb2786d5ede41b3d95b73188d745/pkg/api-server/service_insight_endpoints.go#L28-L32).

## Reliability implications

* Endpoint is additive.
* No changes to existing flows.

## Implications for Kong Mesh

* None

## Decision

We choose "Option A: Single endpoint `/_kri/{kri}` only with manual casting and/or runtime type guards", along with a generic `/_kri/:kri` endpoint.
Backend parses the KRI and uses existing store lookup routines.
SectionName is stripped as needed.

This approach ensures KRIs are first-class in the system.
It avoids adding unnecessary complexity to the API surface.
