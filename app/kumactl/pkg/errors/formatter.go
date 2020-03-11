package errors

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Kong/kuma/pkg/core/rest/errors/types"
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
	msg := fmt.Sprintf("%s (%s)", apiErr.Title, apiErr.Details)
	for _, cause := range apiErr.Causes {
		msg += fmt.Sprintf("\n* %s: %s", cause.Field, cause.Message)
	}
	return errors.New(msg)
}
