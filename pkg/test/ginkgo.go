package test

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kumahq/kuma/pkg/core"
)

// RunSpecs wraps ginkgo+gomega test suite initialization.
func RunSpecs(t *testing.T, description string) {
	// Make resetting the core logger a no-op so that internal
	// code doesn't interfere with testing.
	core.SetLogger = func(l logr.Logger) {}

	// Log to the Ginkgo writer. This makes Ginkgo emit logs on
	// test failure.
	log.SetLogger(zap.LoggerTo(ginkgo.GinkgoWriter, true))

	gomega.RegisterFailHandlerWithT(t, ginkgo.Fail)
	ginkgo.RunSpecs(t, description)
}
