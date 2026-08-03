package v1alpha1

import (
	"hash/fnv"
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
		_, _ = hasher.Write([]byte(m.Spec.Address))
		_, _ = hasher.Write([]byte(strconv.Itoa(int(m.Spec.Port))))
	}
	return hasher.Sum(nil)
}
