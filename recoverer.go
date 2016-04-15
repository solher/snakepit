package snakepit

import (
	"fmt"
	"net/http"

	"github.com/pressly/chi"
	"golang.org/x/net/context"
)

var APIInternal = APIError{
	Description: "An internal error occured. Please retry later.",
	ErrorCode:   "INTERNAL_ERROR",
}

type Recoverer struct {
	JSON *JSON
}

func NewRecoverer(j *JSON) func(next chi.Handler) chi.Handler {
	recoverer := &Recoverer{JSON: j}
	return recoverer.middleware
}

func (rec *Recoverer) middleware(next chi.Handler) chi.Handler {
	return chi.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		defer func() {
			if msg := recover(); msg != nil {
				rec.JSON.RenderError(ctx, w, http.StatusInternalServerError, APIInternal, fmt.Errorf("%v", msg))
			}
		}()

		next.ServeHTTPC(ctx, w, r)
	})
}
