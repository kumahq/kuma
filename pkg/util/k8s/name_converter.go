package k8s

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
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
