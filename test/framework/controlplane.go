package framework

import (
	"encoding/base64"
	"encoding/json"
)

func ExtractSecretDataFromResponse(output string) (string, error) {
	// Only "data" is needed; the response also carries non-string fields like "labels".
	var secret struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal([]byte(output), &secret); err != nil {
		return "", err
	}
	token, err := base64.StdEncoding.DecodeString(secret.Data)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
