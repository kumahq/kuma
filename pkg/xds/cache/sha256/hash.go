package sha256

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
)

func Hash(s string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(s)) // sha256.Write implementation doesn't return err
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func HashAny(a any) (string, error) {
	bytes, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	return Hash(string(bytes)), nil
}
