package test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kumahq/kuma/pkg/core"
	kuma_log "github.com/kumahq/kuma/pkg/log"
)

// RunSpecs wraps ginkgo+gomega test suite initialization.
func RunSpecs(t *testing.T, description string) {
	setLogger := core.SetLogger
	logger := kuma_log.NewLoggerTo(ginkgo.GinkgoWriter, kuma_log.DebugLevel)

	// Log to the Ginkgo writer. This makes Ginkgo emit logs on
	// test failure.
	log.SetLogger(logger)

	// Make resetting the core logger a no-op so that internal
	// code doesn't interfere with testing.
	core.SetLogger = func(l logr.Logger) {}
	defer func() {
		core.SetLogger = setLogger
	}()

	gomega.RegisterFailHandlerWithT(t, ginkgo.Fail)

	resultsDir, ok := os.LookupEnv("GINKGO_XUNIT_RESULTS_DIR")
	if !ok {
		ginkgo.RunSpecs(t, description)
		return
	}

	// Silence deprecation warning for using custom reporters. Ginkgo V2
	// will apparently have a command-line flag to do xunit reporting.
	_ = os.Setenv("ACK_GINKGO_DEPRECATIONS", "1.16.4")

	filename := fmt.Sprintf("%s.xml", strings.ReplaceAll(strings.ToLower(description), " ", "-"))
	ginkgo.RunSpecsWithDefaultAndCustomReporters(t, description, []ginkgo.Reporter{
		reporters.NewJUnitReporter(path.Join(resultsDir, filename)),
	})
}
