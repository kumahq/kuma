package cryptor

func None() Cryptor {
	return &none{}
}

var _ Cryptor = &none{}

type none struct{}

func (_ none) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (_ none) Decrypt(data []byte) ([]byte, error) {
	return data, nil
}
