package util_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/kds/util"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("TrimSuffixFromName", func() {
	type testCase struct {
		name   string
		suffix string
	}

	name := func(given testCase) string {
		return fmt.Sprintf("%s.%s", given.name, given.suffix)
	}

	DescribeTable("should remove provided suffix from the name of "+
		"the provided resource",
		func(given testCase) {
			// given
			meta := &test_model.ResourceMeta{Name: name(given)}
			resource := &test_model.Resource{Meta: meta}

			// when
			util.TrimSuffixFromName(resource, given.suffix)

			// then
			Expect(resource.GetMeta().GetName()).To(Equal(given.name))
		},
		// entry description generator
		func(given testCase) string {
			return fmt.Sprintf("name: %q, suffix: %q", name(given), given.suffix)
		},
		Entry(nil, testCase{name: "foo", suffix: "bar"}),
		Entry(nil, testCase{name: "bar", suffix: "baz"}),
		Entry(nil, testCase{name: "baz", suffix: "kuma-system"}),
		Entry(nil, testCase{name: "faz", suffix: "daz.kuma-system"}),
	)
})
