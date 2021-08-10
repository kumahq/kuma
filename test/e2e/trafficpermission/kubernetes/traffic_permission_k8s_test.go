package kubernetes_test

import (
	"github.com/kumahq/kuma/test/e2e/trafficpermission/kubernetes"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Traffic Permission on Kubernetes", kubernetes.TrafficPermission)
