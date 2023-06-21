package errors

import (
	"fmt"
	"strconv"

	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"

	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/access"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/multitenant"
)

func HandleError(response *restful.Response, err error, title string) {
	switch {
	case store.IsResourceNotFound(err):
		handleNotFound(title, response)
	case store.IsResourcePreconditionFailed(err):
		handlePreconditionFailed(title, response)
	case errors.Is(err, &store.PreconditionError{}):
		var err2 *store.PreconditionError
		errors.As(err, &err2)
		writeError(response, types.Error{
			Status: "400",
			Title:  "Bad Request",
			Detail: err2.Reason,
		})
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
	case tokens.IsSigningKeyNotFound(err):
		handleSigningKeyNotFound(err, response)
	case errors.Is(err, &MethodNotAllowed{}):
		writeError(response, types.Error{
			Status: "405",
			Title:  "Method not Allowed",
			Detail: err.Error(),
		})
	case errors.Is(err, &Conflict{}):
		writeError(response, types.Error{
			Status: "409",
			Title:  "Conflict",
			Detail: err.Error(),
		})

	case errors.Is(err, &access.AccessDeniedError{}):
		var accessErr *access.AccessDeniedError
		errors.As(err, &accessErr)
		handleAccessDenied(accessErr, response)
	case errors.Is(err, &Unauthenticated{}):
		var unauthenticated *Unauthenticated
		errors.As(err, &err)
		handleUnauthenticated(unauthenticated, title, response)
	case err == tokens.IssuerDisabled:
		handleIssuerDisabled(err, title, response)
	case err == multitenant.TenantMissingErr:
		handleTenantMissing(err, title, response)
	default:
		handleUnknownError(err, title, response)
	}
}

func handleIssuerDisabled(err error, title string, response *restful.Response) {
	kumaErr := types.Error{
		Status: "400",
		Title:  title,
		Detail: err.Error(),
	}
	writeError(response, kumaErr)
}

func handleInvalidPageSize(title string, response *restful.Response) {
	kumaErr := types.Error{
		Status: "400",
		Title:  title,
		Detail: "Invalid page size",
		InvalidParameters: []types.InvalidParameter{
			{
				Field:  "size",
				Reason: "Invalid format",
			},
		},
	}
	writeError(response, kumaErr)
}

func handleNotFound(title string, response *restful.Response) {
	kumaErr := types.Error{
		Status: "404",
		Title:  title,
		Detail: "Not found",
	}
	writeError(response, kumaErr)
}

func handlePreconditionFailed(title string, response *restful.Response) {
	kumaErr := types.Error{
		Status: "412",
		Title:  title,
		Detail: "Precondition Failed",
	}
	writeError(response, kumaErr)
}

func handleMeshNotFound(title string, err *manager.MeshNotFoundError, response *restful.Response) {
	kumaErr := types.Error{
		Status: "400",
		Title:  title,
		Detail: "Mesh is not found",
		InvalidParameters: []types.InvalidParameter{
			{
				Field:  "mesh",
				Reason: fmt.Sprintf("mesh of name %s is not found", err.Mesh),
			},
		},
	}
	writeError(response, kumaErr)
}

func handleValidationError(title string, err *validators.ValidationError, response *restful.Response) {
	kumaErr := types.Error{
		Status: "400",
		Title:  title,
		Detail: "Resource is not valid",
	}
	for _, violation := range err.Violations {
		kumaErr.InvalidParameters = append(kumaErr.InvalidParameters, types.InvalidParameter{
			Field:  violation.Field,
			Reason: violation.Message,
		})
	}
	writeError(response, kumaErr)
}

func handleInvalidOffset(title string, response *restful.Response) {
	kumaErr := types.Error{
		Status: "400",
		Title:  title,
		Detail: "Invalid offset",
		InvalidParameters: []types.InvalidParameter{
			{
				Field:  "offset",
				Reason: "Invalid format",
			},
		},
	}
	writeError(response, kumaErr)
}

func handleMaxPageSizeExceeded(title string, err error, response *restful.Response) {
	kumaErr := types.Error{
		Status: "400",
		Title:  title,
		Detail: "Invalid page size",
		InvalidParameters: []types.InvalidParameter{
			{
				Field:  "size",
				Reason: err.Error(),
			},
		},
	}
	writeError(response, kumaErr)
}

func handleUnknownError(err error, title string, response *restful.Response) {
	core.Log.Error(err, title)
	kumaErr := types.Error{
		Status: "500",
		Title:  title,
		Detail: "Internal Server Error",
	}
	writeError(response, kumaErr)
}

func handleSigningKeyNotFound(err error, response *restful.Response) {
	kumaErr := types.Error{
		Status: "404",
		Title:  "Signing Key not found",
		Detail: err.Error(),
	}
	writeError(response, kumaErr)
}

func handleAccessDenied(err *access.AccessDeniedError, response *restful.Response) {
	kumaErr := types.Error{
		Status: "403",
		Title:  "Access Denied",
		Detail: err.Reason,
	}
	writeError(response, kumaErr)
}

func handleUnauthenticated(err *Unauthenticated, title string, response *restful.Response) {
	kumaErr := types.Error{
		Status: "401",
		Title:  title,
		Detail: err.Error(),
	}
	writeError(response, kumaErr)
}

func handleTenantMissing(err error, title string, response *restful.Response) {
	kumaErr := types.Error{
		Status: "400",
		Title:  title,
		Detail: err.Error(),
	}
	writeError(response, kumaErr)
}

func writeError(response *restful.Response, kumaErr types.Error) {
	// Fix to handle legacy errors
	kumaErr.Type = "/std-errors"
	kumaErr.Details = kumaErr.Detail
	for _, ip := range kumaErr.InvalidParameters {
		kumaErr.Causes = append(kumaErr.Causes, types.Cause{Field: ip.Field, Message: ip.Reason})
	}
	httpStatusCode, err := strconv.Atoi(kumaErr.Status)
	if err != nil {
		httpStatusCode = 500
	}
	if err := response.WriteHeaderAndJson(httpStatusCode, kumaErr, "application/json"); err != nil {
		core.Log.Error(err, "Could not write the error response")
	}
}
