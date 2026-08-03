package cipher

func None() Cipher {
	return &none{}
}

var _ Cipher = &none{}

type none struct{}

func (none) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (none) Decrypt(data []byte) ([]byte, error) {
	return data, nil
}
