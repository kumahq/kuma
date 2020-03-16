package generate

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	kuma_cmd "github.com/Kong/kuma/pkg/cmd"
	"github.com/Kong/kuma/pkg/tls"
)

var (
	// overridable by unit tests
	NewSelfSignedCert = tls.NewSelfSignedCert
)

type generateCertificateContext struct {
	*kumactl_cmd.RootContext

	args struct {
		key                  string
		cert                 string
		certType             string
		controlPlaneHostname []string
	}
}

func NewGenerateCertificateCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &generateCertificateContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "tls-certificate",
		Short: "Generate a TLS certificate",
		Long:  `Generate self signed key and certificate pair that can be used for example in Dataplane Token Server setup.`,
		Example: `
  # Generate a TLS certificate for use by an HTTPS server, i.e. by the Dataplane Token server
  kumactl generate tls-certificate --type=server

  # Generate a TLS certificate for use by a client of an HTTPS server, i.e. by the 'kumactl generate dataplane-token' command
  kumactl generate tls-certificate --type=client`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if ctx.args.certType != "client" && ctx.args.certType != "server" {
				return errors.New(`--type has to be either "client" or "server"`)
			}
			certType := tls.CertType(ctx.args.certType)
			if certType == tls.ClientCertType && len(ctx.args.controlPlaneHostname) != 0 {
				return errors.New(`--cp-hostname cannot be used with "client" type`)
			}
			if certType == tls.ServerCertType && len(ctx.args.controlPlaneHostname) < 1 {
				return errors.New(`--cp-hostname has to be specified with "server" type`)
			}

			keyPair, err := NewSelfSignedCert("kuma", certType, append(ctx.args.controlPlaneHostname, "localhost")...)
			if err != nil {
				return errors.Wrap(err, "could not generate certificate")
			}
			if err := ioutil.WriteFile(ctx.args.key, keyPair.KeyPEM, 0400); err != nil {
				return errors.Wrap(err, "could not write the key file")
			}
			if err := ioutil.WriteFile(ctx.args.cert, keyPair.CertPEM, 0644); err != nil {
				return errors.Wrap(err, "could not write the cert file")
			}
			_, err = cmd.OutOrStdout().Write([]byte(fmt.Sprintf(`Certificates generated
Key was saved in: %s
Cert was saved in: %s
`, ctx.args.key, ctx.args.cert)))
			return err
		},
	}
	cmd.Flags().StringVar(&ctx.args.key, "key-file", "key.pem", "path to a file with a generated private key")
	cmd.Flags().StringVar(&ctx.args.cert, "cert-file", "cert.pem", "path to a file with a generated TLS certificate")
	cmd.Flags().StringVar(&ctx.args.certType, "type", "", kuma_cmd.UsageOptions("type of the certificate", "client", "server"))
	cmd.Flags().StringSliceVar(&ctx.args.controlPlaneHostname, "cp-hostname", []string{}, "DNS name of the control plane")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}
