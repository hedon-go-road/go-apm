package dogapm

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/xwb1989/sqlparser"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey string

const (
	ctxKeyBeginTime ctxKey = "begin"
	mysqlTracerName string = "dogapm/mysql"
	maxLength       int    = 1024

	slowSqlThreshold = 1 * time.Second
	longTxThreshold  = 3 * time.Second
)

func wrap(d driver.Driver, connectURL string) driver.Driver {
	tracer := otel.Tracer(mysqlTracerName)
	dsn, _ := mysql.ParseDSN(connectURL)
	return &Driver{d, Hooks{
		Before: func(ctx context.Context, query string, args ...any) (context.Context, error) {
			ctx = context.WithValue(ctx, ctxKeyBeginTime, time.Now())
			if ctx, span := tracer.Start(ctx, "sqltrace"); span != nil {
				span.SetAttributes(
					attribute.String("sql", truncate(query)),
					attribute.String("param", truncate(sliceToString(args))),
				)
				return ctx, nil
			}
			return ctx, nil
		},
		After: func(ctx context.Context, query string, args ...any) (context.Context, error) {
			// metric
			table, op, err, multiTable := SQLParser.parseTable(query)
			if !multiTable && err == nil {
				libraryCounter.WithLabelValues(LibraryTypeMySQL, sqlparser.StmtType(op), table, dsn.DBName+"."+dsn.Addr).Inc()
			}

			beginTime := time.Now()
			if begin := ctx.Value(ctxKeyBeginTime); begin != nil {
				beginTime = begin.(time.Time)
			}
			span := trace.SpanFromContext(ctx)
			elapsed := time.Since(beginTime)
			if elapsed >= slowSqlThreshold {
				span.SetAttributes(
					attribute.Bool("slowsql", true),
					attribute.Int64("sql_duration_ms", elapsed.Milliseconds()),
				)
			}
			span.End()
			return ctx, nil
		},
		OnError: func(ctx context.Context, err error, query string, args ...any) error {
			span := trace.SpanFromContext(ctx)
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

func sliceToString(args []any) string {
	if len(args) == 0 {
		return "[]"
	}
	return fmt.Sprintf("%v", args)
}
