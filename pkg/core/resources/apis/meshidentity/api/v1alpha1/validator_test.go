package v1alpha1_test

import (
	"errors"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/validator"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/test"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var _ = Describe("MeshIdentity", func() {
	DescribeTable("should validate all valid folders", func(inputFile string) {
		// setup
		meshIdentity := v1alpha1.NewMeshIdentityResource()

		// when
		contents, err := os.ReadFile(inputFile)
		Expect(err).ToNot(HaveOccurred())
		err = core_model.FromYAML(contents, &meshIdentity.Spec)
		Expect(err).ToNot(HaveOccurred())

		meshIdentity.SetMeta(&test_model.ResourceMeta{
			Name: "test",
			Mesh: core_model.DefaultMesh,
		})

		// and
		verr := meshIdentity.Validate()
		actual, err := yaml.Marshal(verr)
		if string(actual) == "null\n" {
			actual = []byte{}
		}
		Expect(err).ToNot(HaveOccurred())

		// then
		goldenFile := strings.ReplaceAll(inputFile, ".input.yaml", ".golden.yaml")
		Expect(actual).To(matchers.MatchGoldenYAML(goldenFile))
	}, test.EntriesForFolder("spec"))

	DescribeTable("should reject SPIFFE ID edits on update",
		func(previous *v1alpha1.SpiffeID, current *v1alpha1.SpiffeID, expectedFields []string) {
			err := validator.ValidateUpdate(identityWithSpiffeID(previous), identityWithSpiffeID(current))

			if len(expectedFields) == 0 {
				Expect(err).ToNot(HaveOccurred())
				return
			}
			verr := &validators.ValidationError{}
			Expect(errors.As(err, &verr)).To(BeTrue())
			var fields []string
			for _, violation := range verr.Violations {
				fields = append(fields, violation.Field)
				Expect(violation.Message).To(HavePrefix("is immutable, cannot be changed"))
			}
			Expect(fields).To(Equal(expectedFields))
		},
		Entry("unchanged SPIFFE ID",
			spiffeID("{{ .Mesh }}.{{ .Zone }}.mesh.local", "/ns/{{ .Namespace }}"),
			spiffeID("{{ .Mesh }}.{{ .Zone }}.mesh.local", "/ns/{{ .Namespace }}"),
			nil,
		),
		Entry("changed trust domain",
			spiffeID("old.mesh.local", "/ns/{{ .Namespace }}"),
			spiffeID("new.mesh.local", "/ns/{{ .Namespace }}"),
			[]string{"spec.spiffeID.trustDomain"},
		),
		Entry("changed path",
			spiffeID("old.mesh.local", "/ns/{{ .Namespace }}"),
			spiffeID("old.mesh.local", "/workload/{{ .Workload }}"),
			[]string{"spec.spiffeID.path"},
		),
		Entry("both changed",
			spiffeID("old.mesh.local", "/ns/{{ .Namespace }}"),
			spiffeID("new.mesh.local", "/workload/{{ .Workload }}"),
			[]string{"spec.spiffeID.trustDomain", "spec.spiffeID.path"},
		),
		Entry("trust domain set on an identity that relied on the default",
			nil,
			spiffeID("new.mesh.local", ""),
			[]string{"spec.spiffeID.trustDomain"},
		),
		Entry("trust domain removed to fall back on the default",
			spiffeID("old.mesh.local", ""),
			nil,
			[]string{"spec.spiffeID.trustDomain"},
		),
		Entry("SPIFFE ID absent on both revisions", nil, nil, nil),
	)
})

func spiffeID(trustDomain string, path string) *v1alpha1.SpiffeID {
	id := &v1alpha1.SpiffeID{}
	if trustDomain != "" {
		id.TrustDomain = pointer.To(trustDomain)
	}
	if path != "" {
		id.Path = pointer.To(path)
	}
	return id
}

func identityWithSpiffeID(id *v1alpha1.SpiffeID) *v1alpha1.MeshIdentityResource {
	identity := v1alpha1.NewMeshIdentityResource()
	identity.Spec.SpiffeID = id
	identity.SetMeta(&test_model.ResourceMeta{Name: "test", Mesh: core_model.DefaultMesh})
	return identity
}
