package get

import (
	"github.com/spf13/cobra"
)

func newGetEntitiesCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}
