package types

type EchoResponse struct {
	Instance string              `json:"instance"`
	Received EchoReceivedRequest `json:"received"`
}

type EchoReceivedRequest struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
}
