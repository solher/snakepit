package snakepit

import "fmt"

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

type InternalError struct {
	Description string
	Params      map[string]interface{}
}

func (e InternalError) Error() string {
	return e.Description
}

func NewInternalError(d string) InternalError {
	return InternalError{Description: d}
}
