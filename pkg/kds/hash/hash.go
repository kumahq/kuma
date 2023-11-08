package hash

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"

	"k8s.io/apimachinery/pkg/util/rand"
	k8s_strings "k8s.io/utils/strings"
)

// SyncedNameInZone returns the resource's name after syncing from Global to Zone.
// It creates a new name by adding a hash suffix constructed from the 'mesh' and
// the original 'name'.
func SyncedNameInZone(mesh, name string) string {
	return addSuffix(name, hash(mesh, name))
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
