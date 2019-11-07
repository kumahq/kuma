package validators_test

import (
	"github.com/Kong/kuma/pkg/core/validators"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
})
