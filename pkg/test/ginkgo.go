package test

import (
	"os"
	"path"
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
	"github.com/kumahq/kuma/pkg/core/plugins"
	core_apis "github.com/kumahq/kuma/pkg/core/resources/apis"
	"github.com/kumahq/kuma/pkg/plugins/policies"
)

// RunSpecs wraps ginkgo+gomega test suite initialization.
func RunSpecs(t *testing.T, description string) {
	plugins.InitAll(core_apis.NameToModule)
	plugins.InitAll(policies.NameToModule)
	format.TruncatedDiff = false
	if strings.HasPrefix(description, "E2E") {
		panic("Use RunE2ESpecs for e2e tests!")
	}
	runSpecs(t, description)
}

func RunE2ESpecs(t *testing.T, description string) {
	plugins.InitAll(core_apis.NameToModule)
	plugins.InitAll(policies.NameToModule)
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

// EntriesForFolder returns all files in the folder as gingko table entries for files *.input.yaml this makes it easier to add test by only adding input and golden files
// if you prefix the file with a `F` we'll focus this specific test
func EntriesForFolder(folder string, pathPrefix ...string) []ginkgo.TableEntry {
	ginkgo.GinkgoHelper()
	var entries []ginkgo.TableEntry
	testDir := path.Join(append(pathPrefix, "testdata", folder)...)
	files, err := os.ReadDir(testDir)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".input.yaml") {
			input := path.Join(testDir, f.Name())
			switch {
			case strings.HasPrefix(f.Name(), "F"):
				entries = append(entries, ginkgo.FEntry(input, input))
			case strings.HasPrefix(f.Name(), "P"):
				entries = append(entries, ginkgo.PEntry(input, input))
			default:
				entries = append(entries, ginkgo.Entry(input, input))
			}
		}
	}
	return entries
}
