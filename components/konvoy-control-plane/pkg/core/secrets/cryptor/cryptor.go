package cryptor

type Cryptor interface {
	Encryptor
	Decryptor
}

type Encryptor interface {
	Encrypt([]byte) ([]byte, error)
}

type Decryptor interface {
	Decrypt([]byte) ([]byte, error)
}
