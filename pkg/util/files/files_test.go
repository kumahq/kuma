package files_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/util/files"
)

var _ = DescribeTable("",
	func(in string, out string) string {
		return fmt.Sprintf(`ToUnixFilename(%q)=%q`, in, out)
	},
	func(in string, out string) {
		Expect(files.ToValidUnixFilename(in)).To(Equal(out))
	},
	Entry(nil, "", ""),
	Entry(nil, "foo", "foo"),
	Entry(nil, "-.foo", "foo"),
	Entry(nil, "C?on_between_old_an/d_new_DPP_\"from_version:_2-7-12", "C-on_between_old_an-d_new_DPP_-from_version-_2-7-12"),
	Entry(nil, "Compatibility_connection_between_old_and_new_DPP_from_version:_2-7-12", "Compatibility_connection_between_old_and_new_DPP_from_version-_2-7-12"),
)
