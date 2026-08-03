# Computed API Endpoints

* Status: accepted

## Context and Problem Statement

There are two types of API endpoints in Kuma:

- Resource endpoints - These are endpoints that lead to direct reads to the database. They map to Kuma resources. For example: `GET /mesh/{mesh}/dataplanes/{dataplane}`.
- Computed endpoints - These are endpoints that provide computed data usually using existing resources. For example: `GET /mesh/{mesh}/dataplanes/{dataplane}/_overview` 
  or `GET /mesh/{mesh}/dataplanes/{dataplane}/_metrics`.

Computed endpoints represent runtime and composed state, this state is not stored in the database and there's no way to update it from the API.

This MADR provides a convention for computed endpoints. The goal is to provide a consistent way to define and document computed endpoints.
This doesn't prevent non computed endpoints to include part of the response as computed (examples: warnings or status fields)

## Design

Computed endpoints should be always prefixed with `_`, for example: `GET /mesh/{mesh}/dataplanes/{dataplane}/_overview`.

This has the following benefits:

- It is easy to identify computed endpoints.
- Names of resources and resource type names can never start with `_`, so there is no risk of collision between a resource and a computed endpoint. For example: `GET /mesh/{mesh}/dataplanes/{dataplane}` and `GET /mesh/{mesh}/dataplanes/_overview` are two different endpoints there's never a risk of having a dataplane names `_overview`.
- endpoints without `_` prefix are always resource endpoints. This makes it easier to understand the API and reduces the risk of confusion, it also makes generation of clients easier.

## Implications for Kong Mesh

None

## Decision

The decision is to use `_` as a prefix for computed endpoints. This is a convention that should be followed in all future development.

For existing endpoints, we should provide alternative endpoints and progressively deprecate old endpoints.
All new endpoints should follow this convention and be documented in openAPI spec.

Some examples of endpoints that should be prefixed with `_`:

- xds, clusters, stats: https://github.com/kumahq/kuma/issues/11620
- `/global-insights`
- `/config`
- `/who-am-i`

## Alternatives Considered

- No prefix - This would make it harder to identify computed endpoints and could lead to confusion. It could also lead to collision between resource and computed endpoints.
- Using request param (`?computed=true`) - the schema of computed endpoints is different than their resource, this would make the API very complex, especially as this kind of response is hard to define in openAPI.
- Using `Accept` header - accept headers are not meant to change the actual content of the response (see [RFC7231](https://datatracker.ietf.org/doc/html/rfc7231#section-5.3.2)) it's only meant to define the format of the response.