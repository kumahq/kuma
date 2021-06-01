package k8s_api_bypass_test

import (
	"testing"

	"github.com/kumahq/kuma/test/framework"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kumahq/kuma/pkg/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2EExternalServices(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		RegisterFailHandler(Fail)
		RunSpecs(t, "Kubernetes API Bypass")
	} else {
		t.SkipNow()
	}
}

var _ = BeforeSuite(func() {
	core.SetLogger = func(l logr.Logger) {}
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))
})
