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
	mysqlTracerName = "dogapm/mysql"
	ctxKeyBeginTime = ctxKey("go.dogapm.mysql.begin_time")
	maxLength       = 1024
)

func wrap(dri driver.Driver) driver.Driver {
	tracer := otel.Tracer(mysqlTracerName)

	return &Driver{
		Driver: dri,
		hooks: Hooks{
			Before: func(ctx context.Context, query string, args ...any) (context.Context, error) {
				ctx = context.WithValue(ctx, ctxKeyBeginTime, time.Now())
				if ctx, span := tracer.Start(ctx, "sqltrace"); span != nil {
					span.SetAttributes(
						attribute.String("sql", truncateParams(query)),
						attribute.String("param", truncateParams(fmt.Sprintf("%v", args))),
					)
					return ctx, nil
				}
				return ctx, nil
			},
			After: func(ctx context.Context, query string, args ...any) error {
				beginTime := time.Now()
				if bt, ok := ctx.Value(ctxKeyBeginTime).(time.Time); ok {
					beginTime = bt
				}
				elapsed := time.Since(beginTime)
				span := trace.SpanFromContext(ctx)
				if elapsed.Seconds() >= 1 {
					span.SetAttributes(
						attribute.Bool("slowsql", true),
						attribute.Int64("sql_duration_ms", elapsed.Milliseconds()),
					)
				}
				span.End()
				return nil
			},
			OnError: func(ctx context.Context, err error, query string, args ...any) error {
				span := trace.SpanFromContext(ctx)

				// Ignore driver.ErrSkip because it just means the driver would run sql with prepared statement.
				if errors.Is(err, driver.ErrSkip) {
					// So we need to drop current span, because the following op is executing sql with prepared statement,
					// which would create a new span.
					span.SetAttributes(attribute.Bool("drop", true))
				} else {
					// If not ignore, set error flag and record error.
					span.SetAttributes(attribute.Bool("error", true))
					span.RecordError(err)
				}

				span.End()
				return err
			},
		},
	}
}

func truncateParams(params string) string {
	if len(params) < maxLength {
		return params
	}
	return params[:maxLength] + "..."
}
