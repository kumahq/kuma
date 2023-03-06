package mux

type Filter interface {
	InterceptSession(session Session) error
}
