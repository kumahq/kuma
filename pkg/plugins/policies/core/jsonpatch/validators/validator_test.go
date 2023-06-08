package validators_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	jsonpatch_validators "github.com/kumahq/kuma/pkg/plugins/policies/core/jsonpatch/validators"
)

var _ = Describe("JsonPatchBlock Validator", func() {
	type testCase struct {
		patchBlocks []string
		// only used for failed testCases
		expected string
	}

	DescribeTable("should pass validation for operation",
		func(given testCase) {
			// given
			Expect(given.expected).To(BeEmpty())
			var patchBlocks []common_api.JsonPatchBlock

			for _, block := range given.patchBlocks {
				var patchBlock common_api.JsonPatchBlock
				Expect(yaml.Unmarshal([]byte(block), &patchBlock)).To(Succeed())
				patchBlocks = append(patchBlocks, patchBlock)
			}

			// when
			validationErr := jsonpatch_validators.ValidateJsonPatchBlock(
				validators.RootedAt("jsonPatches"),
				patchBlocks,
			)

			// then
			Expect(validationErr.OrNil()).To(Succeed())
		},
		Entry("add", testCase{
			patchBlocks: []string{
				`
                op: add
                path: /a/b/c
                value: d
                `,
				`
                op: add
                path: /foo
                value:
                - bar: yes
                  baz: true
                `,
				`
                op: add
                path: /foo
                value:
                  bar: a
                  baz:
                    faz: ["baz"]
                `,
			},
		}),
		Entry("replace", testCase{
			patchBlocks: []string{
				`
                op: replace
                path: /a/b/c
                value: d
                `,
				`
                op: replace
                path: /foo
                value:
                - bar: yes
                  baz: true
                `,
				`
                op: replace
                path: /foo
                value:
                  bar: a
                  baz:
                    faz: ["baz"]
                `,
			},
		}),
		Entry("remove", testCase{
			patchBlocks: []string{
				`
                op: remove
                path: /a/b/c
                `,
				`
                op: remove
                path: /foo/0
                `,
				`
                op: remove
                path: /foo/-
                `,
			},
		}),
		Entry("copy", testCase{
			patchBlocks: []string{
				`
                op: copy
                from: /a/b/c
                path: /a/b/d
                `,
				`
                op: copy
                from: /foo/0
                path: /foo/1
                `,
				`
                op: copy
                from: /foo/-
                path: /foo/0
                `,
			},
		}),
		Entry("move", testCase{
			patchBlocks: []string{
				`
                op: move
                from: /a/b/c
                path: /a/b/d
                `,
				`
                op: move
                from: /foo/0
                path: /bar/0
                `,
				`
                op: move
                from: /foo/-
                path: /bar/-
                `,
			},
		}),
	)

	DescribeTable("should return as much individual errors as possible with",
		func(given testCase) {
			// given
			var patchBlocks []common_api.JsonPatchBlock

			for _, block := range given.patchBlocks {
				var patchBlock common_api.JsonPatchBlock
				Expect(yaml.Unmarshal([]byte(block), &patchBlock)).To(Succeed())
				patchBlocks = append(patchBlocks, patchBlock)
			}

			// when
			validationErr := jsonpatch_validators.ValidateJsonPatchBlock(
				validators.RootedAt("jsonPatches"),
				patchBlocks,
			)

			// and
			actual, err := yaml.Marshal(validationErr)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("empty", testCase{
			patchBlocks: nil,
			expected: `
            violations:
              - field: jsonPatches
                message: must not be empty
            `,
		}),
		Entry("unsupported operations", testCase{
			patchBlocks: []string{
				`
                op: unsupported_operation
                path: ""
                `,
				`
                op: another_one
                path: "/"
                `,
			},
			expected: `
            violations:
              - field: jsonPatches[0].op
                message: '"op" must be one of ["add", "remove", "replace", "move", "copy"]'
              - field: jsonPatches[1].op
                message: '"op" must be one of ["add", "remove", "replace", "move", "copy"]'
            `,
		}),
		Entry("value field provided for unsupported operations", testCase{
			patchBlocks: []string{
				`
                op: remove
                path: /a/b/c
                value: foo
                `,
				`
                op: copy
                from: /d/e/f
                path: /a/b/c
                value:
                  foo: foo
                  bar: { "baz": [1, 2, 3] }
                `,
				`
                op: move
                from: /d/e/f
                path: /a/b/c
                value:
                - a: foo
                - a: bar
                `,
			},
			expected: `
            violations:
              - field: jsonPatches[0].value
                message: '"value" is allowed only when "op" is one of ["add", "replace"]'
              - field: jsonPatches[1].value
                message: '"value" is allowed only when "op" is one of ["add", "replace"]'
              - field: jsonPatches[2].value
                message: '"value" is allowed only when "op" is one of ["add", "replace"]'
            `,
		}),
		Entry("value field missing when expected to be provided", testCase{
			patchBlocks: []string{
				`
                op: add
                path: /a/b/c
                `,
				`
                op: replace
                path: ""
                `,
			},
			expected: `
            violations:
              - field: jsonPatches[0].value
                message: '"value" must not be empty when "op" is one of ["add", "replace"]'
              - field: jsonPatches[1].value
                message: '"value" must not be empty when "op" is one of ["add", "replace"]'
            `,
		}),
		Entry("from field present for unsupported operations", testCase{
			patchBlocks: []string{
				`
                op: add
                path: /a/b/c
                from: /
                value: foo
                `,
				`
                op: replace
                path: /a/b/c
                from: ""
                value:
                  foo: foo
                  bar: { "baz": [1, 2, 3] }
                `,
				`
                op: remove
                path: /a/b/c
                from: /d/e/f
                `,
			},
			expected: `
            violations:
              - field: jsonPatches[0].from
                message: '"from" is allowed only when "op" is one of ["move", "copy"]'
              - field: jsonPatches[1].from
                message: '"from" is allowed only when "op" is one of ["move", "copy"]'
              - field: jsonPatches[2].from
                message: '"from" is allowed only when "op" is one of ["move", "copy"]'
            `,
		}),
		Entry("from field missing when expected to be provided", testCase{
			patchBlocks: []string{
				`
                op: copy
                path: /a/b/c
                `,
				`
                op: move
                path: ""
                `,
			},
			expected: `
            violations:
              - field: jsonPatches[0].from
                message: '"from" must not be empty when "op" is one of ["move", "copy"]'
              - field: jsonPatches[1].from
                message: '"from" must not be empty when "op" is one of ["move", "copy"]'
            `,
		}),
		Entry("remove root path operation", testCase{
			patchBlocks: []string{
				`
                op: remove
                path: ""
                `,
				`
                op: remove
                path: "/"
                `,
			},
			expected: `
            violations:
              - field: jsonPatches[0].path
                message: root path cannot be removed
              - field: jsonPatches[1].path
                message: root path cannot be removed
            `,
		}),
	)
})
