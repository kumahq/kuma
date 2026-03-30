package output

import (
	"io"
)

type Printer interface {
	Print(any, io.Writer) error
}
