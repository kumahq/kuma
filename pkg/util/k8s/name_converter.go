package k8s

import (
	"fmt"
	"hash"
	"hash/fnv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/rand"
)

func CoreNameToK8sName(coreName string) (string, string, error) {
	idx := strings.LastIndex(coreName, ".")
	if idx == -1 {
		return "", "", errors.Errorf(`name %q must include namespace after the dot, ex. "name.namespace"`, coreName)
	}
	// namespace cannot contain "." therefore it's always the last part
	namespace := coreName[idx+1:]
	if namespace == "" {
		return "", "", errors.New("namespace must be non-empty")
	}
	return coreName[:idx], namespace, nil
}

func K8sNamespacedNameToCoreName(name, namespace string) string {
	return fmt.Sprintf("%s.%s", name, namespace)
}

func NewHasher() hash.Hash32 {
	return fnv.New32a()
}

// HashToString calculates a hash the same way Pod template hashes are computed
func HashToString(h hash.Hash32) string {
	return rand.SafeEncodeString(fmt.Sprint(h.Sum32()))
}

// MaxHashStringLength is the max length of a string returned by HashToString
const MaxHashStringLength = 10

// EnsureMaxLength truncates the string if it's too long
func EnsureMaxLength(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[0:length]
}
