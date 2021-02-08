package golden

import (
	"os"
)

func UpdateGoldenFiles() bool {
	value, found := os.LookupEnv("UPDATE_GOLDEN_FILES")
	return found && value == "true"
}

const RerunMsg = "Rerun the test with UPDATE_GOLDEN_FILES=true flag. Example: make test UPDATE_GOLDEN_FILES=true"
