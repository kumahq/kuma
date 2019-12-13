package types

type KeyPair struct {
	Key  string `json:"key"`
	Cert string `json:"cert"`
}

type SigningCert struct {
	Id   string `json:"id"`
	Cert string `json:"cert"`
}
