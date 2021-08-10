package kubernetes_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/trafficpermission/kubernetes"
)

var _ = Describe("Traffic Permission on Kubernetes", kubernetes.TrafficPermission)
