package core

type Port interface {
	GetName() string
	GetValue() uint32
	GetNameOrStringifyPort() string
	GetProtocol() string
}
