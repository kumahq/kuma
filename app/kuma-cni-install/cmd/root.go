package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/kumahq/kuma/app/kuma-cni-install/pkg/health"
	"github.com/kumahq/kuma/app/kuma-cni-install/pkg/installer"

	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "install-cni",
	Short: "Install and configure Kuma CNI plugin on a node",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ctx := cmd.Context()

		var cfg *installer.Config
		if cfg, err = constructConfig(); err != nil {
			return
		}
		log.Printf("install cni with configuration: \n%+v", cfg)

		isReady := health.Start()

		installer := installer.NewInstaller(cfg, isReady)

		if err = installer.Run(ctx); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				// Error was caused by interrupt/termination signal
				err = nil
			}
		}

		if cleanErr := installer.Cleanup(); cleanErr != nil {
			if err != nil {
				err = fmt.Errorf("%v; %v", err, cleanErr.Error())
			} else {
				err = cleanErr
			}
		}

		return
	},
}

// GetCommand returns the main cobra.Command object for this application
func GetCommand() *cobra.Command {
	return rootCmd
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	registerStringParameter(installer.CNINetDir, "/etc/cni/net.d", "Directory on the host where CNI networks are installed")
	registerStringParameter(installer.CNIConfName, "", "Name of the CNI configuration file")
	registerBooleanParameter(installer.ChainedCNIPlugin, true, "Whether to install CNI plugin as a chained or standalone")
	registerStringParameter(installer.CNINetworkConfig, "", "CNI config template as a string")
	registerStringParameter(installer.LogLevel, "warn", "Fallback value for log level in CNI config file, if not specified in helm template")

	// Not configurable in CNI helm charts
	registerStringParameter(installer.MountedCNINetDir, "/host/etc/cni/net.d", "Directory on the container where CNI networks are installed")
	registerStringParameter(installer.CNINetworkConfigFile, "", "CNI config template as a file")
	registerStringParameter(installer.KubeconfigFilename, "ZZZ-kuma-cni-kubeconfig", "Name of the kubeconfig file")
	registerIntegerParameter(installer.KubeconfigMode, installer.DefaultKubeconfigMode, "File mode of the kubeconfig file")
	registerStringParameter(installer.KubeCAFile, "", "CA file for kubeconfig. Defaults to the pod one")
	registerBooleanParameter(installer.SkipTLSVerify, false, "Whether to use insecure TLS in kubeconfig file")
}

func registerStringParameter(name, value, usage string) {
	rootCmd.Flags().String(name, value, usage)
	bindViper(name)
}

func registerIntegerParameter(name string, value int, usage string) {
	rootCmd.Flags().Int(name, value, usage)
	bindViper(name)
}

func registerBooleanParameter(name string, value bool, usage string) {
	rootCmd.Flags().Bool(name, value, usage)
	bindViper(name)
}

func bindViper(name string) {
	if err := viper.BindPFlag(name, rootCmd.Flags().Lookup(name)); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
}

func constructConfig() (*installer.Config, error) {
	cfg := &installer.Config{
		CNINetDir:        viper.GetString(installer.CNINetDir),
		MountedCNINetDir: viper.GetString(installer.MountedCNINetDir),
		CNIConfName:      viper.GetString(installer.CNIConfName),
		ChainedCNIPlugin: viper.GetBool(installer.ChainedCNIPlugin),

		CNINetworkConfigFile: viper.GetString(installer.CNINetworkConfigFile),
		CNINetworkConfig:     viper.GetString(installer.CNINetworkConfig),

		LogLevel:           viper.GetString(installer.LogLevel),
		KubeconfigFilename: viper.GetString(installer.KubeconfigFilename),
		KubeconfigMode:     viper.GetInt(installer.KubeconfigMode),
		KubeCAFile:         viper.GetString(installer.KubeCAFile),
		SkipTLSVerify:      viper.GetBool(installer.SkipTLSVerify),
		K8sServiceProtocol: os.Getenv("KUBERNETES_SERVICE_PROTOCOL"),
		K8sServiceHost:     os.Getenv("KUBERNETES_SERVICE_HOST"),
		K8sServicePort:     os.Getenv("KUBERNETES_SERVICE_PORT"),
		K8sNodeName:        os.Getenv("KUBERNETES_NODE_NAME"),

		CNIBinSourceDir:      installer.CNIBinDir,
		CNIBinDestinationDir: installer.HostCNIBinDir,
	}

	if len(cfg.K8sNodeName) == 0 {
		var err error
		cfg.K8sNodeName, err = os.Hostname()
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
