package test

import (
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kumahq/kuma/pkg/core"
)

// RunSpecs wraps ginkgo+gomega test suite initialization.
func RunSpecs(t *testing.T, description string) {
	if strings.HasPrefix(description, "E2E") {
		panic("Use RunE2ESpecs for e2e tests!")
	}
	runSpecs(t, description)
}

func RunE2ESpecs(t *testing.T, description string) {
	gomega.SetDefaultConsistentlyDuration(time.Second * 5)
	gomega.SetDefaultConsistentlyPollingInterval(time.Millisecond * 200)
	gomega.SetDefaultEventuallyPollingInterval(time.Millisecond * 500)
	gomega.SetDefaultEventuallyTimeout(time.Second * 30)
	// Set MaxLength to larger value than default 4000, so we can print objects full like Pod on test failure
	format.MaxLength = 100000
	runSpecs(t, description)
}

func runSpecs(t *testing.T, description string) {
	// Make resetting the core logger a no-op so that internal
	// code doesn't interfere with testing.
	core.SetLogger = func(l logr.Logger) {}

	// Log to the Ginkgo writer. This makes Ginkgo emit logs on
	// test failure.
	log.SetLogger(zap.New(
		zap.UseDevMode(true),
		zap.WriteTo(ginkgo.GinkgoWriter),
	))

	gomega.RegisterFailHandler(ginkgo.Fail)

	ginkgo.RunSpecs(t, description)
}
