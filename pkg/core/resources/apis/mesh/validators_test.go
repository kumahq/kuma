package mesh_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
