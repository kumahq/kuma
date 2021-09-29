package config

import (
	"github.com/spf13/cobra"
)

// DefaultConfigLoader is an object that can load a default configuration
type DefaultConfigLoader interface {
	Load() Config
}

// DefaultConfigLoaderFunc wraps a function to satisfy the
// DefaultConfigLoader interface.
type DefaultConfigLoaderFunc func() Config

// Load calls the loader function.
func (f DefaultConfigLoaderFunc) Load() Config {
	if f != nil {
		return f()
	}

	return nil
}

// NewConfigCmd returns a new "config" subcommand that shows the default
// configuration, and the configuration after being loaded from a specified
// file. This makes it easy for operators to obtain a valid initial configuration
// file for modification, and to observe the effects of environment variable
// overrides.
func NewConfigCmd(loader DefaultConfigLoader) *cobra.Command {
	defaultCmd := cobra.Command{
		Use:   "default",
		Short: "print the default configuration file",
		Long:  "Print the default configuration file.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := loader.Load()
			if cfg == nil {
				return nil
			}

			bytes, err := ToYAML(cfg)
			if err != nil {
				return nil
			}

			cmd.Println(string(bytes))
			return nil
		},
	}

	showCmd := cobra.Command{
		Use:   "view",
		Short: "print the current configuration file",
		Long:  "Print the current configuration file.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := loader.Load()
			if cfg == nil {
				return nil
			}

			flag := cmd.Flag("config-file")
			if err := Load(flag.Value.String(), cfg); err != nil {
				return err
			}

			bytes, err := ToYAML(cfg)
			if err != nil {
				return nil
			}

			cmd.Println(string(bytes))
			return nil
		},
	}

	showCmd.Flags().StringP("config-file", "c", "", "configuration file")

	cmd := &cobra.Command{
		Use:   "config",
		Short: "show configuration information",
		Long:  "Show configuration information.",
	}

	cmd.AddCommand(&showCmd)
	cmd.AddCommand(&defaultCmd)

	return cmd
}
