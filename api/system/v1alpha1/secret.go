package v1alpha1

// Secret defines an encrypted value in Kuma.
type Secret struct {
	// Value of the secret.
	Data *BytesValue `json:"data,omitempty"`
}

func (s *Secret) GetData() *BytesValue {
	if s == nil {
		return nil
	}
	return s.Data
}
