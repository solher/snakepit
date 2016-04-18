package snakepit

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pressly/chi"
	"golang.org/x/net/context"
)

var dumbLogger *logrus.Entry

func init() {
	l := logrus.New()
	l.Out = nil
	dumbLogger = logrus.NewEntry(l)
}

const (
	contextLogger      CtxKey = "logger"
	contextResLogEntry CtxKey = "logEntry"
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

func GetResLogEntry(ctx context.Context) (*logrus.Entry, error) {
	if ctx == nil {
		return dumbLogger, errors.New("nil context")
	}

	entry, ok := ctx.Value(contextResLogEntry).(*logrus.Entry)
	if !ok {
		return dumbLogger, errors.New("unexpected type")
	}

	if entry == nil {
		return dumbLogger, errors.New("nil value in context")
	}

	return entry, nil
}

var (
	xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
	xRealIP       = http.CanonicalHeaderKey("X-Real-IP")
)

type Logger struct {
	log *logrus.Logger
}

func NewLogger(log *logrus.Logger) func(next chi.Handler) chi.Handler {
	logger := &Logger{log: log}
	return logger.middleware
}

func (l *Logger) middleware(next chi.Handler) chi.Handler {
	return chi.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if rip := l.realIP(r); rip != "" {
			r.RemoteAddr = rip
		}

		reqID, _ := GetRequestID(ctx)

		logger := l.log.WithFields(logrus.Fields{
			"reqId": reqID,
		})

		entry := logger.WithFields(logrus.Fields{
			"uri":    r.RequestURI,
			"method": r.Method,
			"remote": r.RemoteAddr,
		})
		entry.Info("Request arrived.")

		ctx = context.WithValue(ctx, contextLogger, logger)
		ctx = context.WithValue(ctx, contextResLogEntry, entry)

		proxy := wrapWriter(w)
		next.ServeHTTPC(ctx, proxy, r)
		proxy.maybeWriteHeader()

		status := proxy.status()

		entry = entry.WithFields(logrus.Fields{
			"status":  status,
			"latency": time.Since(start),
		})

		if status >= 500 && status < 600 {
			entry.Error("An unexpected error occured.")
			return
		}

		entry.Info("Request served.")
	})
}

func (l *Logger) realIP(r *http.Request) string {
	var ip string

	if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ", ")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	}

	return ip
}

func LogTime(log *logrus.Entry, name string, start time.Time) {
	if log != nil {
		log.WithField("latency", time.Now().Sub(start)).Debugf("%s time.", name)
	}
}
