// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/miekg/dns"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"istio.io/pkg/env"

	"github.com/kumahq/kuma/pkg/transparentproxy/istio/tools/istio-clean-iptables/pkg/config"
	common "github.com/kumahq/kuma/pkg/transparentproxy/istio/tools/istio-iptables/pkg/cmd"
	"github.com/kumahq/kuma/pkg/transparentproxy/istio/tools/istio-iptables/pkg/constants"
)

var (
	envoyUserVar = env.RegisterStringVar(constants.EnvoyUser, "istio-proxy", "Envoy proxy username")
	// Enable interception of DNS.
	dnsCaptureByAgent = env.RegisterBoolVar("ISTIO_META_DNS_CAPTURE", false,
		"If set to true, enable the capture of outgoing DNS packets on port 53, redirecting to istio-agent on :15053").Get()
)

var rootCmd = &cobra.Command{
	Use:    "istio-clean-iptables",
	Short:  "Clean up iptables rules for Istio Sidecar",
	Long:   "Script responsible for cleaning up iptables rules",
	PreRun: bindFlags,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := constructConfig()
		cleanup(cfg)
	},
}

func constructConfig() *config.Config {
	cfg := &config.Config{
		DryRun:                    viper.GetBool(constants.DryRun),
		ProxyUID:                  viper.GetString(constants.ProxyUID),
		ProxyGID:                  viper.GetString(constants.ProxyGID),
		RedirectDNS:               viper.GetBool(constants.RedirectDNS),
		RedirectAllDNSTraffic:     viper.GetBool(constants.RedirectAllDNSTraffic),
		AgentDNSListenerPort:      viper.GetString(constants.AgentDNSListenerPort),
		DNSUpstreamTargetChain:    viper.GetString(constants.DNSUpstreamTargetChain),
		SkipDNSConntrackZoneSplit: viper.GetBool(constants.SkipDNSConntrackZoneSplit),
	}

	// TODO: Make this more configurable, maybe with an allowlist of users to be captured for output instead of a denylist.
	if cfg.ProxyUID == "" {
		usr, err := user.Lookup(envoyUserVar.Get())
		var userID string
		// Default to the UID of ENVOY_USER
		if err != nil {
			userID = constants.DefaultProxyUID
		} else {
			userID = usr.Uid
		}
		cfg.ProxyUID = userID
	}
	// For TPROXY as its uid and gid are same.
	if cfg.ProxyGID == "" {
		cfg.ProxyGID = cfg.ProxyUID
	}

	// Lookup DNS nameservers. We only do this if DNS is enabled in case of some obscure theoretical
	// case where reading /etc/resolv.conf could fail.
	if cfg.RedirectDNS {
		if cfg.RedirectAllDNSTraffic {
			cfg.DNSServersV4, cfg.DNSServersV6 = []string{"0.0.0.0"}, []string{"::"}
		} else {
			dnsConfig, err := dns.ClientConfigFromFile("/etc/resolv.conf")
			if err != nil {
				panic(fmt.Sprintf("failed to load /etc/resolv.conf: %v", err))
			}
			cfg.DNSServersV4, cfg.DNSServersV6 = common.SplitV4V6(dnsConfig.Servers)
		}
	}

	return cfg
}

// https://github.com/spf13/viper/issues/233.
// Any viper mutation and binding should be placed in `PreRun` since they should be dynamically bound to the subcommand being executed.
func bindFlags(cmd *cobra.Command, args []string) {
	// Read in all environment variables
	viper.AutomaticEnv()
	// Replace - with _; so that environment variables are looked up correctly.
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	if err := viper.BindPFlag(constants.DryRun, cmd.Flags().Lookup(constants.DryRun)); err != nil {
		handleError(err)
	}
	viper.SetDefault(constants.DryRun, false)

	if err := viper.BindPFlag(constants.ProxyUID, cmd.Flags().Lookup(constants.ProxyUID)); err != nil {
		handleError(err)
	}
	viper.SetDefault(constants.ProxyUID, "")

	if err := viper.BindPFlag(constants.ProxyGID, cmd.Flags().Lookup(constants.ProxyGID)); err != nil {
		handleError(err)
	}
	viper.SetDefault(constants.ProxyGID, "")

	if err := viper.BindPFlag(constants.RedirectDNS, cmd.Flags().Lookup(constants.RedirectDNS)); err != nil {
		handleError(err)
	}
	viper.SetDefault(constants.RedirectDNS, dnsCaptureByAgent)

	if err := viper.BindPFlag(constants.RedirectAllDNSTraffic, cmd.Flags().Lookup(constants.RedirectAllDNSTraffic)); err != nil {
		handleError(err)
	}
	viper.SetDefault(constants.RedirectAllDNSTraffic, false)

	if err := viper.BindPFlag(constants.AgentDNSListenerPort, cmd.Flags().Lookup(constants.AgentDNSListenerPort)); err != nil {
		handleError(err)
	}
	viper.SetDefault(constants.AgentDNSListenerPort, constants.IstioAgentDNSListenerPort)

	if err := viper.BindPFlag(constants.DNSUpstreamTargetChain, cmd.Flags().Lookup(constants.DNSUpstreamTargetChain)); err != nil {
		handleError(err)
	}
	viper.SetDefault(constants.DNSUpstreamTargetChain, constants.RETURN)

	if err := viper.BindPFlag(constants.SkipDNSConntrackZoneSplit, cmd.Flags().Lookup(constants.SkipDNSConntrackZoneSplit)); err != nil {
		handleError(err)
	}
	viper.SetDefault(constants.SkipDNSConntrackZoneSplit, false)
}

// https://github.com/spf13/viper/issues/233.
// Only adding flags in `init()` while moving its binding to Viper and value defaulting as part of the command execution.
// Otherwise, the flag with the same name shared across subcommands will be overwritten by the last.
func init() {
	rootCmd.Flags().BoolP(constants.DryRun, "n", false, "Do not call any external dependencies like iptables")

	rootCmd.Flags().StringP(constants.ProxyUID, "u", "",
		"Specify the UID of the user for which the redirection is not applied. Typically, this is the UID of the proxy container")

	rootCmd.Flags().StringP(constants.ProxyGID, "g", "",
		"Specify the GID of the user for which the redirection is not applied. (same default value as -u param)")

	rootCmd.Flags().Bool(constants.RedirectDNS, dnsCaptureByAgent, "Enable capture of dns traffic by istio-agent")

	rootCmd.Flags().Bool(constants.RedirectAllDNSTraffic, false, "Enable capture of all dns traffic by istio-agent ")

	rootCmd.Flags().String(constants.AgentDNSListenerPort, constants.IstioAgentDNSListenerPort, "set listen port for DNS agent")

	rootCmd.Flags().String(constants.DNSUpstreamTargetChain, constants.RETURN, "(optional) the iptables chain where the upstream DNS requests should be directed to. It is only applied for IP V4. Use with care.")

	rootCmd.Flags().String(constants.SkipDNSConntrackZoneSplit, constants.RETURN, "Skip applying conntrack zone splitting iptables rules")
}

func GetCommand() *cobra.Command {
	return rootCmd
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		handleError(err)
	}
}

func handleError(err error) {
	fmt.Printf("%v\n", err)
	os.Exit(1)
}
