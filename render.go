package snakepit

import (
	"net/http"
	"runtime/debug"
	"strings"

	"golang.org/x/net/context"

	"github.com/pquerna/ffjson/ffjson"
)

var APIJsonRendering = APIError{
	Description: "The JSON rendering failed.",
	ErrorCode:   "JSON_RENDERING_ERROR",
}

// Render is a ffjson based JSON renderer, customized to increase
// the expressiveness of API error rendering.
type Render struct{}

// NewRender returns a new instance of Render.
func NewRender() *Render {
	return &Render{}
}

// JSONError forges and writes an APIError into the response writer.
func (r *Render) JSONError(ctx context.Context, w http.ResponseWriter, status int, apiError APIError, e error) {
	entry, err := GetResLogEntry(ctx)
	if err != nil {
		panic(err)
	}

	*entry = *entry.WithError(e)

	if status >= 500 && status < 600 {
		*entry = *entry.WithField("stacktrace", r.stacktrace())
	}

	apiError.Status = status

	r.renderJSON(ctx, w, status, apiError)
}

// JSON writes the argument object into the response writer.
func (r *Render) JSON(ctx context.Context, w http.ResponseWriter, status int, object interface{}) {
	if object == nil {
		w.WriteHeader(status)
	} else {
		r.renderJSON(ctx, w, status, object)
	}
}

func (r *Render) renderJSON(ctx context.Context, w http.ResponseWriter, status int, object interface{}) {
	r.writeHeaders(w, status)

	// Encode
	buf, err := ffjson.Marshal(&object)
	if err != nil {
		r.renderRenderingError(ctx, w, err)
		return
	}

	// Write the buffer
	_, err = w.Write(buf)
	if err != nil {
		r.renderRenderingError(ctx, w, err)
		return
	}

	// We no longer need the buffer so we pool it.
	ffjson.Pool(buf)
}

func (r *Render) renderRenderingError(ctx context.Context, w http.ResponseWriter, e error) {
	entry, err := GetResLogEntry(ctx)
	if err != nil {
		panic(err)
	}

	*entry = *entry.WithError(e).WithField("stacktrace", r.stacktrace())

	status := 500

	apiErr := APIJsonRendering
	apiErr.Status = status

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

func (r *Render) stacktrace() string {
	stacktrace := string(debug.Stack())

	split := strings.SplitAfterN(stacktrace, "\n", 8)
	if len(split) >= 8 {
		stacktrace = split[0] + split[7]
	}

	return stacktrace
}
