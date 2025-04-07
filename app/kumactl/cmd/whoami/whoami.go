package whoami

import (
	"encoding/json"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

func NewWhoAmICmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "who-am-i",
		Short: "Print user information",
		Long:  `Print user information including name and groups.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			userInfo, err := pctx.FetchUserInfo()
			if userInfo != nil {
				s, err := json.MarshalIndent(userInfo, "", "   ")
				if err != nil {
					return err
				}
				cmd.Printf("User: %s\n", string(s))
			} else {
				cmd.PrintErrf("Unable to connect to control plane: %v\n", err)
			}

			return nil
		},
	}

	return cmd
}
