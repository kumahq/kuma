package tokens

import (
	"os"

	config_types "github.com/kumahq/kuma/pkg/config/types"
)

type PublicKey struct {
	PEM string
	KID string
}

func PublicKeyFromConfig(publicKeys []config_types.PublicKey) ([]PublicKey, error) {
	var keys []PublicKey
	for _, key := range publicKeys {
		publicKey, err := configKeyToCoreKey(key)
		if err != nil {
			return nil, err
		}
		keys = append(keys, publicKey)
	}
	return keys, nil
}

func PublicKeyByMeshFromConfig(publicKeys []config_types.MeshedPublicKey) (map[string][]PublicKey, error) {
	byMesh := map[string][]PublicKey{}
	for _, key := range publicKeys {
		keys, ok := byMesh[key.Mesh]
		if !ok {
			keys = []PublicKey{}
		}
		publicKey, err := configKeyToCoreKey(key.PublicKey)
		if err != nil {
			return nil, err
		}
		keys = append(keys, publicKey)
		byMesh[key.Mesh] = keys
	}
	return byMesh, nil
}

func configKeyToCoreKey(key config_types.PublicKey) (PublicKey, error) {
	publicKey := PublicKey{
		KID: key.KID,
	}
	if key.KeyFile != "" {
		content, err := os.ReadFile(key.KeyFile)
		if err != nil {
			return PublicKey{}, err
		}
		publicKey.PEM = string(content)
	}
	if key.Key != "" {
		publicKey.PEM = key.Key
	}
	return publicKey, nil
}
