package snakepit

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/pressly/chi"
)

const contextTimer CtxKey = "timer"

type Timer struct {
	description string
}

func NewTimer(description string) *Timer {
	return &Timer{description: description}
}

func (t *Timer) Start(next chi.Handler) chi.Handler {
	return chi.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		ctx = context.WithValue(ctx, contextTimer, time.Now())
		next.ServeHTTPC(ctx, w, r)
	})
}

func (t *Timer) End(next chi.Handler) chi.Handler {
	return chi.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		logger, err := GetLogger(ctx)
		start, ok := ctx.Value(contextTimer).(time.Time)
		if err == nil && ok {
			LogTime(logger, t.description, start)
		}
		next.ServeHTTPC(ctx, w, r)
	})
}
