package golden

import (
	"fmt"
	"os"
	"path/filepath"
)

func UpdateGoldenFiles() bool {
	value, found := os.LookupEnv("UPDATE_GOLDEN_FILES")
	return found && value == "true"
}

func RerunMsg(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path + " Failed to retrieve cwd"
	}
	return fmt.Sprintf("Rerun the test with UPDATE_GOLDEN_FILES=true flag to update file: %s. Example: make test UPDATE_GOLDEN_FILES=true", absPath)
}
