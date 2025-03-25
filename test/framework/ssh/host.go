package ssh

type Host struct {
	Address string
	Port    int
	User    string

	PrivateKeyData []byte
	PrivateKeyFile string
}
