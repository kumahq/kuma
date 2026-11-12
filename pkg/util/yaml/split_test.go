package yaml_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_yaml "github.com/kumahq/kuma/v3/pkg/util/yaml"
)

var _ = Describe("SplitYAML", func() {
	DescribeTable("should split documents and skip comment-only ones",
		func(input string, expected []string) {
			Expect(util_yaml.SplitYAML(input)).To(Equal(expected))
		},
		Entry("single document",
			"type: Mesh\nname: default\n",
			[]string{"type: Mesh\nname: default"},
		),
		Entry("multiple documents",
			"type: Mesh\nname: m1\n---\ntype: Mesh\nname: m2\n",
			[]string{"type: Mesh\nname: m1", "type: Mesh\nname: m2"},
		),
		Entry("single-line comment document",
			"# comment\n---\ntype: Mesh\nname: default\n",
			[]string{"type: Mesh\nname: default"},
		),
		Entry("multi-line comment document",
			"# first line\n# second line\n---\ntype: Mesh\nname: default\n",
			[]string{"type: Mesh\nname: default"},
		),
		Entry("commented out document",
			"type: Mesh\nname: m1\n---\n# type: Mesh\n# name: m2\n---\ntype: Mesh\nname: m3\n",
			[]string{"type: Mesh\nname: m1", "type: Mesh\nname: m3"},
		),
		Entry("comments inside a document are kept",
			"# Mesh m1\ntype: Mesh\n# name of the mesh\nname: m1\n",
			[]string{"# Mesh m1\ntype: Mesh\n# name of the mesh\nname: m1"},
		),
	)
})
