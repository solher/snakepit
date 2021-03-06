package snakepit

// Ported from Goji's middleware, source:
// https://github.com/zenazn/goji/tree/master/web/middleware

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/pressly/chi"
	"golang.org/x/net/context"
)

const (
	contextRequestID CtxKey = "requestID"
)

// GetRequestID returns a request ID from the given context if one is present.
// Returns the empty string if a request ID cannot be found.
func GetRequestID(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", errors.New("nil context")
	}

	reqID, ok := ctx.Value(contextRequestID).(string)
	if !ok {
		return "", errors.New("unexpected type")
	}

	if len(reqID) == 0 {
		return "", errors.New("empty value in context")
	}

	return reqID, nil
}

var (
	prefix string
	reqid  uint64
)

// A quick note on the statistics here: we're trying to calculate the chance that
// two randomly generated base62 prefixes will collide. We use the formula from
// http://en.wikipedia.org/wiki/Birthday_problem
//
// P[m, n] \approx 1 - e^{-m^2/2n}
//
// We ballpark an upper bound for $m$ by imagining (for whatever reason) a server
// that restarts every second over 10 years, for $m = 86400 * 365 * 10 = 315360000$
//
// For a $k$ character base-62 identifier, we have $n(k) = 62^k$
//
// Plugging this in, we find $P[m, n(10)] \approx 5.75%$, which is good enough for
// our purposes, and is surely more than anyone would ever need in practice -- a
// process that is rebooted a handful of times a day for a hundred years has less
// than a millionth of a percent chance of generating two colliding IDs.

func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	prefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}

// RequestID is a middleware that injects a request ID into the context of each
// request. A request ID is a string of the form "host.example.com/random-0001",
// where "random" is a base62 random string that uniquely identifies this go
// process, and where the last number is an atomically incremented request
// counter.
type RequestID struct{}

func NewRequestID() func(next chi.Handler) chi.Handler {
	requestID := &RequestID{}
	return requestID.middleware
}

func (r *RequestID) middleware(next chi.Handler) chi.Handler {
	return chi.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		myid := atomic.AddUint64(&reqid, 1)
		ctx = context.WithValue(ctx, contextRequestID, fmt.Sprintf("%s-%06d", prefix, myid))
		next.ServeHTTPC(ctx, w, r)
	})
}
