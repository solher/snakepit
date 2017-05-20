package snakepit

import (
	"net/http"

	"github.com/ansel1/merry"
)

var APIInternal = APIError{
	Description: "An internal error occured. Please retry later.",
	ErrorCode:   "INTERNAL_ERROR",
}

type Recoverer struct {
	JSON *JSON
}

func NewRecoverer(j *JSON) func(next http.Handler) http.Handler {
	recoverer := &Recoverer{JSON: j}
	return recoverer.middleware
}

func (rec *Recoverer) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if msg := recover(); msg != nil {
				err := merry.Errorf("%v", msg).WithStackSkipping(5)
				rec.JSON.RenderError(r.Context(), w, http.StatusInternalServerError, APIInternal, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
