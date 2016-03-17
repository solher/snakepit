package snakepit

import (
	"net/http"

	"github.com/pquerna/ffjson/ffjson"
)

// Render is a ffjson based JSON renderer, customized to increase
// the expressiveness of API error rendering.
type Render struct{}

// NewRender returns a new instance of Render.
func NewRender() *Render {
	return &Render{}
}

// JSONError forges and writes an APIError into the response writer.
func (r *Render) JSONError(w http.ResponseWriter, status int, apiError APIError, err error) {
	if err != nil {
		apiError.Raw = err.Error()
	}

	apiError.Status = status

	r.renderJSON(w, status, apiError)
}

// JSON writes the argument object into the response writer.
func (r *Render) JSON(w http.ResponseWriter, status int, object interface{}) {
	if object == nil {
		w.WriteHeader(status)
	} else {
		r.renderJSON(w, status, object)
	}
}

func (r *Render) renderJSON(w http.ResponseWriter, status int, object interface{}) {
	r.writeHeaders(w, status)

	// Encode
	buf, err := ffjson.Marshal(&object)
	if err != nil {
		r.renderError(w, err)
		return
	}

	// Write the buffer
	_, err = w.Write(buf)
	if err != nil {
		r.renderError(w, err)
		return
	}

	// We no longer need the buffer so we pool it.
	ffjson.Pool(buf)
}

func (r *Render) renderError(w http.ResponseWriter, err error) {
	status := 500
	apiErr := &APIError{
		Status:      status,
		Description: "The JSON rendering failed.",
		Raw:         err.Error(),
		ErrorCode:   "JSON_RENDERING_ERROR",
	}

	r.writeHeaders(w, status)

	// Encode
	buf, _ := ffjson.Marshal(apiErr)

	// Write the buffer
	_, _ = w.Write(buf)

	// We no longer need the buffer so we pool it.
	ffjson.Pool(buf)
}

func (r *Render) writeHeaders(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}
