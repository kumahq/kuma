package golden

import "os"

func UpdateGoldenFiles() bool {
	value, found := os.LookupEnv("UPDATE_GOLDEN_FILES")
	return found && value == "true"
}
