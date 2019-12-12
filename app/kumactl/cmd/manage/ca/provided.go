package ca

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"fmt"
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
		Short: `Manage "provided" certificate authorities`,
		Long:  `Manage "provided" certificate authorities.`,
	}
	// sub-commands
	cmd.AddCommand(newCertificatesCmd(pctx))
	return cmd
}

func newCertificatesCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certificates",
		Short: `Manage signing certificates used by a "provided" certificate authority`,
		Long:  `Manage signing certificates used by a "provided" certificate authority.`,
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
			cmd.Printf("added certificate %q", signingCert.Id)
			return nil
		},
	}
	cmd.Flags().StringVar(&ctx.args.keyFile, "key-file", "", "path to a file with a private key")
	_ = cmd.MarkFlagRequired("key-file")
	cmd.Flags().StringVar(&ctx.args.certFile, "cert-file", "", "path to a file with a CA certificate")
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
	x509Certs := make([]*x509.Certificate, len(certs))
	for i, cert := range certs {
		block, _ := pem.Decode([]byte(cert.Cert))
		x509Cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return errors.Wrap(err, "could not parse certificate")
		}
		x509Certs[i] = x509Cert
	}
	data := printers.Table{
		Headers: []string{"ID", "COMMON NAME", "SERIAL NUMBER", "NOT VALID BEFORE", "NOT VALID AFTER", "SHA-1 FINGERPRINT"},
		NextRow: func() func() []string {
			i := 0
			return func() []string {
				defer func() { i++ }()
				if len(certs) <= i {
					return nil
				}
				cert := certs[i]
				x509Cert := x509Certs[i]
				return []string{
					cert.Id,                                   // ID
					x509Cert.Subject.CommonName,               // COMMON NAME
					x509Cert.SerialNumber.String(),            // SERIAL NUMBER
					x509Cert.NotBefore.String(),               // NOT VALID BEFORE
					x509Cert.NotAfter.String(),                // NOT VALID AFTER
					fmt.Sprintf("%x", sha1.Sum(x509Cert.Raw)), // SHA-1 FINGERPRINT
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
			cmd.Printf("removed certificate %q", ctx.args.id)
			return nil
		},
	}
	cmd.Flags().StringVar(&ctx.args.id, "id", "", "id of the certificate")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}
