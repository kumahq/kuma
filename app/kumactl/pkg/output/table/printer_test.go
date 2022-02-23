package table_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/pkg/output/table"
)

var _ = Describe("printer", func() {

	var printer table.Printer
	var buf *bytes.Buffer

	BeforeEach(func() {
		printer = table.NewPrinter()
		buf = &bytes.Buffer{}
	})

	type testCase struct {
		data       table.Table
		goldenFile string
	}

	DescribeTable("should produce formatted output",
		func(given testCase) {
			// when
			err := printer.Print(given.data, buf)
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			expected, err := os.ReadFile(filepath.Join("testdata", given.goldenFile))
			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(buf.String()).To(Equal(string(expected)))
		},
		Entry("empty Table", testCase{
			data:       table.Table{},
			goldenFile: "empty.golden.txt",
		}),
		Entry("Table with a header but no rows", testCase{
			data: table.Table{
				Headers: []string{"MESH", "NAME"},
			},
			goldenFile: "header.golden.txt",
		}),
		Entry("Table with a header and 1 row", testCase{
			data: table.Table{
				Headers: []string{"MESH", "NAME"},
				NextRow: func() func() []string {
					i := 0
					return func() []string {
						defer func() { i++ }()
						switch i {
						case 0:
							return []string{"default", "example"}
						default:
							return nil
						}
					}
				}(),
			},
			goldenFile: "header-and-1-row.golden.txt",
		}),
		Entry("Table with a header and 2 rows", testCase{
			data: table.Table{
				Headers: []string{"MESH", "NAME"},
				NextRow: func() func() []string {
					i := 0
					return func() []string {
						defer func() { i++ }()
						switch i {
						case 0:
							return []string{"default", "example"}
						case 1:
							return []string{"playground", "demo"}
						default:
							return nil
						}
					}
				}(),
			},
			goldenFile: "header-and-2-rows.golden.txt",
		}),
	)
})
