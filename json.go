package snakepit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/ansel1/merry"
)

var (
	APIJsonRendering = APIError{
		Description: "The JSON rendering failed.",
		ErrorCode:   "JSON_RENDERING_ERROR",
	}
	APIBodyDecoding = APIError{
		Description: "Could not decode the JSON request.",
		ErrorCode:   "BODY_DECODING_ERROR",
	}
)

type JSON struct{}

func NewJSON() *JSON {
	return &JSON{}
}

func (j *JSON) RenderError(
	ctx context.Context,
	w http.ResponseWriter,
	status int,
	apiError APIError,
	e error,
) {
	if entry, err := GetLogger(ctx); err == nil {
		file, line := merry.Location(e)
		*entry = *entry.WithError(e).WithField("location", fmt.Sprintf("%s:%d", file, line))

		if status >= 500 && status < 600 {
			*entry = *entry.WithField("stacktrace", merry.Details(e))
		}
	}

	apiError.Status = status
	apiError.Params = make(map[string]interface{})

	for k, v := range merry.Values(e) {
		strKey, ok := k.(string)
		if ok {
			apiError.Params[strKey] = v
		}
	}

	j.renderJSON(ctx, w, status, apiError)
}

func (j *JSON) Render(
	ctx context.Context,
	w http.ResponseWriter,
	status int,
	object interface{},
) {
	if object == nil {
		w.WriteHeader(status)
	} else {
		j.renderJSON(ctx, w, status, object)
	}
}

func (j *JSON) UnmarshalBody(
	ctx context.Context,
	w http.ResponseWriter,
	body io.ReadCloser,
	obj interface{},
) bool {
	buffer, _ := ioutil.ReadAll(body)

	if err := json.Unmarshal(buffer, obj); err != nil {
		j.RenderError(ctx, w, http.StatusBadRequest, APIBodyDecoding, err)
		return false
	}

	return true
}

func (j *JSON) UnmarshalBodyBulk(
	ctx context.Context,
	w http.ResponseWriter,
	body io.ReadCloser,
	objSlice interface{},
) (bool, bool) {
	bulk := false
	buffer, _ := ioutil.ReadAll(body)

	if len(buffer) < 2 {
		j.RenderError(ctx, w, http.StatusBadRequest, APIBodyDecoding, errors.New("empty or invalid body"))
		return false, false
	}

	if buffer[0] != '[' && buffer[len(buffer)-1] != ']' {
		buffer = append(append([]byte{'['}, buffer...), ']')
	} else {
		bulk = true
	}

	if err := json.Unmarshal(buffer, objSlice); err != nil {
		j.RenderError(ctx, w, http.StatusBadRequest, APIBodyDecoding, err)
		return false, false
	}

	return true, bulk
}

// func (j *JSON) Unmarshal(
// 	l *logrus.Entry,
// 	name string,
// 	raw []byte,
// 	obj interface{},
// ) error {
// 	if err := json.Unmarshal(raw, obj); err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (j *JSON) Marshal(
// 	l *logrus.Entry,
// 	name string,
// 	obj interface{},
// ) ([]byte, error) {
// 	buf, err := json.Marshal(obj)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf, nil
// }

func (j *JSON) renderJSON(
	ctx context.Context,
	w http.ResponseWriter,
	status int,
	object interface{},
) {
	j.writeHeaders(w, status)
	// Encode
	buf, err := json.Marshal(&object)
	if err != nil {
		j.renderRenderingError(ctx, w, err)
		return
	}

	// Write the buffer
	_, err = w.Write(buf)
	if err != nil {
		j.renderRenderingError(ctx, w, err)
		return
	}
}

func (j *JSON) renderRenderingError(
	ctx context.Context,
	w http.ResponseWriter,
	e error,
) {
	if entry, err := GetLogger(ctx); err == nil {
		*entry = *entry.WithError(e).WithField("stacktrace", j.stacktrace())
	}

	status := 500

	apiErr := APIJsonRendering
	apiErr.Status = status

	j.writeHeaders(w, status)

	// Encode
	buf, _ := json.Marshal(apiErr)

	// Write the buffer
	_, _ = w.Write(buf)
}

func (j *JSON) writeHeaders(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
}

func (j *JSON) stacktrace() string {
	stacktrace := string(debug.Stack())

	split := strings.SplitAfterN(stacktrace, "\n", 8)
	if len(split) >= 8 {
		stacktrace = split[0] + split[7]
	}

	return stacktrace
}
