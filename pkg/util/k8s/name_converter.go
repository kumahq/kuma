package k8s

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

func CoreNameToK8sName(coreName string) (string, string, error) {
	parts := strings.Split(coreName, ".")
	if len(parts) < 2 {
		return "", "", errors.New(`name must include namespace after the dot, ex. "name.namespace"`)
	}
	nameParts := []string{}
	for i := 0; i < len(parts)-1; i++ {
		nameParts = append(nameParts, parts[i])
	}
	// namespace cannot contain "." therefore it's always the last part
	return strings.Join(nameParts, "."), parts[len(parts)-1], nil
}

func K8sNamespacedNameToCoreName(name, namespace string) string {
	return fmt.Sprintf("%s.%s", name, namespace)
}
