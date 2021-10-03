package errors

import (
	"fmt"

	"github.com/emicklei/go-restful"
	"github.com/pkg/errors"

	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/rbac"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
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
	case issuer.IsSigningKeyNotFoundErr(err):
		handleSigningKeyNotFound(err, response)
	case errors.Is(err, &rbac.AccessDeniedError{}):
		var rbacErr *rbac.AccessDeniedError
		errors.As(err, &rbacErr)
		handleRbacAccessDenied(rbacErr, response)
	case errors.Is(err, &Unauthenticated{}):
		var unauthenticated *Unauthenticated
		errors.As(err, &err)
		handleUnauthenticated(unauthenticated, title, response)
	case errors.Is(err, &AccessDenied{}):
		var accessDenied *AccessDenied
		errors.As(err, &err)
		handleAccessDenied(accessDenied, title, response)
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

func handleSigningKeyNotFound(err error, response *restful.Response) {
	kumaErr := types.Error{
		Title:   "Signing Key not found",
		Details: err.Error(),
	}
	writeError(response, 404, kumaErr)
}

func handleRbacAccessDenied(err *rbac.AccessDeniedError, response *restful.Response) {
	kumaErr := types.Error{
		Title:   "Access Denied",
		Details: err.Reason,
	}
	writeError(response, 403, kumaErr)
}

func handleUnauthenticated(err *Unauthenticated, title string, response *restful.Response) {
	kumaErr := types.Error{
		Title:   title,
		Details: err.Error(),
	}
	writeError(response, 401, kumaErr)
}

func handleAccessDenied(err *AccessDenied, title string, response *restful.Response) {
	kumaErr := types.Error{
		Title:   title,
		Details: err.Error(),
	}
	writeError(response, 403, kumaErr)
}

func writeError(response *restful.Response, httpStatus int, kumaErr types.Error) {
	if err := response.WriteHeaderAndJson(httpStatus, kumaErr, "application/json"); err != nil {
		core.Log.Error(err, "Could not write the error response")
	}
}
