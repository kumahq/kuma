package get

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
)

// RegexArgs return error if arguments don't pass regex test
func RegexArgs(validationMap map[int][]string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		// check if there are more args than expected
		if err := cobra.MaximumNArgs(len(validationMap))(cmd, args); err != nil {
			return err
		}

		for i, arg := range args {
			re := regexp.MustCompile(validationMap[i][0])
			if !re.MatchString(arg) {
				return fmt.Errorf("expected '%s', got '%s'", validationMap[i][1], arg)
			}
		}
		return nil
	}
}
