package generate

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

type generateDataplaneTokenContext struct {
	*kumactl_cmd.RootContext

	args struct {
		name      string
		proxyType string
		tags      map[string]string
		validFor  time.Duration
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
			token, err := client.Generate(name, pctx.Args.Mesh, tags, ctx.args.proxyType, ctx.args.validFor)
			if err != nil {
				return errors.Wrap(err, "failed to generate a dataplane token")
			}
			_, err = cmd.OutOrStdout().Write([]byte(token))
			return err
		},
	}
	cmd.Flags().StringVar(&ctx.args.name, "name", "", "name of the Dataplane")
	cmd.Flags().StringVar(&ctx.args.proxyType, "type", "", `type of the Dataplane ("dataplane", "ingress")`)
	_ = cmd.Flags().MarkDeprecated("type", "please use --proxy-type instead")
	cmd.Flags().StringVar(&ctx.args.proxyType, "proxy-type", "", `type of the Dataplane ("dataplane", "ingress")`)
	cmd.Flags().StringToStringVar(&ctx.args.tags, "tag", nil, "required tag values for dataplane (split values by comma to provide multiple values)")
	// Backwards compatibility with 1.3.x. Right now we pick 10 years as default, but in the future this should be required argument without default.
	cmd.Flags().DurationVar(&ctx.args.validFor, "valid-for", 24*time.Hour*365*10, `how long the token will be valid (for example "24h")`)
	return cmd
}
