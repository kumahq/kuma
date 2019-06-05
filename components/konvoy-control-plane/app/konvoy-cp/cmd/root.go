package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	controlPlaneLog = ctrl.Log.WithName("konvoy-cp")

	debug bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "konvoy-control-plane",
	Short: "Universal Control Plane for Envoy-based Service Mesh",
	Long:  `Universal Control Plane for Envoy-based Service Mesh.`,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		ctrl.SetLogger(zap.Logger(debug))
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", true, "enable debug-level logging")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
