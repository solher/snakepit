package snakepit

import (
	"net/http"
	"os"

	"github.com/pressly/chi"
	"golang.org/x/net/context"
)

type Swagger struct {
	path string
}

func NewSwagger() func(next chi.Handler) chi.Handler {
	path := "./swagger.json"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = os.Getenv("HOME") + "/swagger.json"
	}

	swagger := &Swagger{path: path}
	return swagger.middleware
}

func (rec *Swagger) middleware(next chi.Handler) chi.Handler {
	return chi.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/swagger" {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Methods", "GET")

			http.ServeFile(w, r, rec.path)
			return
		}

		next.ServeHTTPC(ctx, w, r)
	})
}
