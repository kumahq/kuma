package naming

import (
	"fmt"
)

const (
	contextualPrefix = "self"
	inboundNamespace = "inbound"
)

func ContextualInboundName[T ~string | ~uint32](sectionName T) string {
	return fmt.Sprintf("%s_%s_%v", contextualPrefix, inboundNamespace, sectionName)
}
