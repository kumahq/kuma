package firewalld

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/test/matchers"
)

type testCase struct {
	inputFile  string
	goldenFile string
}

var _ = Describe("firewalld", func() {
	DescribeTable("check xml generation",
		func(given testCase) {
			rules, err := os.ReadFile(path.Join("testdata", given.inputFile))
			Expect(err).To(Succeed())

			Expect(NewIptablesTranslator().WithDryRun(true).StoreRules(string(rules))).
				To(MatchGoldenXML("testdata", given.goldenFile))
		},
		Entry("should generate xml", testCase{
			inputFile:  "full_direct.input.txt",
			goldenFile: "full_direct.golden.xml",
		}),
		Entry("should generate xml for verbose rules", testCase{
			inputFile:  "full_direct_verbose.input.txt",
			goldenFile: "full_direct_verbose.golden.xml",
		}),
		Entry("should generate xml without duplicates", testCase{
			inputFile:  "no_duplicates_direct.input.txt",
			goldenFile: "no_duplicates_direct.golden.xml",
		}),
	)
})
