package snakepit

import (
	"net/http"
	"os"

	"github.com/pressly/chi"
	"golang.org/x/net/context"
)

type Swagger struct{}

func NewSwagger() func(next chi.Handler) chi.Handler {
	swagger := &Swagger{}
	return swagger.middleware
}

func (rec *Swagger) middleware(next chi.Handler) chi.Handler {
	return chi.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/swagger" {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Methods", "GET")

			file := "./swagger.json"
			if _, err := os.Stat(file); os.IsNotExist(err) {
				file = "~/swagger.json"
			}

			http.ServeFile(w, r, file)
			return
		}

		next.ServeHTTPC(ctx, w, r)
	})
}
