package table_test

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/output/table"
)

var _ = Describe("Printer", func() {

	var printer table.Printer
	var buf *bytes.Buffer

	BeforeEach(func() {
		printer = table.NewPrinter()
		buf = &bytes.Buffer{}
	})

	It("should not fail on empty Table object", func() {
		// given
		data := table.Table{}

		// when
		err := printer.Print(data, buf)

		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(buf.String()).To(Equal(``))
	})

	It("should support Table with no rows", func() {
		// given
		data := table.Table{
			Headers: []string{"NAMESPACE", "NAME"},
		}

		// when
		err := printer.Print(data, buf)

		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(strings.TrimSpace(buf.String())).To(Equal(strings.TrimSpace(`
NAMESPACE   NAME
`)))
	})

	It("should support Table with 1 row", func() {
		// given
		data := table.Table{
			Headers: []string{"NAMESPACE", "NAME"},
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
		}

		// when
		err := printer.Print(data, buf)

		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(strings.TrimSpace(buf.String())).To(Equal(strings.TrimSpace(`
NAMESPACE   NAME
default     example
`)))
	})

	It("should support Table with 2 rows", func() {
		// given
		data := table.Table{
			Headers: []string{"NAMESPACE", "NAME"},
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
		}

		// when
		err := printer.Print(data, buf)

		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(strings.TrimSpace(buf.String())).To(Equal(strings.TrimSpace(`
NAMESPACE    NAME
default      example
playground   demo
`)))
	})
})
