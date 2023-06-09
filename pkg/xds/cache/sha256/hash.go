package sha256

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
)

func Hash(s string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(s)) // sha256.Write implementation doesn't return err
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func HashAny(a any) (string, error) {
	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(a); err != nil {
		return "", err
	}
	return Hash(b.String()), nil
}
