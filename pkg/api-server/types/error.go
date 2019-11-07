package types

type Error struct {
	Title   string  `json:"title"`
	Details string  `json:"details"`
	Causes  []Cause `json:"causes,omitempty"`
}

type Cause struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
