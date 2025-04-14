package ssh

type Host struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	User    string `json:"user"`

	PrivateKeyData []byte `json:"privateKeyData,omitempty"`
	PrivateKeyFile string `json:"privateKeyFile,omitempty"`
}
