package snakepit

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pressly/chi"
	"golang.org/x/net/context"
)

type swaggerConf struct {
	Swagger     json.RawMessage `json:"swagger,omitempty"`
	BasePath    string          `json:"basePath,omitempty"`
	Schemes     []string        `json:"schemes,omitempty"`
	Consumes    json.RawMessage `json:"consumes,omitempty"`
	Produces    json.RawMessage `json:"produces,omitempty"`
	Info        json.RawMessage `json:"info,omitempty"`
	Paths       json.RawMessage `json:"paths,omitempty"`
	Definitions json.RawMessage `json:"definitions,omitempty"`
	Responses   json.RawMessage `json:"responses,omitempty"`
}

type Swagger struct {
	conf []byte
}

func NewSwagger(basePath, scheme string) func(next chi.Handler) chi.Handler {
	path := "./swagger.json"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = os.Getenv("HOME") + "/swagger.json"
	}

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	conf := &swaggerConf{}

	if err := json.Unmarshal(buf, conf); err != nil {
		panic(err)
	}

	conf.BasePath = basePath
	conf.Schemes = []string{scheme}

	raw, err := json.Marshal(conf)
	if err != nil {
		panic(err)
	}

	swagger := &Swagger{conf: raw}
	return swagger.middleware
}

func (rec *Swagger) middleware(next chi.Handler) chi.Handler {
	return chi.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/swagger" {
			w.WriteHeader(200)

			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.Header().Add("Access-Control-Allow-Methods", "GET")
			w.Header().Set("Content-Type", "application/json")

			w.Write(rec.conf)

			return
		}

		next.ServeHTTPC(ctx, w, r)
	})
}
