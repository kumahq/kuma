package validators_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

var _ = Describe("TargetRef Validator", func() {
	type testCase struct {
		inputYaml string
		opts      *matcher_validators.ValidateTargetRefOpts
		// only used for failed testCases
		expected string
	}

	DescribeTable("should pass validation with",
		func(given testCase) {
			// given
			Expect(given.expected).To(BeEmpty())
			targetRef := common_api.TargetRef{}
			err := yaml.Unmarshal([]byte(given.inputYaml), &targetRef)
			Expect(err).ToNot(HaveOccurred())

			// when
			validationErr := validators.ValidationError{}
			validationErr.AddError("targetRef", matcher_validators.ValidateTargetRef(targetRef, given.opts))

			// then
			Expect(validationErr.OrNil()).To(Succeed())
		},
		Entry("Mesh", testCase{
			inputYaml: `
kind: Mesh
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Mesh,
				},
			},
		}),
		Entry("MeshSubset with tags", testCase{
			inputYaml: `
kind: MeshSubset
tags:
  kuma.io/zone: us-east
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshSubset,
				},
			},
		}),
		Entry("MeshSubset without tags", testCase{
			inputYaml: `
kind: MeshSubset
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshSubset,
				},
			},
		}),
		Entry("MeshService", testCase{
			inputYaml: `
kind: MeshService
name: backend
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
		}),
		Entry("MeshServiceSubset", testCase{
			inputYaml: `
kind: MeshServiceSubset
name: backend
tags:
  version: v1
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshServiceSubset,
				},
			},
		}),
		Entry("MeshServiceSubset without tags", testCase{
			inputYaml: `
kind: MeshServiceSubset
name: backend
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshServiceSubset,
				},
			},
		}),
		Entry("MeshServiceSubset with empty tags", testCase{
			inputYaml: `
kind: MeshServiceSubset
name: backend
tags: {}
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshServiceSubset,
				},
			},
		}),
		Entry("MeshGatewayRoute", testCase{
			inputYaml: `
kind: MeshGatewayRoute
name: backend-gateway-route
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshGatewayRoute,
				},
			},
		}),
	)

	DescribeTable("should return as much individual errors as possible with",
		func(given testCase) {
			// given
			targetRef := common_api.TargetRef{}
			err := yaml.Unmarshal([]byte(given.inputYaml), &targetRef)
			Expect(err).ToNot(HaveOccurred())

			// when
			validationErr := validators.ValidationError{}
			validationErr.AddError("targetRef", matcher_validators.ValidateTargetRef(targetRef, given.opts))
			// and
			actual, err := yaml.Marshal(validationErr)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("empty", testCase{
			inputYaml: `
{}
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Mesh,
				},
			},
			expected: `
violations:
  - field: targetRef.kind
    message: must be set 
`,
		}),
		Entry("Mesh when it's not supported", testCase{
			inputYaml: `
kind: Mesh
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshSubset,
				},
			},
			expected: `
violations:
  - field: targetRef.kind
    message: value is not supported
`,
		}),
		Entry("Mesh with mesh and tags", testCase{
			inputYaml: `
kind: Mesh
mesh: mesh-1
tags:
  tag1: value1
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Mesh,
				},
			},
			expected: `
violations:
  - field: targetRef.mesh
    message: cannot be set with kind Mesh
  - field: targetRef.tags
    message: cannot be set with kind Mesh
`,
		}),
		Entry("Mesh with name", testCase{
			inputYaml: `
kind: Mesh
name: mesh-1
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Mesh,
				},
			},
			expected: `
violations:
  - field: targetRef.name
    message: using name with kind Mesh is not yet supported 
`,
		}),
		Entry("MeshSubset when it's not supported", testCase{
			inputYaml: `
kind: MeshSubset
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Mesh,
				},
			},
			expected: `
violations:
  - field: targetRef.kind
    message: value is not supported`,
		}),
		Entry("MeshSubset with name", testCase{
			inputYaml: `
kind: MeshSubset
name: mesh-1
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshSubset,
				},
			},
			expected: `
violations:
  - field: targetRef.name
    message: cannot be set with kind MeshSubset`,
		}),
		Entry("MeshService when it's not supported", testCase{
			inputYaml: `
kind: MeshService
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshServiceSubset,
				},
			},
			expected: `
violations:
  - field: targetRef.kind
    message: value is not supported`,
		}),
		Entry("MeshService with mesh", testCase{
			inputYaml: `
kind: MeshService
name: backend
mesh: mesh-1
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.mesh
    message: cannot be set with kind MeshService
`,
		}),
		Entry("MeshService without name with tags", testCase{
			inputYaml: `
kind: MeshService
tags:
  tag1: value1
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.name
    message: must be set with kind MeshService 
  - field: targetRef.tags
    message: cannot be set with kind MeshService
`,
		}),
		Entry("MeshServiceSubset when it's not supported", testCase{
			inputYaml: `
kind: MeshServiceSubset
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.kind
    message: value is not supported
`,
		}),
		Entry("MeshServiceSubset without name with empty tags", testCase{
			inputYaml: `
kind: MeshServiceSubset
tags: {}
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshServiceSubset,
				},
			},
			expected: `
violations:
  - field: targetRef.name
    message: must be set with kind MeshServiceSubset
`,
		}),
		Entry("MeshServiceSubset with mesh", testCase{
			inputYaml: `
kind: MeshServiceSubset
name: backend
mesh: mesh-1
tags:
  version: v1
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshServiceSubset,
				},
			},
			expected: `
violations:
  - field: targetRef.mesh
    message: cannot be set with kind MeshServiceSubset 
`,
		}),
		Entry("MeshGatewayRoute when it's not supported", testCase{
			inputYaml: `
kind: MeshGatewayRoute
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Mesh,
				},
			},
			expected: `
violations:
  - field: targetRef.kind
    message: value is not supported
`,
		}),
		Entry("MeshGatewayRoute without name with mesh and tags", testCase{
			inputYaml: `
kind: MeshGatewayRoute
mesh: mesh-1
tags:
  tag1: value1
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshGatewayRoute,
				},
			},
			expected: `
violations:
  - field: targetRef.name
    message: must be set with kind MeshGatewayRoute
  - field: targetRef.mesh
    message: cannot be set with kind MeshGatewayRoute
`,
		}),
		Entry("MeshHTTPRoute when it's not supported", testCase{
			inputYaml: `
kind: MeshHTTPRoute
`,
			opts: &matcher_validators.ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshGatewayRoute,
				},
			},
			expected: `
violations:
  - field: targetRef.kind
    message: value is not supported
`,
		}),
	)
})
