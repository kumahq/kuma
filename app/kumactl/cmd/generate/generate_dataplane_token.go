package generate

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/tokens/builtin/issuer"
)

type generateDataplaneTokenContext struct {
	*kumactl_cmd.RootContext

	args struct {
		name           string
		proxyType      string
		tags           map[string]string
		validFor       time.Duration
		kid            string
		signingKeyPath string
	}
}

func NewGenerateDataplaneTokenCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	ctx := &generateDataplaneTokenContext{RootContext: pctx}
	cmd := &cobra.Command{
		Use:   "dataplane-token",
		Short: "Generate Dataplane Token",
		Long:  `Generate Dataplane Token that is used to prove Dataplane identity.`,
		Example: `
Generate token bound by name and mesh
$ kumactl generate dataplane-token --mesh demo --name demo-01 --valid-for 24h

Generate token bound by mesh
$ kumactl generate dataplane-token --mesh demo --valid-for 24h

Generate Ingress token
$ kumactl generate dataplane-token --type ingress --valid-for 24h

Generate token bound by tag
$ kumactl generate dataplane-token --mesh demo --tag kuma.io/service=web,web-api --valid-for 24h
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := pctx.CurrentDataplaneTokenClient()
			if err != nil {
				return errors.Wrap(err, "failed to create dataplane token client")
			}

			tags := map[string][]string{}
			for k, v := range ctx.args.tags {
				tags[k] = strings.Split(v, ",")
			}
			name := ctx.args.name

			var token string

			if ctx.args.signingKeyPath != "" {
				if ctx.args.kid == "" {
					return errors.New("--kid is required when --signing-key-path is used")
				}
				dpTokenIssuer := issuer.NewDataplaneTokenIssuer(func(_ string) tokens.Issuer {
					return tokens.NewTokenIssuer(tokens.NewFileSigningKeyManager(ctx.args.signingKeyPath, ctx.args.kid))
				})
				token, err = dpTokenIssuer.Generate(cmd.Context(), issuer.DataplaneIdentity{
					Name: name,
					Mesh: pctx.CurrentMesh(),
					Tags: mesh_proto.MultiValueTagSetFrom(tags),
					Type: mesh_proto.ProxyType(ctx.args.proxyType),
				}, ctx.args.validFor)
				if err != nil {
					return err
				}
			} else {
				if ctx.args.kid != "" {
					return errors.New("--kid cannot be used when --signing-key-path is used")
				}
				token, err = client.Generate(name, pctx.CurrentMesh(), tags, ctx.args.proxyType, ctx.args.validFor)
				if err != nil {
					return errors.Wrap(err, "failed to generate a dataplane token")
				}
			}

			_, err = cmd.OutOrStdout().Write([]byte(token))
			return err
		},
	}
	cmd.Flags().StringVar(&ctx.args.name, "name", "", "name of the Dataplane")
	cmd.PersistentFlags().StringVarP(&pctx.Args.Mesh, "mesh", "m", "default", "mesh to use")
	cmd.Flags().StringVar(&ctx.args.proxyType, "type", "", `type of the Dataplane ("dataplane", "ingress")`)
	_ = cmd.Flags().MarkDeprecated("type", "please use --proxy-type instead")
	cmd.Flags().StringVar(&ctx.args.proxyType, "proxy-type", "", `type of the Dataplane ("dataplane", "ingress")`)
	cmd.Flags().StringToStringVar(&ctx.args.tags, "tag", nil, "required tag values for dataplane (split values by comma to provide multiple values)")
	cmd.Flags().DurationVar(&ctx.args.validFor, "valid-for", 0, `how long the token will be valid (for example "24h")`)
	cmd.Flags().StringVar(&ctx.args.signingKeyPath, "signing-key-path", "", "path to a file that contains private signing key. When specified, control plane won't be used to issue the token.")
	cmd.Flags().StringVar(&ctx.args.kid, "kid", "", "ID of the key that is used to issue a token. Required when --signing-key-path is used.")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("valid-for")
	return cmd
}
