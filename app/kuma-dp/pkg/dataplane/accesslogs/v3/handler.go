package v3

type handler struct {
	sender logSender
}

func (h *handler) Handle(msg []byte) error {
	return h.sender.Send(msg)
}

func (h *handler) Close() error {
	return h.sender.Close()
}
