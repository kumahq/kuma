package ca

import (
	kumactl_cmd "github.com/Kong/kuma/app/kumactl/pkg/cmd"
	"github.com/Kong/kuma/app/kumactl/pkg/output/printers"
	"github.com/Kong/kuma/pkg/core/ca/provided/rest/types"
	"github.com/Kong/kuma/pkg/tls"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
)

func newProvidedCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provided",
		Short: "Manage certificate provided authorities",
		Long:  `Manage certificate provided authorities.`,
	}
	// sub-commands
	cmd.AddCommand(newCertificatesCmd(pctx))
	cmd.AddCommand(newDeleteCaCmd(pctx))
	return cmd
}

func newDeleteCaCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete certificate authority",
		Long:  `Delete certificate authority.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentProvidedCaClient()
			if err != nil {
				return err
			}
			if err := client.DeleteCa(pctx.CurrentMesh()); err != nil {
				return errors.Wrap(err, "could not delete certificate authority")
			}
			return nil
		},
	}
	return cmd
}

func newCertificatesCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certificates",
		Short: "Manage certificates",
		Long:  `Manage certificate.`,
	}
	// sub-commands
	cmd.AddCommand(newAddCertificateCmd(pctx))
	cmd.AddCommand(newListCertificatesCmd(pctx))
	cmd.AddCommand(newDeleteCertificateCmd(pctx))
	return cmd
}

type addCertificateContext struct {
	*kumactl_cmd.RootContext

	args struct {
		keyFile  string
		certFile string
	}
}

func newAddCertificateCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := addCertificateContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add signing certificate",
		Long:  `Add signing certificate.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := ctx.CurrentProvidedCaClient()
			if err != nil {
				return err
			}
			certBytes, err := ioutil.ReadFile(ctx.args.certFile)
			if err != nil {
				return errors.Wrap(err, "could not read content of the certificate file")
			}
			keyBytes, err := ioutil.ReadFile(ctx.args.keyFile)
			if err != nil {
				return errors.Wrap(err, "could not read content of the key file")
			}
			pair := tls.KeyPair{
				CertPEM: certBytes,
				KeyPEM:  keyBytes,
			}
			signingCert, err := client.AddSigningCertificate(ctx.CurrentMesh(), pair)
			if err != nil {
				return err
			}
			cmd.Printf("Signing certificate add. Id: %s", signingCert.Id)
			return nil
		},
	}
	cmd.Flags().StringVar(&ctx.args.keyFile, "key-file", "", "path to a file with a private key")
	_ = cmd.MarkFlagRequired("key-file")
	cmd.Flags().StringVar(&ctx.args.certFile, "cert-file", "", "path to a file with a TLS certificate")
	_ = cmd.MarkFlagRequired("cert-file")
	return cmd
}

func newListCertificatesCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List signing certificates",
		Long:  `List signing certificates.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentProvidedCaClient()
			if err != nil {
				return err
			}
			certs, err := client.SigningCertificates(pctx.CurrentMesh())
			if err != nil {
				return errors.Wrap(err, "could not retrieve signing certificates")
			}
			if err := printListCertificates(certs, cmd.OutOrStdout()); err != nil {
				return errors.Wrap(err, "could not print certificates")
			}
			return nil
		},
	}
	return cmd
}

func printListCertificates(certs []types.SigningCert, out io.Writer) error {
	data := printers.Table{
		Headers: []string{"ID", "CERT"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(certs) <= i {
					return nil
				}
				cert := certs[i]
				return []string{
					cert.Id,   // ID
					cert.Cert, // CERT
				}
			}
		}(),
	}
	return printers.NewTablePrinter().Print(data, out)
}

type deleteCertificateContext struct {
	*kumactl_cmd.RootContext
	args struct {
		id string
	}
}

func newDeleteCertificateCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := deleteCertificateContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete signing certificate",
		Long:  `Delete signing certificate.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentProvidedCaClient()
			if err != nil {
				return err
			}
			if err := client.DeleteSigningCertificate(pctx.CurrentMesh(), ctx.args.id); err != nil {
				return errors.Wrap(err, "could not delete signing certificate")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&ctx.args.id, "id", "", "id of the certificate")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
