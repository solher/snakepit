package snakepit

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/ansel1/merry"
	"github.com/pquerna/ffjson/ffjson"
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
	if entry, err := GetResLogEntry(ctx); err == nil {
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
	logger, _ := GetLogger(ctx)
	buffer, _ := ioutil.ReadAll(body)

	if err := j.Unmarshal(logger, "Request body", buffer, obj); err != nil {
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
	logger, _ := GetLogger(ctx)
	buffer, _ := ioutil.ReadAll(body)

	if buffer[0] != '[' && buffer[len(buffer)-1] != ']' {
		buffer = append(append([]byte{'['}, buffer...), ']')
	} else {
		bulk = true
	}

	if err := j.Unmarshal(logger, "Request body", buffer, objSlice); err != nil {
		j.RenderError(ctx, w, http.StatusBadRequest, APIBodyDecoding, err)
		return false, false
	}

	return true, bulk
}

func (j *JSON) Unmarshal(
	l *logrus.Entry,
	name string,
	raw []byte,
	obj interface{},
) error {
	start := time.Now()

	if err := ffjson.Unmarshal(raw, obj); err != nil {
		return err
	}

	LogTime(l, name+" unmarshalling", start)

	return nil
}

func (j *JSON) Marshal(
	l *logrus.Entry,
	name string,
	obj interface{},
) ([]byte, error) {
	start := time.Now()

	buf, err := ffjson.Marshal(obj)
	if err != nil {
		return nil, err
	}

	LogTime(l, name+" marshalling", start)

	return buf, nil
}

func (j *JSON) renderJSON(
	ctx context.Context,
	w http.ResponseWriter,
	status int,
	object interface{},
) {
	j.writeHeaders(w, status)

	logger, _ := GetLogger(ctx)

	// Encode
	buf, err := j.Marshal(logger, "Response", &object)
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

	// We no longer need the buffer so we pool it.
	ffjson.Pool(buf)
}

func (j *JSON) renderRenderingError(
	ctx context.Context,
	w http.ResponseWriter,
	e error,
) {
	if entry, err := GetResLogEntry(ctx); err == nil {
		*entry = *entry.WithError(e).WithField("stacktrace", j.stacktrace())
	}

	status := 500

	apiErr := APIJsonRendering
	apiErr.Status = status

	j.writeHeaders(w, status)

	// Encode
	buf, _ := ffjson.Marshal(apiErr)

	// Write the buffer
	_, _ = w.Write(buf)

	// We no longer need the buffer so we pool it.
	ffjson.Pool(buf)
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
