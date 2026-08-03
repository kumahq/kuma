package v1alpha1

import (
	"hash/fnv"
	"net"
	"strconv"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

// Hash returns a content-based hash of the MeshZoneAddress. The address is
// hashed on top of the meta because the mesh context stores it after DNS
// resolution: a load balancer hostname keeps the same resourceVersion while
// resolving to a new IP, and meta-only hashing would keep serving the stale
// endpoint.
func (m *MeshZoneAddressResource) Hash() []byte {
	hasher := fnv.New128a()
	_, _ = hasher.Write(core_model.HashMeta(m))
	if m.Spec != nil {
		// address and port are joined instead of written back to back so that
		// distinct pairs can't serialize to the same bytes, e.g. ("10.0.0.1", 23)
		// and ("10.0.0.12", 3)
		_, _ = hasher.Write([]byte(net.JoinHostPort(m.Spec.Address, strconv.Itoa(int(m.Spec.Port)))))
	}
	return hasher.Sum(nil)
}
