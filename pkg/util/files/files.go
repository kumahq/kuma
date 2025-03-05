package files

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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

// IsDirWriteable checks if dir is writable by writing and removing a file
// to dir. It returns nil if dir is writable.
func IsDirWriteable(dir string) error {
	f := filepath.Join(dir, ".touch")
	perm := 0o600
	if err := os.WriteFile(f, []byte(""), fs.FileMode(perm)); err != nil {
		return err
	}
	return os.Remove(f)
}

// ToValidUnixFilename sanitizes input strings and concatenates them into a valid Unix filename.
func ToValidUnixFilename(input ...string) string {
	// Replace spaces with underscores
	noSpaces := strings.ReplaceAll(strings.Join(input, "_"), " ", "_")

	// Replace characters that are problematic in Unix filenames
	// This includes control characters, /, and other special characters
	// We also include a few additional characters that are problematic in GitHub Actions:
	// Double quote ", Colon :, Less than <, Greater than >, Vertical bar |, Asterisk *, Question mark ?, Carriage return \r, Line feed \n
	reg := regexp.MustCompile(`[^\w\-|?<>*:"\r\n]+`)
	sanitized := reg.ReplaceAllString(noSpaces, "-")

	// Trim leading/trailing periods and dashes which can cause issues
	sanitized = strings.Trim(sanitized, ".-")

	return sanitized
}
