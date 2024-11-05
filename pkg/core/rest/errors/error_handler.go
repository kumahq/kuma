package errors

import (
	"context"
	"errors"
	"fmt"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel/trace"

	api_server_types "github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/core/access"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/rest/errors/types"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/intercp/envoyadmin"
	kds_envoyadmin "github.com/kumahq/kuma/pkg/kds/envoyadmin"
	kuma_log "github.com/kumahq/kuma/pkg/log"
	"github.com/kumahq/kuma/pkg/multitenant"
)

func HandleError(ctx context.Context, response *restful.Response, err error, title string) {
	log := kuma_log.AddFieldsFromCtx(logr.FromContextOrDiscard(ctx), ctx, context.Background())

	var kumaErr *types.Error
	switch {
	case store.IsResourceNotFound(err) || errors.Is(err, &NotFound{}):
		kumaErr = &types.Error{
			Status: 404,
			Title:  title,
			Detail: "Not found",
		}
	case errors.Is(err, &rest.InvalidResourceError{}), errors.Is(err, &registry.InvalidResourceTypeError{}), errors.Is(err, &store.PreconditionError{}), errors.Is(err, &BadRequest{}):
		kumaErr = &types.Error{
			Status: 400,
			Title:  "Bad Request",
			Detail: err.Error(),
		}
	case manager.IsMeshNotFound(err):
		kumaErr = &types.Error{
			Status: 400,
			Title:  title,
			Detail: "Mesh is not found",
			InvalidParameters: []types.InvalidParameter{
				{
					Field:  "mesh",
					Reason: fmt.Sprintf("mesh of name %s is not found", err.(*manager.MeshNotFoundError).Mesh),
				},
			},
		}
	case validators.IsValidationError(err):
		kumaErr = &types.Error{
			Status: 400,
			Title:  title,
			Detail: "Resource is not valid",
		}
		for _, violation := range err.(*validators.ValidationError).Violations {
			kumaErr.InvalidParameters = append(kumaErr.InvalidParameters, types.InvalidParameter{
				Field:  violation.Field,
				Reason: violation.Message,
			})
		}
	case err == store.ErrorInvalidOffset:
		kumaErr = &types.Error{
			Status: 400,
			Title:  title,
			Detail: "Invalid offset",
			InvalidParameters: []types.InvalidParameter{
				{
					Field:  "offset",
					Reason: "Invalid format",
				},
			},
		}
	case errors.Is(err, &api_server_types.InvalidPageSizeError{}):
		kumaErr = &types.Error{
			Status: 400,
			Title:  title,
			Detail: "Invalid page size",
			InvalidParameters: []types.InvalidParameter{
				{
					Field:  "size",
					Reason: err.Error(),
				},
			},
		}
	case tokens.IsSigningKeyNotFound(err):
		kumaErr = &types.Error{
			Status: 404,
			Title:  "Signing Key not found",
			Detail: err.Error(),
		}
	case errors.Is(err, &MethodNotAllowed{}):
		kumaErr = &types.Error{
			Status: 405,
			Title:  "Method not Allowed",
			Detail: err.Error(),
		}
	case errors.Is(err, &Conflict{}) || errors.Is(err, &store.ResourceConflictError{}):
		kumaErr = &types.Error{
			Status: 409,
			Title:  "Conflict",
			Detail: err.Error(),
		}
	case errors.Is(err, &ServiceUnavailable{}):
		kumaErr = &types.Error{
			Status: 503,
			Title:  "Service unavailable",
			Detail: err.Error(),
		}
	case errors.Is(err, &access.AccessDeniedError{}):
		var accessErr *access.AccessDeniedError
		errors.As(err, &accessErr)
		kumaErr = &types.Error{
			Status: 403,
			Title:  "Access Denied",
			Detail: accessErr.Reason,
		}
	case errors.Is(err, &Unauthenticated{}):
		var unauthenticated *Unauthenticated
		errors.As(err, &unauthenticated)
		kumaErr = &types.Error{
			Status: 401,
			Title:  title,
			Detail: unauthenticated.Error(),
		}
	case errors.Is(err, &envoyadmin.StreamNotConnectedError{}):
		kumaErr = &types.Error{
			Status: 400,
			Title:  "Bad Request",
			Detail: err.Error(),
		}
	case err == tokens.IssuerDisabled:
		kumaErr = &types.Error{
			Status: 400,
			Title:  title,
			Detail: err.Error(),
		}
	case err == multitenant.TenantMissingErr:
		kumaErr = &types.Error{
			Status: 400,
			Title:  title,
			Detail: err.Error(),
		}
	case errors.Is(err, &kds_envoyadmin.KDSTransportError{}), errors.Is(err, &envoyadmin.ForwardKDSRequestError{}), errors.Is(err, &envoyadmin.ZoneOfflineError{}):
		kumaErr = &types.Error{
			Status: 400,
			Title:  title,
			Detail: err.Error(),
		}
	case errors.Is(err, context.Canceled):
		kumaErr = &types.Error{
			Status: 499,
			Title:  title,
			Detail: err.Error(),
		}
	case errors.As(err, &kumaErr):
	default:
		log.Error(err, title)
		kumaErr = &types.Error{
			Status: 500,
			Title:  title,
			Detail: "Internal Server Error",
		}
	}
	if ctx != nil {
		span := trace.SpanFromContext(ctx)
		span.RecordError(err, trace.WithStackTrace(true))
		if span.IsRecording() {
			kumaErr.Instance = span.SpanContext().TraceID().String()
		}
	}
	// Fix to handle legacy errors
	kumaErr.Type = "/std-errors"
	kumaErr.Details = kumaErr.Detail
	for _, ip := range kumaErr.InvalidParameters {
		kumaErr.Causes = append(kumaErr.Causes, types.Cause{Field: ip.Field, Message: ip.Reason})
	}
	if err := response.WriteHeaderAndJson(kumaErr.Status, kumaErr, "application/json"); err != nil {
		log.Error(err, "Could not write the error response")
	}
}
