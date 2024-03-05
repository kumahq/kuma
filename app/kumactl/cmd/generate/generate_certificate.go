package generate

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/tls"
)

// overridable by unit tests
var NewSelfSignedCert = tls.NewSelfSignedCert

type generateCertificateContext struct {
	*kumactl_cmd.RootContext

	args struct {
		key       string
		cert      string
		certType  string
		keyType   string
		hostnames []string
	}
}

func NewGenerateCertificateCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &generateCertificateContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "tls-certificate --type=server|client --hostname=HOST1[,HOST2...]",
		Short: "Generate a TLS certificate",
		Long:  `Generate self signed key and certificate pair that can be used for example in Dataplane Token Server setup.`,
		Example: `
  # Generate a TLS certificate for use by an HTTPS server, i.e. by the Dataplane Token server
  kumactl generate tls-certificate --type=server --hostname=localhost

  # Generate a TLS certificate for use by a client of an HTTPS server, i.e. by the 'kumactl generate dataplane-token' command
  kumactl generate tls-certificate --type=client --hostname=dataplane-1`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			certType := tls.CertType(ctx.args.certType)
			switch certType {
			case tls.ClientCertType, tls.ServerCertType:
				if len(ctx.args.hostnames) == 0 {
					return errors.New("at least one hostname must be given")
				}
			default:
				return errors.Errorf("invalid certificate type %q", certType)
			}

			keyType := tls.DefaultKeyType
			switch ctx.args.keyType {
			case "":
			case "rsa":
				keyType = tls.RSAKeyType
			case "ecdsa":
				keyType = tls.ECDSAKeyType
			default:
				return errors.Errorf("invalid key type %q", ctx.args.keyType)
			}

			keyPair, err := NewSelfSignedCert(certType, keyType, ctx.args.hostnames...)
			if err != nil {
				return errors.Wrap(err, "could not generate certificate")
			}

			if ctx.args.key == "-" {
				_, err = cmd.OutOrStdout().Write(keyPair.KeyPEM)
			} else {
				err = os.WriteFile(ctx.args.key, keyPair.KeyPEM, 0o400)
			}
			if err != nil {
				return errors.Wrap(err, "could not write the key file")
			}

			if ctx.args.cert == "-" {
				_, err = cmd.OutOrStdout().Write(keyPair.CertPEM)
			} else {
				err = os.WriteFile(ctx.args.cert, keyPair.CertPEM, 0o600)
			}
			if err != nil {
				return errors.Wrap(err, "could not write the cert file")
			}

			if ctx.args.cert != "-" && ctx.args.key != "-" {
				fmt.Fprintf(cmd.OutOrStdout(), "Private key saved in %s\n", ctx.args.key)
				fmt.Fprintf(cmd.OutOrStdout(), "Certificate saved in %s\n", ctx.args.cert)
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&ctx.args.key, "key-file", "key.pem", "path to a file with a generated private key ('-' for stdout)")
	cmd.Flags().StringVar(&ctx.args.cert, "cert-file", "cert.pem", "path to a file with a generated TLS certificate ('-' for stdout)")
	cmd.Flags().StringVar(&ctx.args.certType, "type", "", kuma_cmd.UsageOptions("type of the certificate", "client", "server"))
	cmd.Flags().StringVar(&ctx.args.keyType, "key-type", "", kuma_cmd.UsageOptions("type of the private key", "rsa", "ecdsa"))
	cmd.Flags().StringSliceVar(&ctx.args.hostnames, "hostname", []string{}, "DNS hostname(s) to issue the certificate for")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("hostname")

	return cmd
}
