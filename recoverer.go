package snakepit

import (
	"fmt"
	"net/http"

	"github.com/pressly/chi"
	"golang.org/x/net/context"
)

type Recoverer struct {
	r *Render
}

func NewRecoverer(r *Render) func(next chi.Handler) chi.Handler {
	recoverer := &Recoverer{r: r}
	return recoverer.middleware
}

func (rec *Recoverer) middleware(next chi.Handler) chi.Handler {
	return chi.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		defer func() {
			if msg := recover(); msg != nil {
				err := APIError{Description: "An internal error occured. Please retry later.", ErrorCode: "INTERNAL_ERROR"}
				rec.r.JSONError(ctx, w, http.StatusInternalServerError, err, fmt.Errorf("%v", msg))
			}
		}()

		next.ServeHTTPC(ctx, w, r)
	})
}
