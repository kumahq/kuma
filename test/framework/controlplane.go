package framework

import (
	"encoding/base64"
	"encoding/json"
)

func ExtractSecretDataFromResponse(output string) (string, error) {
	var secret map[string]string
	if err := json.Unmarshal([]byte(output), &secret); err != nil {
		return "", err
	}
	data := secret["data"]
	token, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
