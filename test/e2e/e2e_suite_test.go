package e2e_test

import (
	"testing"
	"time"

	"github.com/Kong/kuma/test/framework"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/Kong/kuma/pkg/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2E(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		RegisterFailHandler(Fail)
		RunSpecs(t, "E2E Suite")
	} else {
		t.SkipNow()
	}
}

var _ = BeforeSuite(func() {
	core.SetLogger = func(l logr.Logger) {}
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))
})

const (
	defaultRetries = 60
	defaultTimeout = 3 * time.Second
)
