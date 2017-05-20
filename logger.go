package snakepit

import (
	"context"
	"errors"
	"net/http"

	"github.com/Sirupsen/logrus"
)

var dumbLogger *logrus.Entry

func init() {
	l := logrus.New()
	l.Out = nil
	dumbLogger = logrus.NewEntry(l)
}

const (
	contextLogger CtxKey = "logger"
)

func GetLogger(ctx context.Context) (*logrus.Entry, error) {
	if ctx == nil {
		return dumbLogger, errors.New("nil context")
	}

	logger, ok := ctx.Value(contextLogger).(*logrus.Entry)
	if !ok {
		return dumbLogger, errors.New("unexpected type")
	}

	if logger == nil {
		return dumbLogger, errors.New("nil value in context")
	}

	return logger, nil
}

type Logger struct {
	log *logrus.Logger
}

func NewLogger(log *logrus.Logger) func(next http.Handler) http.Handler {
	logger := &Logger{log: log}
	return logger.middleware
}

func (l *Logger) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry := l.log.WithFields(logrus.Fields{
			"path":   r.URL.RawPath,
			"method": r.Method,
		})
		r = r.WithContext(context.WithValue(r.Context(), contextLogger, entry))
		next.ServeHTTP(w, r)
	})
}
