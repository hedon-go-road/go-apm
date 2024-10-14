package dogapm

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey string

const (
	ctxKeyBeginTime ctxKey = "begin"
	mysqlTracerName string = "dogapm/mysql"
	maxLength       int    = 1024
)

func wrap(d driver.Driver) driver.Driver {
	tracer := otel.Tracer(mysqlTracerName)
	return &Driver{d, Hooks{
		Before: func(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
			ctx = context.WithValue(ctx, ctxKeyBeginTime, time.Now())
			if ctx, span := tracer.Start(ctx, "sqltrace"); span != nil {
				span.SetAttributes(
					attribute.String("sql", truncate(query)),
					attribute.String("param", truncate(fmt.Sprintf("%v", args...))),
				)
				return ctx, nil
			}
			return ctx, nil
		},
		After: func(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
			beginTime := time.Now()
			if begin := ctx.Value(ctxKeyBeginTime); begin != nil {
				beginTime = begin.(time.Time)
			}
			now := time.Now()
			span := trace.SpanFromContext(ctx)
			elapsed := now.Sub(beginTime)
			if elapsed.Seconds() >= 1 {
				span.SetAttributes(
					attribute.Bool("slowsql", true),
					attribute.Int64("sql_duration_ms", elapsed.Milliseconds()),
				)
			}
			span.End()
			return ctx, nil
		},
		OnError: func(ctx context.Context, err error, query string, args ...interface{}) error {
			span := trace.SpanFromContext(ctx)
			fmt.Println("sql hook onerror: ", err)
			if !errors.Is(err, driver.ErrSkip) {
				span.SetAttributes(
					attribute.Bool("error", true),
				)
				span.RecordError(err, trace.WithStackTrace(true))
				span.End()
				return err
			}
			span.SetAttributes(attribute.Bool("drop", true))
			span.End()
			return err
		},
	}}
}

func truncate(query string) string {
	if len(query) > maxLength {
		return query[:maxLength]
	}
	return query
}
