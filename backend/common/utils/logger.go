package utils

import (
	"context"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	logger      *logrus.Logger
	loggerOnce  sync.Once
	serviceName string
)

// InitLogger initializes a global logrus logger for the process.
// It is safe to call multiple times; the underlying logger will
// only be created once, but the latest serviceName will be recorded.
func InitLogger(sName string) {
	serviceName = sName
	loggerOnce.Do(func() {
		l := logrus.New()
		l.SetOutput(os.Stdout)
		// l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
		l.SetFormatter(&logrus.JSONFormatter{})
		l.SetLevel(logrus.InfoLevel)
		logger = l
	})
}

// Logger returns the process-global logger, initializing it with a
// default service name if needed.
func Logger() *logrus.Logger {
	if logger == nil {
		InitLogger("unknown-service")
	}
	return logger
}

// ServiceName returns the configured logical service name.
func ServiceName() string {
	return serviceName
}

// WithContext returns a log entry enriched with service name and
// trace_id from the given context when available.
func WithContext(ctx context.Context) *logrus.Entry {
	entry := logrus.NewEntry(Logger())
	if serviceName != "" {
		entry = entry.WithField("service", serviceName)
	}
	if traceID := TraceIDFromContext(ctx); traceID != "" {
		entry = entry.WithField("trace_id", traceID)
	}
	return entry
}
