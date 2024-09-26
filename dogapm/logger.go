package dogapm

import (
	"context"

	"github.com/sirupsen/logrus"
)

func init() {
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
