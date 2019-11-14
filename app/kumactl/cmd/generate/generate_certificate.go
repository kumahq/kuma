package generate

import (
	"fmt"
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	cmd2 "github.com/Kong/kuma/pkg/cmd"
	"github.com/Kong/kuma/pkg/tls"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var (
	// overridable by unit tests
	NewSelfSignedCert = tls.NewSelfSignedCert
)

type generateCertificateContext struct {
	*kumactl_cmd.RootContext

	args struct {
		key      string
		cert     string
		certType string
	}
}

func NewGenerateCertificateCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &generateCertificateContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "certificate",
		Short: "Generate certificate",
		Long:  `Generate self signed key and certificate pair that can be used for example in Dataplane Token Server setup.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if ctx.args.certType != "client" && ctx.args.certType != "server" {
				return errors.New(`--type has to be either "client" or "server"`)
			}
			keyPair, err := NewSelfSignedCert("kuma", tls.CertType(ctx.args.certType))
			if err != nil {
				return errors.Wrap(err, "could not generate certificate")
			}
			if err := ioutil.WriteFile(ctx.args.key, keyPair.KeyPEM, 0664); err != nil {
				return errors.Wrap(err, "could not write the key file")
			}
			if err := ioutil.WriteFile(ctx.args.cert, keyPair.CertPEM, 0664); err != nil {
				return errors.Wrap(err, "could not write the cert file")
			}
			_, err = cmd.OutOrStdout().Write([]byte(fmt.Sprintf(`Certificates generated
Key path: %s
Cert path: %s
`, ctx.args.key, ctx.args.cert)))
			return err
		},
	}
	cmd.Flags().StringVar(&ctx.args.key, "key", "key.pem", "path to the generated key")
	cmd.Flags().StringVar(&ctx.args.cert, "cert", "cert.pem", "path to the generated certificate")
	cmd.Flags().StringVar(&ctx.args.certType, "type", "", cmd2.UsageOptions("type of the certificate", "client", "server"))
	_ = cmd.MarkFlagRequired("type")
	return cmd
}
