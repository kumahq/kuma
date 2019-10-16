package files

import "os"

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FileEmpty(path string) (bool, error) {
	file, err := os.Stat(path)
	if err != nil {
		return true, err
	}
	return file.Size() == 0, nil
}
