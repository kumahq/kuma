package validators_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("TargetRef Validator", func() {

	Context("pass validation", func() {

		type testCase struct {
			inputYaml string
			opts      *matcher_validators.ValidateTargetRefOpts
		}

		DescribeTable("should pass validation",
			func(given testCase) {
				// given
				targetRef := &common_proto.TargetRef{}
				err := util_proto.FromYAML([]byte(given.inputYaml), targetRef)
				Expect(err).ToNot(HaveOccurred())

				// when
				validationErr := matcher_validators.ValidateTargetRef(validators.RootedAt("targetRef"), targetRef, given.opts)

				// then
				Expect(validationErr.OrNil()).To(BeNil())
			},
			Entry("targetRef for Mesh", testCase{
				inputYaml: `
kind: Mesh
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_Mesh,
					},
				},
			}),
			Entry("targetRef for Mesh with name", testCase{
				inputYaml: `
kind: Mesh
name: mesh-1
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_Mesh,
					},
				},
			}),
			Entry("targetRef for MeshSubset", testCase{
				inputYaml: `
kind: MeshSubset
tags:
  kuma.io/zone: us-east
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshSubset,
					},
				},
			}),
			Entry("targetRef for MeshSubset with name", testCase{
				inputYaml: `
kind: MeshSubset
name: mesh-1
tags:
  kuma.io/zone: us-east
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshSubset,
					},
				},
			}),
			Entry("targetRef for MeshSubset with name without tags", testCase{
				inputYaml: `
kind: MeshSubset
name: mesh-1
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshSubset,
					},
				},
			}),
			Entry("targetRef for MeshService", testCase{
				inputYaml: `
kind: MeshService
name: backend
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshService,
					},
				},
			}),
			Entry("targetRef for MeshService with mesh", testCase{
				inputYaml: `
kind: MeshService
name: backend
mesh: mesh-1
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshService,
					},
				},
			}),
			Entry("targetRef for MeshServiceSubset", testCase{
				inputYaml: `
kind: MeshServiceSubset
name: backend
tags:
  version: v1
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshServiceSubset,
					},
				},
			}),
			Entry("targetRef for MeshServiceSubset with mesh", testCase{
				inputYaml: `
kind: MeshServiceSubset
name: backend
mesh: mesh-1
tags:
  version: v1
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshServiceSubset,
					},
				},
			}),
			Entry("targetRef for MeshGatewayRoute", testCase{
				inputYaml: `
kind: MeshGatewayRoute
name: backend-gateway-route
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshGatewayRoute,
					},
				},
			}),
			Entry("targetRef for MeshHTTPRoute", testCase{
				inputYaml: `
kind: MeshHTTPRoute
name: backend-http-route
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshHTTPRoute,
					},
				},
			}),
		)
	})

	Context("fail validation", func() {

		type testCase struct {
			inputYaml string
			opts      *matcher_validators.ValidateTargetRefOpts
			expected  string
		}

		DescribeTable("should validate all fields and return as much individual errors as possible",
			func(given testCase) {
				// given
				targetRef := &common_proto.TargetRef{}
				err := util_proto.FromYAML([]byte(given.inputYaml), targetRef)
				Expect(err).ToNot(HaveOccurred())

				// when
				validationErr := matcher_validators.ValidateTargetRef(validators.RootedAt("targetRef"), targetRef, given.opts)
				// and
				actual, err := yaml.Marshal(validationErr)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("targetRef for Mesh when it's not supported", testCase{
				inputYaml: `
kind: Mesh
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshSubset,
					},
				},
				expected: `
violations:
  - field: targetRef.kind
    message: value is not supported
`,
			}),
			Entry("targetRef for Mesh with mesh and tags", testCase{
				inputYaml: `
kind: Mesh
mesh: mesh-1
tags:
  tag1: value1
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_Mesh,
					},
				},
				expected: `
violations:
  - field: targetRef.tags
    message: could not be set with kind Mesh
  - field: targetRef.mesh
    message: could not be set with kind Mesh
`,
			}),
			Entry("targetRef for MeshSubset when it's not supported", testCase{
				inputYaml: `
kind: MeshSubset
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_Mesh,
					},
				},
				expected: `
violations:
  - field: targetRef.kind
    message: value is not supported`,
			}),
			Entry("targetRef for MeshSubset with empty tags", testCase{
				inputYaml: `
kind: MeshSubset
tags: {}
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshSubset,
					},
				},
				expected: `
violations:
  - field: targetRef.tags
    message: cannot be empty`,
			}),
			Entry("targetRef for MeshService when it's not supported", testCase{
				inputYaml: `
kind: MeshService
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshServiceSubset,
					},
				},
				expected: `
violations:
  - field: targetRef.kind
    message: value is not supported`,
			}),
			Entry("targetRef for MeshService without name with tags", testCase{
				inputYaml: `
kind: MeshService
tags:
  tag1: value1
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshService,
					},
				},
				expected: `
violations:
  - field: targetRef.tags
    message: could not be set with kind MeshService
  - field: targetRef.name
    message: cannot be empty
`,
			}),
			Entry("targetRef for MeshServiceSubset when it's not supported", testCase{
				inputYaml: `
kind: MeshServiceSubset
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshService,
					},
				},
				expected: `
violations:
  - field: targetRef.kind
    message: value is not supported
`,
			}),
			Entry("targetRef for MeshServiceSubset without name with empty tags", testCase{
				inputYaml: `
kind: MeshServiceSubset
tags: {}
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshServiceSubset,
					},
				},
				expected: `
violations:
  - field: targetRef.name
    message: cannot be empty
  - field: targetRef.tags
    message: cannot be empty
`,
			}),
			Entry("targetRef for MeshGatewayRoute when it's not supported", testCase{
				inputYaml: `
kind: MeshGatewayRoute
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshHTTPRoute,
					},
				},
				expected: `
violations:
  - field: targetRef.kind
    message: value is not supported
`,
			}),
			Entry("targetRef for MeshGatewayRoute without name with mesh and tags", testCase{
				inputYaml: `
kind: MeshGatewayRoute
mesh: mesh-1
tags:
  tag1: value1
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshGatewayRoute,
					},
				},
				expected: `
violations:
  - field: targetRef.name
    message: cannot be empty
  - field: targetRef.mesh
    message: could not be set with kind MeshGatewayRoute
`,
			}),
			Entry("targetRef for MeshHTTPRoute when it's not supported", testCase{
				inputYaml: `
kind: MeshHTTPRoute
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshGatewayRoute,
					},
				},
				expected: `
violations:
  - field: targetRef.kind
    message: value is not supported
`,
			}),
			Entry("targetRef for MeshHTTPRoute without name with mesh and tags", testCase{
				inputYaml: `
kind: MeshHTTPRoute
mesh: mesh-1
tags:
  tag1: value1
`,
				opts: &matcher_validators.ValidateTargetRefOpts{
					SupportedKinds: []common_proto.TargetRef_Kind{
						common_proto.TargetRef_MeshHTTPRoute,
					},
				},
				expected: `
violations:
  - field: targetRef.name
    message: cannot be empty
  - field: targetRef.mesh
    message: could not be set with kind MeshHTTPRoute
`,
			}),
		)
	})
})
