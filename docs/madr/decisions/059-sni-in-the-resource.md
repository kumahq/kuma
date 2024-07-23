# SNI in the resource definition

* Status: accepted

Technical Story: 

## Context and Problem Statement

SNI is used for cross-zone communication to proxy the traffic through zone proxies.
Recently we decided on SNI format in MADR 053.

The SNI is constructed on the server and the client side. This approach has a couple of downsides
* We hit multiple edge cases of constructing it on the client side, because of resource hashing.
  This logic is not as simple as it should be.
* Any migration of this format would be hard to execute
* If we introduce a functionality to enter the mesh directly through zone ingress (assuming that cert is acquired) it would be hard for the client to construct this.

## Considered Options

MeshService:
* Option 1 - SNI in the resource definition
* Option 2 - Keep it implicit

MeshMultiZoneService:
* Option 1 - SNI in the resource definition and require ports from user
* Option 2 - SNI in the resource definition and compute ports on global cp
* Option 3 - Keep it implicit

## Decision Outcome

Chosen option: "Option 1 - SNI in the resource definition", because it solves the problems described in previous section.
For `MeshMultiZoneService` "Option 1 - SNI in the resource definition and require ports from user", because of benefits of being explicit.
If we find this to be to annoying, the right route is probably to autogenerate MeshMultiZoneService.

## Pros and Cons of the Options

### MeshService

#### Option 1 - SNI in the resource definition

SNI is unique for each port, therefore we need to define it for each port

```yaml
kind: MeshService
spec:
  ports:
    - port: 80
      targetPort: 8080
      snis: # new field
      - value: "ae10a8071b8a8eeb8.backend.8080.demo.ms"
```

The `snis` field will be filled out automatically by the original Zone CP and synced cross zone.
The field will be filled out by the same mechanism as Mesh Defaulter, which is webhook on Kubernetes and `ResourceManager` on Universal.
This way we can avoid a situation that we drop SNI by the resource update.

**Migration**
SNIs is a list, not a field, so we can perform a migration of SNI. If we ever change the format we can provide multiple snis in the list.
The client should use the first SNI on the list. This way we can go from
```yaml
snis:
- value: old
```
to
```yaml
snis:
- value: new
- value: old
```
to
```yaml
snis:
- value: new
```

Item in an array is an object with just `value` field.
It's an object, so we can potentially expand this field if we need it to add fields like `type` or `version` etc.

Advantages:
* Slightly better debuggability because SNI is in MeshService object
* Easy migration
* Easy usage from the client side

Disadvantages:
* Putting more stuff in the object

#### Option 2 - Keep it implicit

Advantages:
* No new fields in objects

Disadvantages:
* Keep the existing tradeoffs

### MeshMultiZoneService

SNI should always be filled out by the source of truth of the resource.
In case of `MeshMultiZoneService`, this is global CP.
However, the problem is that `MeshMultiZoneService` does not require defining ports directly.
They are derived from the `MeshServices` selected by `MeshMultiZoneService`.
Global CP does not have list of ports, so it cannot update them with SNI.

```yaml
kind: MeshMultiZoneService
name: test-server
spec:
  selector:
    meshService:
      matchLabels:
        kuma.io/display-name: test-server
        k8s.kuma.io/namespace: mzmsconnectivity
status: # computed on the zone
  vips:
    - ip: 241.0.0.1
  meshServices:
    - name: test-server.mzmsconnectivity
    - name: test-server-234jhg34yt34.kuma-system
  ports:
    - port: 80
      targetPort: 8080
```

#### Option 1 - SNI in the resource definition and require ports from user

We can require ports from user, so it's

```yaml
spec:
  selector:
    meshService:
      matchLabels:
        kuma.io/display-name: test-server
        k8s.kuma.io/namespace: mzmsconnectivity
  ports:
    - port: 80
      targetPort: 8080
      snis:  # new field
      - value: "ae10a8071b8a8eeb8.backend.80.demo.mzms"
```

SNI is then applied the same way as MeshService (via webhook/ResourceManager) and is synced down to zones.

Advantage:
* We see ports in the GUI
* We avoid mistakes when ports are different between services. What to do when
  * One service has port 80 and other has port 8080
  * One service has appProtocol http on port 80 and other has tcp on port 80

Disadvantage:
* It might be annoying for the user to repeat the information that is already defined on MeshServices

#### Option 2 - SNI in the resource definition and compute ports on global cp

Global CP can have a similar component to Zone CP which is to go over all `MeshServices` and fill this data for the user.
Computed ports are stored in `spec.ports` so it's synced down to all zones

Advantage:
* We see ports in the GUI
* Asking less things from the user

Disadvantage:
* Putting more compute operations on Global CP.
  However, we may need to have a component to compute `status` on global anyway to give better GUI visibility of aggregated services. 

#### Option 3 - Keep it implicit

Advantages:
* No new fields in objects

Disadvantages:
* Keep the existing tradeoffs
* Don't see ports of `MeshMultiZoneService` in the GUI.
