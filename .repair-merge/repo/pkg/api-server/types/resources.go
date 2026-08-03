package types

type CreateOrUpdateSuccessResponse struct {
	Warnings []string `json:"warnings,omitempty"`
}

type DeleteSuccessResponse struct{}
