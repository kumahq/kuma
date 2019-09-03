package cipher

type Cipher interface {
	Encryptor
	Decryptor
}

type Encryptor interface {
	Encrypt([]byte) ([]byte, error)
}

type Decryptor interface {
	Decrypt([]byte) ([]byte, error)
}
