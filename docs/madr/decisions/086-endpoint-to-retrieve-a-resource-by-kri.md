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

#### Generic endpoint

There are multiple options that we are considering mainly around issues with automatic type generation for frontend clients etc. Whichever one of those we choose we would recommend always having an additional generic endpoint specified in the OpenAPI schema. This allows people to discover and use the API simply by using the form `/_kri/:kri` if they wish, but they get no types. Moreover this approach is necessary in order to retrieve user-defined policies which we cannot statically define at build time.

The type associated with endpoint can just use a very generic type containing fields common to all KRI based resources, such as `name`, `type` and preferably a `kri` field itself.

#### Option A: Single endpoint with multiple statically defined variants

**Idea**: Provide a single endpoint rooted under `/_kri` but provide statically defined "variants" for all resources based on the KRI shortname.

- `/_kri/kri_msvc_{mesh}_{zone}_{namespace}_{name}_{sectionName}`
- `/_kri/kri_zi_{mesh}_{zone}_{namespace}_{name}_{sectionName}`

Note: this option differs to Option B in that we are statically specifying _each resource type_ in OpenAPI instead of relying on a single generic specification (`/_kri/{kri}`) for very differently shaped responses.
The key difference here is that we are using `_` separated segments in the KRI instead of `/` separated segments in the URL.

##### Optional parameters

Some of the parameters in the KRI are optional (like `zone` for universal resources, or `mesh` for `ZoneIngress`).
Usually parameters take out the whole segment, e.g. `/meshes/{mesh}`, but here we have a parameter in the middle of a segment.
This is a gray area in the OpenAPI spec because path parameters require `required: true` to be set and it is unclear whether a required parameter can accept an empty string.
OpenAPI currently has a [`allowEmptyValues` parameter](https://spec.openapis.org/oas/v3.2.0.html#common-fixed-fields), but this is only allowed on query parameters and is deprecated, so we should not depend on this.
We could handle this by defining multiple endpoints for each combination of optional parameters, but that would lead to an explosion of paths.
We could also not care about this and just let the parameter be empty and handle it on our side, but that could lead to:
- OpenAPI spec technically being invalid.
- OpenAPI spec being updated in the future to handle this case some other way.
- Some generated tools not handling this case well (e.g. Swagger UI doesn't allow this).


**Pros**

* Minimal API surface for consumers.
* Pure KRI-driven design, consistent with intent of KRIs.
* Stronger typing in generated SDKs.
* Frontend can leverage OpenAPI type information directly.
* Uses existing standards and application patterns.
* Maintains existing engineering boundaries
* Static escape hatch for dynamic retrieval (in the case of policies), exchanging dynamism for a less narrow type.

**Cons**

* Some of the parameters can be empty which is problematic.
* More endpoint definitions as new resource types are added (mostly offset by the fact that these specifications are automatically generated)

#### Option B: Single endpoint with multiple statically defined variants with a single required argument

**Idea**: Provide a single endpoint rooted under `/_kri` but provide statically defined "variants" for all resources based on the KRI shortname. Differently to option A we only provide a single required parameter which is the "remaining" part of the KRI after the shortName.

- `/_kri/kri_msvc_{partialKRI}`: partialKRI would equal `my-mesh_my-zone_name_theSectionName`
- `/_kri/kri_zi_{partialKRI}`: partialKRI would equal `my-mesh_my-zone_name_theSectionName`

Note: This is is essentially the same approach as option A, but avoids the required path parameter problem because you always have to provide the remaining part of the KRI


**Pros**

The pros are the same as Option A's but additionally:

* Avoids problems of Option A due to OpenAPI required path parameters. i.e. technically this is valid OpenAPI.

**Cons**

The cons are the same as Option A's but additionally:

* More awkward to use in Typescript, providing "the rest of the KRI" is more awkward and not as safe nor user-friendly as providing individual KRI parameters.

#### Option C: Single endpoint `/_kri/{kri}` only with runtime type guards

**Idea**: Provide a single generic endpoint for all resources.
Clients rely on `type: <ResourceType>` inside the response and manually specify type at runtime using type guards and/or casting. OpenAPI schema describes the endpoint generically with `oneOf`s.

**Pros**

* Minimal API surface for consumers.
* Pure KRI-driven design, consistent with intent of KRIs.
* Flexible for new resource types without adding new endpoints.

**Cons**

* Weaker compile-time safety compared to OpenAPI typing.
* Requires additional effort of runtime type guard/casting implementation to validate the shape of the requested resource.
* Moves API specification responsibilities from OpenAPI to Typescript against existing boundaries.

#### Option D: Typed endpoint with `shortName` segment (chosen)

**Idea**: Use a separate endpoint per resource type including `shortName` in the path, e.g., `/_kri/{shortName}/{kri}`.
This would allow OpenAPI/Typescript SDKs to generate more precise types for each resource.
This can also be hidden behind `x-internal` so it's not published as a public spec.

**Pros**

* Stronger typing in generated SDKs.
* Frontend can leverage OpenAPI type information directly.

**Cons**

* Increased API surface.
* Redundant endpoints for each resource type.
* More maintenance overhead as new resource types are added.

## Design

### OpenAPI generation

We introduce new automatically generated endpoints for each new resource (note: gotemplate variables begin with `.`, OpenAPI params don't).
These endpoints will be next to existing endpoints in [endpoints.yaml](https://github.com/kumahq/kuma/blob/46b6807f56c08c79f34db20cfc452b8f5203bf90/tools/openapi/templates/endpoints.yaml#L9).

```yaml
openapi: 3.0.0
info:
  title: KRI API
  version: 1.0.0
paths:
  /_kri/kri_mtp_{mesh}_{zone}_{namespace}_{name}_{sectionName}:
    get:
      operationId: getMtpByKri
      summary: Returns a MTP resource by KRI
      tags: [ "KRI" ]
      parameters:
        - in: path
          name: mesh
          schema:
            type: string
          required: true
          description: mesh of the resource
        - in: path
          name: zone
          schema:
            type: string
          required: true
          description: zone of the resource
        - in: path
          name: namespace
          schema:
            type: string
          required: true
          description: namespace of the resource
        - in: path
          name: name
          schema:
            type: string
          required: true
          description: name of the resource
        - in: path
          name: sectionName
          schema:
            type: string
          required: true
          description: section name of the resource
      responses:
        '200':
          description: The resource
          content:
            application/json:
              schema: 
                type: string # ... the rest of the schema of the specific resource type
        '400':
          $ref: "/specs/base/specs/common/error_schema.yaml#/components/responses/BadRequest"
        '404':
          $ref: "/specs/base/specs/common/error_schema.yaml#/components/responses/NotFound"
```

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
                  # ... all other resource types
        '400':
          $ref: "/specs/base/specs/common/error_schema.yaml#/components/responses/BadRequest"
        '404':
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
* Buildtime static specification approach ensures frontend object mapping is correct.

## Implications for Kong Mesh

* None

## Decision

We choose "Option D: Typed endpoint with `shortName` segment", along with a generic `/_kri/:kri` endpoint.
It provides automatic build-time types for frontend usage.
Backend parses the KRI and uses existing store lookup routines.
SectionName is stripped as needed.

This approach ensures KRIs are first-class in the system.
It avoids adding unnecessary complexity to the API surface.
It avoids technical gray areas in OpenAPI specifications.
