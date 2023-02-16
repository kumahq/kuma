package generate

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	util_rsa "github.com/kumahq/kuma/pkg/util/rsa"
)

type generatePublicKeyArgs struct {
	signingKeyPath string
}

func NewGeneratePublicKeyCmd(ctx *cmd.RootContext) *cobra.Command {
	args := generatePublicKeyArgs{}
	cmd := &cobra.Command{
		Use:   "public-key",
		Short: "Generate public key out of signing key",
		Long:  `Generate a public key for validating tokens.`,
		Example: `
Extract a public key out of signing key used to issue tokens.

$ kumactl generate signing-key --format=pem > /tmp/key.pem
$ kumactl generate public-key --signing-key-path=/tmp/key.pem
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			signingKey, err := os.ReadFile(args.signingKeyPath)
			if err != nil {
				return errors.Wrap(err, "could not read a signing key file")
			}
			if !util_rsa.IsPrivateKeyPEMBytes(signingKey) {
				return errors.New("provided file is not a PEM-encoded signing key")
			}
			key, err := util_rsa.FromPrivateKeyPEMBytesToPublicKeyPEMBytes(signingKey)
			if err != nil {
				return errors.Wrap(err, "could not extract public key")
			}
			_, err = cmd.OutOrStdout().Write(key)
			return err
		},
	}
	cmd.Flags().StringVar(&args.signingKeyPath, "signing-key-path", "", "path to a file with PEM-encoded private signing key")
	_ = cmd.MarkFlagRequired("signing-key-path")
	return cmd
}
