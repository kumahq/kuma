package table

import (
	"fmt"

	"github.com/Kong/kuma/pkg/core/resources/model"
)

func PaginationFooter(list model.ResourceList) string {
	if list.GetPagination().NextOffset == "" {
		return ""
	}
	return fmt.Sprintf("Rerun command with --offset=%s argument to retrieve more resources", list.GetPagination().NextOffset)
}
