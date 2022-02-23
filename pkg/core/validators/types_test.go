package validators_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/validators"
)

var _ = Describe("Validation Error", func() {
	It("should construct errors", func() {
		// given
		err := validators.ValidationError{}

		// when
		err.AddViolation("name", "invalid name")

		// and
		addressErr := validators.ValidationError{}
		addressErr.AddViolation("street", "invalid format")
		err.AddError("address", addressErr)

		// then
		Expect(err.HasViolations()).To(BeTrue())
		Expect(validators.IsValidationError(&err)).To(BeTrue())
		Expect(err.OrNil()).To(MatchError("name: invalid name; address.street: invalid format"))
	})

	It("should convert to nil error when there are no violations", func() {
		// given
		validationErr := validators.ValidationError{}

		// when
		err := validationErr.OrNil()

		Expect(err).To(BeNil())
	})

	Describe("Append()", func() {
		It("should add a given error to the end of the list", func() {
			// given
			err := validators.ValidationError{}
			err1 := validators.ValidationError{}
			err1.AddViolationAt(validators.RootedAt("sources"), "unknown error")
			err2 := validators.ValidationError{}
			err2.AddViolationAt(validators.RootedAt("destinations"), "yet another error")

			By("adding the first error")
			// when
			err.Add(err1)
			// then
			Expect(err).To(Equal(validators.ValidationError{
				Violations: []validators.Violation{
					{Field: "sources", Message: "unknown error"},
				},
			}))

			By("adding the second error")
			// when
			err.Add(err2)
			// then
			Expect(err).To(Equal(validators.ValidationError{
				Violations: []validators.Violation{
					{Field: "sources", Message: "unknown error"},
					{Field: "destinations", Message: "yet another error"},
				},
			}))
		})
	})

	Describe("AddViolationAt()", func() {
		It("should accept nil PathBuilder", func() {
			// given
			err := validators.ValidationError{}
			// when
			err.AddViolationAt(nil, "unknown error")
			// then
			Expect(err).To(Equal(validators.ValidationError{
				Violations: []validators.Violation{
					{Field: "", Message: "unknown error"},
				},
			}))
		})

		It("should accept non-nil PathBuilder", func() {
			// given
			err := validators.ValidationError{}
			path := validators.RootedAt("sources").Index(0).Field("match").Key("service")
			// when
			err.AddViolationAt(path, "unknown error")
			// and
			Expect(err).To(Equal(validators.ValidationError{
				Violations: []validators.Violation{
					{Field: `sources[0].match["service"]`, Message: "unknown error"},
				},
			}))
		})
	})

	Describe("Transform()", func() {
		type testCase struct {
			input         *validators.ValidationError
			transformFunc func(validators.Violation) validators.Violation
			expected      *validators.ValidationError
		}

		DescribeTable("should apply given transformation func to every Violation",
			func(given testCase) {
				// when
				actual := given.input.Transform(given.transformFunc)
				// then
				Expect(actual).To(Equal(given.expected))
			},
			Entry("`nil` ValidationError", testCase{
				input:    nil,
				expected: nil,
			}),
			Entry("zero value ValidationError", testCase{
				input:    &validators.ValidationError{},
				expected: &validators.ValidationError{},
			}),
			Entry("`nil` transformFunc", testCase{
				input: &validators.ValidationError{
					Violations: []validators.Violation{
						{Field: "field", Message: "invalid"},
					},
				},
				expected: &validators.ValidationError{
					Violations: []validators.Violation{
						{Field: "field", Message: "invalid"},
					},
				},
			}),
			Entry("identity transform", testCase{
				input: &validators.ValidationError{
					Violations: []validators.Violation{
						{Field: "field", Message: "invalid"},
					},
				},
				transformFunc: func(v validators.Violation) validators.Violation {
					return v
				},
				expected: &validators.ValidationError{
					Violations: []validators.Violation{
						{Field: "field", Message: "invalid"},
					},
				},
			}),
			Entry("real transform", testCase{
				input: &validators.ValidationError{
					Violations: []validators.Violation{
						{Field: "field1", Message: "invalid1"},
						{Field: "field2", Message: "invalid2"},
					},
				},
				transformFunc: func(v validators.Violation) validators.Violation {
					return validators.Violation{
						Field:   fmt.Sprintf("spec.%s", v.Field),
						Message: fmt.Sprintf("prefix: %s", v.Message),
					}
				},
				expected: &validators.ValidationError{
					Violations: []validators.Violation{
						{Field: "spec.field1", Message: "prefix: invalid1"},
						{Field: "spec.field2", Message: "prefix: invalid2"},
					},
				},
			}),
		)
	})
})

var _ = Describe("PathBuilder", func() {
	It("should produce empty path by default", func() {
		Expect(validators.PathBuilder{}.String()).To(Equal(""))
	})

	It("should produce valid root path", func() {
		Expect(validators.RootedAt("spec").String()).To(Equal("spec"))
	})

	It("should produce valid field path", func() {
		Expect(validators.RootedAt("spec").Field("sources").String()).To(Equal("spec.sources"))
	})

	It("should produce valid array index", func() {
		Expect(validators.RootedAt("spec").Field("sources").Index(0).String()).To(Equal("spec.sources[0]"))
	})

	It("should produce valid array index", func() {
		Expect(validators.RootedAt("spec").Field("sources").Index(0).Field("match").Key("service").String()).To(Equal(`spec.sources[0].match["service"]`))
	})
})
