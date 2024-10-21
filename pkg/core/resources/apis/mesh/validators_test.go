package mesh_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	. "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
)

var _ = Describe("AllowedValuesHint()", func() {
	type testCase struct {
		values   []string
		expected string
	}

	DescribeTable("should generate a proper hint",
		func(given testCase) {
			Expect(AllowedValuesHint(given.values...)).To(Equal(given.expected))
		},
		Entry("nil list", testCase{
			values:   nil,
			expected: `Allowed values: (none)`,
		}),
		Entry("empty list", testCase{
			values:   []string{},
			expected: `Allowed values: (none)`,
		}),
		Entry("one-item list", testCase{
			values:   []string{"http"},
			expected: `Allowed values: http`,
		}),
		Entry("multi-item list", testCase{
			values:   []string{"grpc", "http", "http2", "mongo", "mysql", "redis", "tcp"},
			expected: `Allowed values: grpc, http, http2, mongo, mysql, redis, tcp`,
		}),
	)
})

var _ = Describe("selector tag keys", func() {
	type testCase struct {
		tags      map[string]string
		validator TagKeyValidatorFunc
		violation *validators.Violation
	}

	DescribeTable("should validate",
		func(given testCase) {
			err := ValidateSelector(validators.RootedAt("given"), given.tags,
				ValidateTagsOpts{
					ExtraTagKeyValidators: []TagKeyValidatorFunc{given.validator},
				},
			)

			switch len(err.Violations) {
			case 0:
				Expect(given.violation).To(BeNil())
			case 1:
				Expect(err.Violations[0]).To(Equal(*given.violation))
			default:
				Expect(len(err.Violations)).To(BeNumerically("<=", 1))
			}
		},

		Entry("noop", testCase{
			tags: map[string]string{
				"foo": "bar",
			},
			validator: TagKeyValidatorFunc(func(path validators.PathBuilder, key string) validators.ValidationError {
				return validators.ValidationError{}
			}),
		}),

		Entry("selector key is not in set", testCase{
			validator: SelectorKeyNotInSet("foo", "bar"),
			tags: map[string]string{
				"baz": "bar",
				"boo": "bar",
				"bar": "bar",
			},
			violation: &validators.Violation{
				Field:   `given["bar"]`,
				Message: `tag name must not be "bar"`,
			},
		}),

		Entry("selector key is not matched in set", testCase{
			validator: SelectorKeyNotInSet("not", "there"),
			tags: map[string]string{
				"baz": "bar",
				"boo": "bar",
				"bar": "bar",
			},
		}),
	)
})

var _ = Describe("TargetRef Validator", func() {
	type testCase struct {
		inputYaml string
		opts      *ValidateTargetRefOpts
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
			validationErr.AddError("targetRef", ValidateTargetRef(targetRef, given.opts))

			// then
			Expect(validationErr.OrNil()).To(Succeed())
		},
		Entry("Mesh", testCase{
			inputYaml: `
kind: Mesh
`,
			opts: &ValidateTargetRefOpts{
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
  validTagName: foo
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshSubset,
				},
			},
		}),
		Entry("MeshSubset without tags", testCase{
			inputYaml: `
kind: MeshSubset
`,
			opts: &ValidateTargetRefOpts{
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
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
		}),
		Entry("MeshGateway", testCase{
			inputYaml: `
kind: MeshGateway
name: gateway1
tags:
  listener: one
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshGateway,
				},
				GatewayListenerTagsAllowed: true,
			},
		}),
		Entry("MeshGateway with period", testCase{
			inputYaml: `
kind: MeshGateway
name: gateway.namespace
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshGateway,
				},
			},
		}),
		Entry("MeshHTTPRoute", testCase{
			inputYaml: `
kind: MeshHTTPRoute
name: http-route1
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshHTTPRoute,
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
			opts: &ValidateTargetRefOpts{
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
			opts: &ValidateTargetRefOpts{
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
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshServiceSubset,
				},
			},
		}),
		Entry("MeshService with name and namespace", testCase{
			inputYaml: `
kind: MeshService
name: backend
namespace: test-ns
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
		}),
		Entry("MeshService with labels", testCase{
			inputYaml: `
kind: MeshService
labels: 
  kuma.io/zone: east
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
		}),
		Entry("MeshService with name, namespace and sectionName", testCase{
			inputYaml: `
kind: MeshService
name: backend
namespace: test-ns
sectionName: http-port
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
		}),
		Entry("MeshExternalService with name and namespace", testCase{
			inputYaml: `
kind: MeshExternalService
name: backend
namespace: test-ns
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshExternalService,
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
			validationErr.AddError("targetRef", ValidateTargetRef(targetRef, given.opts))
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
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Mesh,
				},
			},
			expected: `
violations:
  - field: targetRef.kind
    message: must be defined
`,
		}),
		Entry("Mesh when it's not supported", testCase{
			inputYaml: `
kind: Mesh
`,
			opts: &ValidateTargetRefOpts{
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
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Mesh,
				},
			},
			expected: `
violations:
  - field: targetRef.mesh
    message: must not be set with kind Mesh
  - field: targetRef.tags
    message: must not be set with kind Mesh
`,
		}),
		Entry("Mesh with name", testCase{
			inputYaml: `
kind: Mesh
name: mesh-1
`,
			opts: &ValidateTargetRefOpts{
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
			opts: &ValidateTargetRefOpts{
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
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshSubset,
				},
			},
			expected: `
violations:
  - field: targetRef.name
    message: must not be set with kind MeshSubset`,
		}),
		Entry("MeshSubset with empty tag name", testCase{
			inputYaml: `
kind: MeshSubset
tags: 
  "": value1
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshSubset,
				},
			},
			expected: `
violations:
  - field: targetRef.tags
    message: tag name must be non-empty`,
		}),
		Entry("MeshSubset with invalid tag name", testCase{
			inputYaml: `
kind: MeshSubset
tags: 
  invalidTag*: value1
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshSubset,
				},
			},
			expected: `
violations:
  - field: targetRef.tags["invalidTag*"]
    message: tag name must consist of alphanumeric characters, dots, dashes, slashes and underscores`,
		}),
		Entry("MeshSubset with empty tag value", testCase{
			inputYaml: `
kind: MeshSubset
tags: 
  tag1: ""
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshSubset,
				},
			},
			expected: `
violations:
  - field: targetRef.tags["tag1"]
    message: tag value must be non-empty`,
		}),
		Entry("MeshSubset with invalid tag value", testCase{
			inputYaml: `
kind: MeshSubset
tags: 
  tag1: invalidValue?
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshSubset,
				},
			},
			expected: `
violations:
  - field: targetRef.tags["tag1"]
    message: tag value must consist of alphanumeric characters, dots, dashes and underscores`,
		}),
		Entry("MeshService when it's not supported", testCase{
			inputYaml: `
kind: MeshService
`,
			opts: &ValidateTargetRefOpts{
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
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.mesh
    message: must not be set with kind MeshService
`,
		}),
		Entry("MeshService without name with tags", testCase{
			inputYaml: `
kind: MeshService
tags:
  tag1: value1
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.tags
    message: must not be set with kind MeshService
  - field: targetRef.name
    message: must be set with kind MeshService
`,
		}),
		Entry("MeshService with invalid name", testCase{
			inputYaml: `
kind: MeshService
name: "*"
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.name
    message: "invalid characters: must consist of lower case alphanumeric characters, '-', '.' and '_'."
`,
		}),
		Entry("MeshService with proxyTypes", testCase{
			inputYaml: `
kind: MeshService
name: "test"
proxyTypes: ["Sidecar"]
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.proxyTypes
    message: must not be set with kind MeshService
`,
		}),
		Entry("MeshServiceSubset with proxyTypes", testCase{
			inputYaml: `
kind: MeshServiceSubset
name: "test"
proxyTypes: ["Sidecar"]
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshServiceSubset,
				},
			},
			expected: `
violations:
  - field: targetRef.proxyTypes
    message: must not be set with kind MeshServiceSubset
`,
		}),
		Entry("MeshGateway with proxyTypes", testCase{
			inputYaml: `
kind: MeshGateway
name: "test"
proxyTypes: ["Sidecar"]
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshGateway,
				},
			},
			expected: `
violations:
  - field: targetRef.proxyTypes
    message: must not be set with kind MeshGateway
`,
		}),
		Entry("MeshGateway when it's not supported", testCase{
			inputYaml: `
kind: MeshGateway
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.kind
    message: value is not supported`,
		}),
		Entry("MeshGateway with mesh", testCase{
			inputYaml: `
kind: MeshGateway
name: gateway1
mesh: mesh-1
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshGateway,
				},
			},
			expected: `
violations:
  - field: targetRef.mesh
    message: must not be set with kind MeshGateway
`,
		}),
		Entry("MeshGateway without name with empty tags", testCase{
			inputYaml: `
kind: MeshGateway
tags: {}
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshGateway,
				},
			},
			expected: `
violations:
  - field: targetRef.name
    message: must be set with kind MeshGateway 
`,
		}),
		Entry("MeshGateway with invalid name", testCase{
			inputYaml: `
kind: MeshGateway
name: "*"
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshGateway,
				},
			},
			expected: `
violations:
  - field: targetRef.name
    message: "invalid characters: must consist of lower case alphanumeric characters, '-', '.' and '_'."
`,
		}),
		Entry("MeshServiceSubset when it's not supported", testCase{
			inputYaml: `
kind: MeshServiceSubset
`,
			opts: &ValidateTargetRefOpts{
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
			opts: &ValidateTargetRefOpts{
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
		Entry("MeshServiceSubset with invalid name with empty tags", testCase{
			inputYaml: `
kind: MeshServiceSubset
name: "*"
tags: {}
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshServiceSubset,
				},
			},
			expected: `
violations:
 - field: targetRef.name
   message: "invalid characters: must consist of lower case alphanumeric characters, '-', '.' and '_'."
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
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshServiceSubset,
				},
			},
			expected: `
violations:
  - field: targetRef.mesh
    message: must not be set with kind MeshServiceSubset 
`,
		}),
		Entry("MeshGatewayRoute when it's not supported", testCase{
			inputYaml: `
kind: MeshGatewayRoute
`,
			opts: &ValidateTargetRefOpts{
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
		Entry("MeshGateway and no tags allowed", testCase{
			inputYaml: `
kind: MeshGateway
name: edge
tags:
  port: http
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshGateway,
				},
			},
			expected: `
violations:
  - field: targetRef.tags
    message: must not be set with kind MeshGateway
`,
		}),
		Entry("MeshService should not combine name with labels", testCase{
			inputYaml: `
kind: MeshService
name: test
labels:
  kuma.io/zone: east
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.labels
    message: either labels or name and namespace must be specified
`,
		}),
		Entry("MeshService should not combine name/namespace with labels", testCase{
			inputYaml: `
kind: MeshService
namespace: test-ns
labels:
  kuma.io/zone: east
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.labels
    message: either labels or name and namespace must be specified
`,
		}),
		Entry("MeshService should not combine name and namespace with labels", testCase{
			inputYaml: `
kind: MeshService
name: test
namespace: test-ns
labels:
  kuma.io/zone: east
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.labels
    message: either labels or name and namespace must be specified
`,
		}),
		Entry("MeshService should have name when labels are not specified", testCase{
			inputYaml: `
kind: MeshService
namespace: test-ns
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.name
    message: must be set with kind MeshService
`,
		}),
		Entry("Mesh should not be used with namespace or labels or sectionName", testCase{
			inputYaml: `
kind: Mesh
namespace: test-ns
labels:
  kuma.io/zone: east
sectionName: port-http
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Mesh,
				},
			},
			expected: `
violations:
- field: targetRef.labels
  message: must not be set with kind Mesh
- field: targetRef.namespace
  message: must not be set with kind Mesh
- field: targetRef.sectionName
  message: must not be set with kind Mesh
`,
		}),
		Entry("MeshSubset should not be used with namespace or labels or sectionName", testCase{
			inputYaml: `
kind: MeshSubset
namespace: test-ns
labels:
  kuma.io/zone: east
sectionName: port-http
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshSubset,
				},
			},
			expected: `
violations:
- field: targetRef.labels
  message: must not be set with kind MeshSubset
- field: targetRef.namespace
  message: must not be set with kind MeshSubset
- field: targetRef.sectionName
  message: must not be set with kind MeshSubset
`,
		}),
		Entry("MeshServiceSubset should not be used with namespace or labels or sectionName", testCase{
			inputYaml: `
kind: MeshServiceSubset
name: test
namespace: test-ns
labels:
  kuma.io/zone: east
sectionName: port-http
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshServiceSubset,
				},
			},
			expected: `
violations:
- field: targetRef.labels
  message: must not be set with kind MeshServiceSubset
- field: targetRef.namespace
  message: must not be set with kind MeshServiceSubset
- field: targetRef.sectionName
  message: must not be set with kind MeshServiceSubset
`,
		}),
	)
})
