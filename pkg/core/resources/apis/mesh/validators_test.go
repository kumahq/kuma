package mesh_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	. "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
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
					"MeshSubset",
				},
			},
		}),
		Entry("MeshSubset without tags", testCase{
			inputYaml: `
kind: MeshSubset
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					"MeshSubset",
				},
			},
		}),
		Entry("MeshService", testCase{
			inputYaml: `
kind: MeshService
labels:
  kuma.io/display-name: backend
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
		}),
		Entry("Dataplane", testCase{
			inputYaml: `
kind: Dataplane
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Dataplane,
				},
			},
		}),
		Entry("Dataplane by labels", testCase{
			inputYaml: `
kind: Dataplane
labels:
  app: demo-app
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Dataplane,
				},
			},
		}),
		Entry("Dataplane with sectionName", testCase{
			inputYaml: `
kind: Dataplane
labels:
  app: demo-app
sectionName: http-port
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Dataplane,
				},
				IsInboundPolicy: true,
			},
		}),
		Entry("MeshHTTPRoute", testCase{
			inputYaml: `
kind: MeshHTTPRoute
labels:
  kuma.io/display-name: http-route1
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
labels:
  kuma.io/display-name: backend
tags:
  version: v1
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					"MeshServiceSubset",
				},
			},
		}),
		Entry("MeshServiceSubset without tags", testCase{
			inputYaml: `
kind: MeshServiceSubset
labels:
  kuma.io/display-name: backend
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					"MeshServiceSubset",
				},
			},
		}),
		Entry("MeshServiceSubset with empty tags", testCase{
			inputYaml: `
kind: MeshServiceSubset
labels:
  kuma.io/display-name: backend
tags: {}
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					"MeshServiceSubset",
				},
			},
		}),
		Entry("MeshService with name and namespace", testCase{
			inputYaml: `
kind: MeshService
labels:
  kuma.io/display-name: backend
  k8s.kuma.io/namespace: test-ns
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
labels:
  kuma.io/display-name: backend
  k8s.kuma.io/namespace: test-ns
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
labels:
  kuma.io/display-name: backend
  k8s.kuma.io/namespace: test-ns
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
					"MeshSubset",
				},
			},
			expected: `
violations:
  - field: targetRef.kind
    message: value 'Mesh' is not supported
`,
		}),
		Entry("Mesh with labels", testCase{
			inputYaml: `
kind: Mesh
labels:
  kuma.io/display-name: mesh-1
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
    message: value 'MeshSubset' is not supported`,
		}),
		Entry("MeshService when it's not supported", testCase{
			inputYaml: `
kind: MeshService
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					"MeshServiceSubset",
				},
			},
			expected: `
violations:
  - field: targetRef.kind
    message: value 'MeshService' is not supported`,
		}),
		Entry("MeshService without labels", testCase{
			inputYaml: `
kind: MeshService
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.labels
    message: must be set when kind is MeshService
`,
		}),
		Entry("MeshService backendRef without labels", testCase{
			inputYaml: `
kind: MeshService
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
				IsBackendRef: true,
			},
			expected: `
violations:
  - field: targetRef.labels
    message: must be set when kind is MeshService
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
    message: value 'MeshGateway' is not supported`,
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
    message: value 'MeshServiceSubset' is not supported
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
    message: value 'MeshGatewayRoute' is not supported
`,
		}),
		Entry("Mesh should not combine labels with sectionName", testCase{
			inputYaml: `
kind: Mesh
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
  - field: targetRef.sectionName
    message: must not be set with kind Mesh
`,
		}),
		Entry("MeshService should require labels", testCase{
			inputYaml: `
kind: MeshService
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshService,
				},
			},
			expected: `
violations:
  - field: targetRef.labels
    message: must be set when kind is MeshService
`,
		}),
		Entry("MeshHTTPRoute should require labels", testCase{
			inputYaml: `
kind: MeshHTTPRoute
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshHTTPRoute,
				},
			},
			expected: `
violations:
  - field: targetRef.labels
    message: must be set when kind is MeshHTTPRoute
`,
		}),
		Entry("MeshExternalService should require labels", testCase{
			inputYaml: `
kind: MeshExternalService
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.MeshExternalService,
				},
			},
			expected: `
violations:
  - field: targetRef.labels
    message: must be set when kind is MeshExternalService
`,
		}),
		Entry("Mesh should not be used with labels or sectionName", testCase{
			inputYaml: `
kind: Mesh
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
- field: targetRef.sectionName
  message: must not be set with kind Mesh
`,
		}),
		Entry("Dataplane with labels only", testCase{
			inputYaml: `
kind: Dataplane
labels:
  kuma.io/zone: east
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Dataplane,
				},
			},
			expected: `
violations: null
`,
		}),
		Entry("Dataplane with sectionName on a non-inbound policy", testCase{
			inputYaml: `
kind: Dataplane
labels:
  app: demo
sectionName: http-port
`,
			opts: &ValidateTargetRefOpts{
				SupportedKinds: []common_api.TargetRefKind{
					common_api.Dataplane,
				},
			},
			expected: `
violations:
- field: targetRef.sectionName
  message: can only be used with inbound policies
`,
		}),
	)
})
