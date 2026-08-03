package context

import (
	"hash/fnv"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
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
