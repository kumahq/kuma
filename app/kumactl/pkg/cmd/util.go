package cmd

import "github.com/spf13/cobra"

// RunParentPreRunE checks if the parent command has a PersistentPreRunE set and
// executes it. This is for use in PersistentPreRun and is necessary because
// only the first PersistentPreRun* of the command's ancestors is executed by
// cobra.
func RunParentPreRunE(cmd *cobra.Command, args []string) error {
	if p := cmd.Parent(); p != nil && p.PersistentPreRunE != nil {
		if err := p.PersistentPreRunE(p, args); err != nil {
			return err
		}
	}
	return nil
}
