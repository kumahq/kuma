package output

import (
	"io"
)

type Printer interface {
	Print(interface{}, io.Writer) error
}
