package types

type EchoResponse struct {
	Instance string               `json:"instance"`
	Received EchoResponseReceived `json:"received"`
}

type EchoResponseReceived struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
}
