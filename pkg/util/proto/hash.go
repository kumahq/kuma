package proto

import (
	"crypto/sha256"
	"encoding/base64"

	"google.golang.org/protobuf/proto"
)

func Hash(message proto.Message) (string, error) {
	bytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(message)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	_, err = hash.Write(bytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}
