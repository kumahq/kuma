package files_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/pkg/util/files"
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

var _ = DescribeTable("RelativeToPkgMod",
	func(in string, out string) string {
		return fmt.Sprintf(`RelativeToPkgMod(%q)=%q`, in, out)
	},
	func(in string, out string) {
		Expect(files.RelativeToPkgMod(in)).To(Equal(out))
	},
	Entry("v1 module path",
		"/home/ubuntu/go/pkg/mod/github.com/kumahq/kuma@v1.8.0-20231106140736-df9c4e43a672/pkg/test/store/postgres/test_container.go",
		"/github.com/kumahq/kuma@v1.8.0-20231106140736-df9c4e43a672/pkg/test/store/postgres/test_container.go",
	),
	Entry("v2 module path",
		"/home/ubuntu/go/pkg/mod/github.com/kumahq/kuma/v2@v2.0.0-20251106140736-df9c4e43a672/pkg/test/store/postgres/test_container.go",
		"/github.com/kumahq/kuma/v2@v2.0.0-20251106140736-df9c4e43a672/pkg/test/store/postgres/test_container.go",
	),
	Entry("v2 module path with app directory",
		"/home/ubuntu/go/pkg/mod/github.com/kumahq/kuma/v2@v2.0.0-20251106140736-df9c4e43a672/app/kumactl/cmd/root.go",
		"/github.com/kumahq/kuma/v2@v2.0.0-20251106140736-df9c4e43a672/app/kumactl/cmd/root.go",
	),
	Entry("different org v2 module path",
		"/home/ubuntu/go/pkg/mod/github.com/example/project/v2@v2.1.0/pkg/main.go",
		"/github.com/example/project/v2@v2.1.0/pkg/main.go",
	),
	Entry("v3 module path",
		"/home/ubuntu/go/pkg/mod/github.com/kumahq/kuma/v3@v3.0.0-20251106140736-df9c4e43a672/pkg/test/store/postgres/test_container.go",
		"/github.com/kumahq/kuma/v3@v3.0.0-20251106140736-df9c4e43a672/pkg/test/store/postgres/test_container.go",
	),
)
