package sha256

import (
	"crypto/sha256"
	"encoding/base64"
)

func Hash(s string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(s)) // sha256.Write implementation doesn't return err
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
