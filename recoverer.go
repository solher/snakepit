package snakepit

import (
	"net/http"

	"github.com/ansel1/merry"
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
				err := merry.Errorf("%v", msg).WithStackSkipping(5)
				rec.JSON.RenderError(ctx, w, http.StatusInternalServerError, APIInternal, err)
			}
		}()

		next.ServeHTTPC(ctx, w, r)
	})
}
