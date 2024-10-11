package dogapm

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const traceID = "trace_id"

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

type logger struct{}

var Logger = &logger{}

func (l *logger) Info(ctx context.Context, action string, kv map[string]any) {
	logrus.WithContext(ctx).
		WithFields(logrus.Fields(kv)).
		Info(action)
}

func (l *logger) Error(ctx context.Context, action string, kv map[string]any, err error) {
	if span := trace.SpanFromContext(ctx); span != nil {
		kv[traceID] = span.SpanContext().TraceID().String()
		span.SetAttributes(attribute.Bool("error", true))
		span.RecordError(err, trace.WithStackTrace(true), trace.WithTimestamp(time.Now()))
	}

	logrus.WithContext(ctx).
		WithFields(logrus.Fields(kv)).
		WithError(err).
		Error(action)
}

func (l *logger) Warn(ctx context.Context, action string, kv map[string]any) {
	logrus.WithContext(ctx).
		WithFields(logrus.Fields(kv)).
		Warn(action)
}

func (l *logger) Debug(ctx context.Context, action string, kv map[string]any) {
	logrus.WithContext(ctx).
		WithFields(logrus.Fields(kv)).
		Debug(action)
}

func (l *logger) Fatal(ctx context.Context, action string, kv map[string]any) {
	logrus.WithContext(ctx).
		WithFields(logrus.Fields(kv)).
		Fatal(action)
}
