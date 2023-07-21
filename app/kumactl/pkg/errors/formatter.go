package errors

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/core/rest/errors/types"
)

func FormatErrorWrapper(fn func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := fn(cmd, args); err != nil {
			cause := errors.Cause(err)
			switch typedErr := cause.(type) {
			case *types.Error:
				return formatApiServerError(typedErr)
			default:
				return err
			}
		}
		return nil
	}
}

func formatApiServerError(apiErr *types.Error) error {
	// Get rid of back compat when we remove `Details` and `Causes`
	if apiErr.Detail == "" {
		apiErr.Detail = apiErr.Details
	}
	msg := fmt.Sprintf("%s (%s)", apiErr.Title, apiErr.Detail)
	if apiErr.Causes != nil && apiErr.InvalidParameters == nil {
		for _, cause := range apiErr.Causes {
			msg += fmt.Sprintf("\n* %s: %s", cause.Field, cause.Message)
		}
	} else {
		for _, cause := range apiErr.InvalidParameters {
			msg += fmt.Sprintf("\n* %s: %s", cause.Field, cause.Reason)
		}
	}
	return errors.New(msg)
}
