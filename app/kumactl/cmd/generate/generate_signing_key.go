package generate

import (
	"encoding/base64"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewGenerateSigningKeyCmd(ctx *cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signing-key",
		Short: "Generate Signing Key",
		Long:  `Generate Signing Key that is used as a private key for signing tokens.`,
		Example: `
Generate a new Signing Key to rotate tokens (for example user-token) on Universal.
$ echo "
type: GlobalSecret
name: user-token-signing-key-0002
data: {{ key }}
" | kumactl apply --var key=$(kumactl generate signing-key) -f -

Generate a new Signing Key to rotate tokens (for example user-token) on Kubernetes.
$ TOKEN="$(kumactl generate signing-key)" && echo "
apiVersion: v1
data:
  value: $TOKEN
kind: Secret
metadata:
  name: user-token-signing-key-0002
  namespace: kong-mesh-system
type: system.kuma.io/global-secret
" | kubectl apply -f - 
`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			key, err := ctx.GenerateContext.NewSigningKey()
			if err != nil {
				return errors.Wrap(err, "could not generate signing key")
			}
			base64.StdEncoding.EncodeToString(key)
			_, err = cmd.OutOrStdout().Write([]byte(base64.StdEncoding.EncodeToString(key)))
			return err
		},
	}
	return cmd
}
