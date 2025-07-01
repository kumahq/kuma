package core

type Port interface {
	GetName() string
	GetValue() int32
	GetNameOrStringifyPort() string
}
