package framework

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/testing"
	"github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/test/framework/versions"
)

var suiteFailed bool

func ShouldSkipCleanup() bool {
	suiteConfig, _ := ginkgo.GinkgoConfiguration()

	return (suiteFailed || ginkgo.CurrentSpecReport().Failed()) && suiteConfig.FailFast
}

func doIfNoSkipCleanup(fn func()) func() {
	return func() {
		if ShouldSkipCleanup() {
			return
		}

		fn()
	}
}

func AfterEachFailure(fn func()) bool {
	return ginkgo.AfterEach(func() {
		if !ginkgo.CurrentSpecReport().Failed() {
			return
		}
		fn()
	})
}

func E2EAfterEach(fn func()) bool {
	return ginkgo.AfterEach(doIfNoSkipCleanup(fn))
}

func E2EAfterAll(fn func()) bool {
	return ginkgo.AfterAll(doIfNoSkipCleanup(fn))
}

func E2EAfterSuite(fn func()) bool {
	return ginkgo.AfterSuite(doIfNoSkipCleanup(fn))
}

func E2ESynchronizedBeforeSuite(process1Body interface{}, allProcessBody interface{}, args ...interface{}) bool {
	ginkgo.AfterEach(func() {
		if ginkgo.CurrentSpecReport().Failed() {
			suiteFailed = true
		}
	})
	return ginkgo.SynchronizedBeforeSuite(process1Body, allProcessBody, args...)
}

func E2EBeforeSuite(fn func()) bool {
	ginkgo.AfterEach(func() {
		if ginkgo.CurrentSpecReport().Failed() {
			suiteFailed = true
		}
	})

	return ginkgo.BeforeSuite(func() {
		fn()
	})
}

func E2EDeferCleanup(args ...interface{}) {
	callback := reflect.ValueOf(args[0])
	if !(callback.Kind() == reflect.Func && callback.Type().NumOut() <= 1) {
		ginkgo.Fail(fmt.Sprintf(
			"first argument in E2EDeferCleanup must be a function and is %T instead",
			args[0],
		))
	}

	fn := func(args []interface{}) error {
		if ShouldSkipCleanup() {
			return nil
		}

		var callArgs []reflect.Value
		for _, arg := range args {
			callArgs = append(callArgs, reflect.ValueOf(arg))
		}

		out := callback.Call(callArgs)
		if len(out) > 0 && !out[len(out)-1].IsNil() {
			return out[len(out)-1].Interface().(error)
		}

		return nil
	}

	ginkgo.DeferCleanup(fn, args[1:])
}

func SupportedVersionEntries(t testing.TestingT) []ginkgo.TableEntry {
	var res []ginkgo.TableEntry
	for _, v := range versions.UpgradableVersionsFromBuild(Config.SupportedVersions()) {
		version, err := GetLatestPatch(t, v)
		if err != nil {
			ginkgo.Fail(fmt.Sprintf("failed to find the newest patch version, error: %v", err))
		}
		res = append(res, ginkgo.Entry(nil, version.Version))
	}
	return res
}

type HelmChart struct {
	Version string `json:"version"`
}

func GetLatestPatch(t testing.TestingT, version string) (*HelmChart, error) {
	out, err := helm.RunHelmCommandAndGetStdOutE(t, &helm.Options{}, "search", "repo", "kuma/kuma", "--version", version, "-o", "json")
	if err != nil {
		return nil, err
	}
	// Parse JSON into a slice of HelmChart structs
	var newestPatch []HelmChart
	err = json.Unmarshal([]byte(out), &newestPatch)
	if err != nil {
		return nil, err
	}
	if len(newestPatch) != 1 {
		return nil, fmt.Errorf("invalid number of helm patch versions, expected: 1, actual %d", len(newestPatch))
	}
	return &newestPatch[0], nil
}
