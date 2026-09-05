package table

import (
	"fmt"

	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

func PaginationFooter(list model.ResourceList) string {
	return PaginationFooterForOffset(list.GetPagination().NextOffset)
}

func PaginationFooterForOffset(offset string) string {
	if offset == "" {
		return ""
	}
	return fmt.Sprintf("Rerun command with --offset=%s argument to retrieve more resources", offset)
}
