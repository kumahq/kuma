package context

import (
	"encoding/binary"
	"hash"
	"hash/fnv"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/dns/vips"
)

type xdsHasher interface {
	XDSHash() []byte
}

func resourceListXDSHash(rl core_model.ResourceList) []byte {
	hasher := fnv.New128()
	for _, entity := range rl.GetItems() {
		_, _ = hasher.Write(resourceXDSHash(entity))
	}
	return hasher.Sum(nil)
}

func resourceXDSHash(resource core_model.Resource) []byte {
	if hasher, ok := resource.(xdsHasher); ok {
		return hasher.XDSHash()
	}
	return core_model.Hash(resource)
}

func virtualOutboundViewHash(view *vips.VirtualOutboundMeshView) []byte {
	hasher := fnv.New128a()
	for _, entry := range view.HostnameEntries() {
		writeHashString(hasher, entry.String())
		virtualOutbound := view.Get(entry)
		writeHashString(hasher, virtualOutbound.Address)
		for _, outbound := range virtualOutbound.Outbounds {
			writeHashString(hasher, outbound.String())
		}
	}
	return hasher.Sum(nil)
}

func writeHashString(hasher hash.Hash, value string) {
	var lenBuf [8]byte
	binary.BigEndian.PutUint64(lenBuf[:], uint64(len(value)))
	_, _ = hasher.Write(lenBuf[:])
	_, _ = hasher.Write([]byte(value))
}
