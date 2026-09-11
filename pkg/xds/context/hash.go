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

// policyMatchingHasher is implemented by resources with fields that change xDS but never which
// policies match, so the policy matching cache is not invalidated when only those change.
type policyMatchingHasher interface {
	PolicyMatchingHash() []byte
}

func policyMatchingListHash(th typeHash) []byte {
	items := th.list.GetItems()
	if len(items) == 0 {
		return th.hash
	}
	if _, ok := items[0].(policyMatchingHasher); !ok {
		return th.hash
	}
	hasher := fnv.New128()
	for _, item := range items {
		_, _ = hasher.Write(item.(policyMatchingHasher).PolicyMatchingHash())
	}
	return hasher.Sum(nil)
}

func resourceXDSHash(resource core_model.Resource) []byte {
	if hasher, ok := resource.(xdsHasher); ok {
		return hasher.XDSHash()
	}
	return core_model.Hash(resource)
}
