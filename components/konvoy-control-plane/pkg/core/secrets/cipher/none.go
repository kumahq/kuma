package cipher

func None() Cipher {
	return &none{}
}

var _ Cipher = &none{}

type none struct{}

func (_ none) Encrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (_ none) Decrypt(data []byte) ([]byte, error) {
	return data, nil
}
