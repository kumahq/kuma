package errors

import (
	"fmt"

	"github.com/emicklei/go-restful"

	api_server_types "github.com/Kong/kuma/pkg/api-server/types"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/rest/errors/types"
	"github.com/Kong/kuma/pkg/core/validators"
)

func HandleError(response *restful.Response, err error, title string) {
	switch {
	case store.IsResourceNotFound(err):
		handleNotFound(title, response)
	case store.IsResourcePreconditionFailed(err):
		handlePreconditionFailed(title, response)
	case err == store.ErrorInvalidOffset:
		handleInvalidOffset(title, response)
	case manager.IsMeshNotFound(err):
		handleMeshNotFound(title, err.(*manager.MeshNotFoundError), response)
	case validators.IsValidationError(err):
		handleValidationError(title, err.(*validators.ValidationError), response)
	case api_server_types.IsMaxPageSizeExceeded(err):
		handleMaxPageSizeExceeded(title, err, response)
	case err == api_server_types.InvalidPageSize:
		handleInvalidPageSize(title, response)
	case err == api_server_types.PaginationNotSupported:
		handlePaginationNotSupported(title, response)
	default:
		handleUnknownError(err, title, response)
	}
}

func handleInvalidPageSize(title string, response *restful.Response) {
	kumaErr := types.Error{
		Title:   title,
		Details: "Invalid page size",
		Causes: []types.Cause{
			{
				Field:   "size",
				Message: "Invalid format",
			},
		},
	}
	writeError(response, 400, kumaErr)
}

func handleNotFound(title string, response *restful.Response) {
	kumaErr := types.Error{
		Title:   title,
		Details: "Not found",
	}
	writeError(response, 404, kumaErr)
}

func handlePreconditionFailed(title string, response *restful.Response) {
	kumaErr := types.Error{
		Title:   title,
		Details: "Precondition Failed",
	}
	writeError(response, 412, kumaErr)
}

func handleMeshNotFound(title string, err *manager.MeshNotFoundError, response *restful.Response) {
	kumaErr := types.Error{
		Title:   title,
		Details: "Mesh is not found",
		Causes: []types.Cause{
			{
				Field:   "mesh",
				Message: fmt.Sprintf("mesh of name %s is not found", err.Mesh),
			},
		},
	}
	writeError(response, 400, kumaErr)
}

func handleValidationError(title string, err *validators.ValidationError, response *restful.Response) {
	kumaErr := types.Error{
		Title:   title,
		Details: "Resource is not valid",
	}
	for _, violation := range err.Violations {
		kumaErr.Causes = append(kumaErr.Causes, types.Cause{
			Field:   violation.Field,
			Message: violation.Message,
		})
	}
	writeError(response, 400, kumaErr)
}

func handleInvalidOffset(title string, response *restful.Response) {
	kumaErr := types.Error{
		Title:   title,
		Details: "Invalid offset",
		Causes: []types.Cause{
			{
				Field:   "offset",
				Message: "Invalid format",
			},
		},
	}
	writeError(response, 400, kumaErr)
}

func handlePaginationNotSupported(title string, response *restful.Response) {
	kumaErr := types.Error{
		Title:   title,
		Details: api_server_types.PaginationNotSupported.Error(),
	}
	writeError(response, 400, kumaErr)
}

func handleMaxPageSizeExceeded(title string, err error, response *restful.Response) {
	kumaErr := types.Error{
		Title:   title,
		Details: "Invalid page size",
		Causes: []types.Cause{
			{
				Field:   "size",
				Message: err.Error(),
			},
		},
	}
	writeError(response, 400, kumaErr)
}

func handleUnknownError(err error, title string, response *restful.Response) {
	core.Log.Error(err, title)
	kumaErr := types.Error{
		Title:   title,
		Details: "Internal Server Error",
	}
	writeError(response, 500, kumaErr)
}

func writeError(response *restful.Response, httpStatus int, kumaErr types.Error) {
	if err := response.WriteHeaderAndJson(httpStatus, kumaErr, "application/json"); err != nil {
		core.Log.Error(err, "Could not write the error response")
	}
}
