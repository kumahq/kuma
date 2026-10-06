# Creating resources without replacing existing objects

Kuma 3.0 adds create-only collection endpoints. Existing PUT clients continue to
work without changes.

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
paths also accept POST. Cross-mesh list endpoints are not creation endpoints.

A successful POST returns the existing JSON warnings response and a `Location`
header containing the resource's item path, including the configured API base
path. Duplicate creation returns the structured API error with status 409 and a
resource identity in `detail`. Uniqueness is enforced by the store's atomic
Create operation; a conflict never falls back to Update. Concurrent creators
cannot replace the winner.

## Client compatibility

Kuma 3.0's remote resource store sends POST for Create and PUT for Update.
`kumactl apply` retains its GET-then-create-or-update behavior: it updates a
resource it found, but reports a conflict if another writer creates the resource
after its lookup. Creation requires a Kuma 3.0 or newer control plane. Upgrade the
control plane before using the new client for creation. There is no automatic
POST-to-PUT fallback on an older server; use the matching older client when
operating an older control plane, with its existing upsert semantics.

## Terraform provider rollout

Provider generation is maintained outside this repository. For a schema with
create-only POST, map Terraform Create to POST and Update to PUT. Keep the PUT
200/201 responses in the API schema: other clients still use upsert. Map a
collection POST with a 201 response to `<Resource>#create`, and its item PUT to
`<Resource>#update`. Do not map read-only POST operations (405) to Create.

Remove the generated plan-time existence check only for provider resources whose
Create uses POST. Keep legacy generation behavior for consumers that still use
PUT, including Konnect versions without the new endpoint. Do not remove the
shared check globally. A released POST-based provider must require a server with
this API; a GET-before-PUT compatibility path cannot provide atomic creation.

A 409 must fail creation without adopting or overwriting the existing object.
Client guidance should explain the collision: import only if the user intends to
manage that object, or choose a different name. Default destroy-before-create
replacement and taint recreation can plan normally and DELETE before POST.
Same-name `create_before_destroy` necessarily conflicts while the original still
exists; import guidance alone is inappropriate for that case.
