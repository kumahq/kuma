package generate

import (
	"encoding/base64"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

type generateSigningKeyArgs struct {
	format string
}

const (
	SigningKeyFormatPEM       = "pem"
	SigningKeyFormatPEMBase64 = "pem-base64"
)

func NewGenerateSigningKeyCmd(ctx *cmd.RootContext) *cobra.Command {
	args := generateSigningKeyArgs{
		format: SigningKeyFormatPEMBase64,
	}
	cmd := &cobra.Command{
		Use:   "signing-key",
		Short: "Generate signing keys",
		Long:  `Generate a private key for signing tokens.`,
		Example: `
Generate a new signing key to rotate tokens (for example user-token) on Universal.
$ echo "
type: GlobalSecret
name: user-token-signing-key-0002
data: {{ key }}
" | kumactl apply --var key=$(kumactl generate signing-key --format=pem-base64) -f -

Generate a new signing key to rotate tokens (for example user-token) on Kubernetes.
$ TOKEN="$(kumactl generate signing-key --format=pem-base64)" && echo "
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
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			key, err := ctx.GenerateContext.NewSigningKey()
			if err != nil {
				return errors.Wrap(err, "could not generate signing key")
			}
			var out []byte
			switch args.format {
			case SigningKeyFormatPEM:
				out = key
			case SigningKeyFormatPEMBase64:
				out = []byte(base64.StdEncoding.EncodeToString(key))
			default:
				return errors.New("invalid format")
			}
			_, err = cmd.OutOrStdout().Write(out)
			return err
		},
	}
	cmd.Flags().StringVar(&args.format, "format", args.format, fmt.Sprintf("format of signing key. Available values :%s, %s", SigningKeyFormatPEMBase64, SigningKeyFormatPEM))
	return cmd
}
