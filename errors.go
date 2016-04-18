package snakepit

import (
	"fmt"

	"github.com/ansel1/merry"
)

// APIError defines a standard format for API errors.
type APIError struct {
	// The status code.
	Status int `json:"status"`
	// The description of the API error.
	Description string `json:"description"`
	// The token uniquely identifying the API error.
	ErrorCode string `json:"errorCode"`
	// Additional infos.
	Params map[string]interface{} `json:"params,omitempty"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s : %s", e.ErrorCode, e.Description)
}

func NewValidationError(field, err string) error {
	return merry.Errorf("%s cannot be %s", field, err).
		WithValue("field", field).
		WithValue("error", err)
}
