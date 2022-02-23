package errors_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_errors "github.com/kumahq/kuma/app/kumactl/pkg/errors"
	"github.com/kumahq/kuma/pkg/core/rest/errors/types"
)

var _ = Describe("Formatter test", func() {

	type testCase struct {
		err error
		msg string
	}
	DescribeTable("should handle errors",
		func(given testCase) {
			fn := kumactl_errors.FormatErrorWrapper(func(_ *cobra.Command, _ []string) error {
				return given.err
			})

			// when
			err := fn(nil, nil)

			// then
			Expect(err).To(HaveOccurred())

			// and
			Expect(err.Error()).To(Equal(given.msg))
		},
		Entry("kuma api error", testCase{
			err: &types.Error{
				Title:   "Could not process the resource",
				Details: "Resource is invalid",
				Causes: []types.Cause{
					{
						Field:   "path",
						Message: "cannot be empty",
					},
					{
						Field:   "mesh",
						Message: "cannot be empty",
					},
				},
			},
			msg: `Could not process the resource (Resource is invalid)
* path: cannot be empty
* mesh: cannot be empty`,
		}),
		Entry("kuma api error even when it is wrapped", testCase{
			err: errors.Wrap(&types.Error{
				Title:   "Could not get the resource",
				Details: "Internal Server Error",
			}, "failed"),
			msg: `Could not get the resource (Internal Server Error)`,
		}),
		Entry("unknown error", testCase{
			err: errors.New("some error"),
			msg: "some error",
		}),
	)

	It("should do nothing when there is no error", func() {
		// given
		fn := kumactl_errors.FormatErrorWrapper(func(_ *cobra.Command, _ []string) error {
			return nil
		})

		// when
		err := fn(nil, nil)

		// then
		Expect(err).ToNot(HaveOccurred())
	})

})
