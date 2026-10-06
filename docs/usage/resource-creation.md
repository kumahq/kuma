# Creating resources without replacing existing objects

Kuma 3.0 adds create-only POST endpoints on writable resource collections.
Existing PUT clients, including kumactl, keep their existing behavior.

| Intent | Mesh-scoped request | Global request | Success | Existing identity |
| --- | --- | --- | --- | --- |
| Create only | `POST /meshes/{mesh}/{type}` | `POST /{type}` | 201 | 409; existing resource unchanged |
| Create or replace | `PUT /meshes/{mesh}/{type}/{name}` | `PUT /{type}/{name}` | 201 on creation, 200 on replacement | Replaced |

For example, create a Mesh with:

```sh
curl -i -X POST http://localhost:5681/meshes \
  -H 'Content-Type: application/json' \
  -d '{"type":"Mesh","name":"example"}'
```

POST takes `name` from the request body. The body must match the collection's
resource type and mesh. The same authorization, validation, label computation,
ownership, and read-only restrictions as PUT apply. Writable alternate collection
paths also accept POST. Mesh-scoped resources must be created under their mesh.
If the API uses a base path, prefix the request path with it.

A successful POST returns the existing JSON warnings response. Duplicate creation
returns the structured API error with status 409 and a resource identity in
`detail`. Uniqueness is enforced by the store's atomic Create operation; a
conflict never falls back to Update. Concurrent creators cannot replace the
winner.

Clients adopting POST creation require Kuma 3.0 or later. Existing clients that
use PUT retain create-or-replace semantics. A preliminary GET followed by PUT
cannot guarantee create-only behavior, and falling back from a failed POST to PUT
can overwrite an existing resource.
