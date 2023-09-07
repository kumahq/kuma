package hash

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"

	"k8s.io/apimachinery/pkg/util/rand"
	k8s_strings "k8s.io/utils/strings"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	util_k8s "github.com/kumahq/kuma/pkg/util/k8s"
)

func ZoneName(rm core_model.ResourceMeta) string {
	if n, ns, err := util_k8s.CoreNameToK8sName(rm.GetName()); err == nil {
		return util_k8s.K8sNamespacedNameToCoreName(addSuffix(n, hash(rm.GetMesh(), n)), ns)
	} else {
		return addSuffix(rm.GetName(), hash(rm.GetMesh(), rm.GetName()))
	}
}

func addSuffix(name, hash string) string {
	const hashLength = 1 + 16 // 1 dash plus 8-byte hash that is encoded with hex
	const k8sNameLengthLimit = 253
	shortenName := k8s_strings.ShortenString(name, k8sNameLengthLimit-hashLength)
	return fmt.Sprintf("%s-%s", shortenName, hash)
}

func hash(ss ...string) string {
	hasher := fnv.New64a()
	for _, s := range ss {
		_, _ = hasher.Write([]byte(s))
	}
	b := []byte{}
	b = hasher.Sum(b)

	return rand.SafeEncodeString(hex.EncodeToString(b))
}
