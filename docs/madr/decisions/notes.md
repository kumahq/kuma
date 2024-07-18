Can we afford always routing MES through egress or we’re cutting some important usecase by not allowing accessing MES directly from the sidecar
---

Running on the sidecar:
- MeshExternalService allows us to gradually migrate existing K8S services into the mesh,
  it would be a bit awkward, but it's technically valid
- You still can access external services using MeshPassthrough
- There was a load balancing issue with TCP traffic using zone egress when we make a connection and change the load
  balancing algorithm we had to wait for the connection to be re-created

Running traffic through egress:
- a bit of operational cost
  - it's a bit annoying to set up on universal
- we've got a bit of a gap in other functionalities on egress
- can we hit scalability issues with egress?
- you need mTLS enabled (but only for services) to run egress, but we know if some traffic is MES, so we can relax that 


Should we make egress required for MeshServices if we make it required for MeshExternalServices?

Conclusion from the meeting
**It seems like we're not cutting any serious use case**
