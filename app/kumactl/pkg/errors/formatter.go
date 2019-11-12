package errors

import (
	"fmt"
	"github.com/Kong/kuma/pkg/api-server/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func FormatErrorWrapper(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(otherCmd *cobra.Command, otherArgs []string) error {
		if err := fn(otherCmd, otherArgs); err != nil {
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
