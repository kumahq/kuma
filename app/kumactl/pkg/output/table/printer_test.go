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
	var buf *bytes.Buffer

	BeforeEach(func() {
		buf = &bytes.Buffer{}
	})

	type testCase struct {
		data       table.Table
		items      interface{}
		goldenFile string
	}

	DescribeTable("should produce formatted output",
		func(given testCase) {
			// when
			Expect(given.data.Print(given.items, buf)).To(Succeed())

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
			items: [][]string{
				{"default", "example"},
			},
			data: table.Table{
				Headers: []string{"MESH", "NAME"},
				RowForItem: func(i int, container interface{}) ([]string, error) {
					items := container.([][]string)
					if i >= len(items) {
						return nil, nil
					}
					return items[i], nil
				},
			},
			goldenFile: "header-and-1-row.golden.txt",
		}),
		Entry("Table with a header and 2 rows", testCase{
			items: [][]string{
				{"default", "example"},
				{"playground", "demo"},
			},
			data: table.Table{
				Headers: []string{"MESH", "NAME"},
				RowForItem: func(i int, container interface{}) ([]string, error) {
					items := container.([][]string)
					if i >= len(items) {
						return nil, nil
					}
					return items[i], nil
				},
			},
			goldenFile: "header-and-2-rows.golden.txt",
		}),
	)
})
