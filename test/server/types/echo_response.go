package types

type EchoResponse struct {
	Instance string               `json:"instance"`
	Received EchoResponseReceived `json:"received"`
}

type EchoResponseReceived struct {
	// StatusCode is the HTTP response code. The echo server never
	// emits this, but it can be fabricated by curl output formatting.
	StatusCode int                 `json:"status,omitempty"`
	Method     string              `json:"method"`
	Path       string              `json:"path"`
	Headers    map[string][]string `json:"headers"`
}
