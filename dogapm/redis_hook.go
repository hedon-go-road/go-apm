package dogapm

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	redisTracerName = "dogapm/redis"
)

type redisHook struct{}

func (h *redisHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *redisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	tracer := otel.Tracer(redisTracerName)
	return func(ctx context.Context, cmd redis.Cmder) error {
		ctx, span := tracer.Start(ctx, "redisProcessCmd")
		defer span.End()

		span.SetAttributes(attribute.String("cmd", truncate(cmd.String())))

		err := next(ctx, cmd)
		if err != nil && !errors.Is(err, redis.Nil) {
			span.SetAttributes(attribute.Bool("error", true))
			span.RecordError(err, trace.WithStackTrace(true))
		}
		return err
	}
}

func (h *redisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	tracer := otel.Tracer(redisTracerName)
	return func(ctx context.Context, cmds []redis.Cmder) error {
		ctx, span := tracer.Start(ctx, "redisProcessPipelineCmd")
		defer span.End()

		span.SetAttributes(attribute.String("cmd", truncate(fmt.Sprintf("%v", cmds))))

		err := next(ctx, cmds)
		if err != nil && !errors.Is(err, redis.Nil) {
			span.SetAttributes(attribute.Bool("error", true))
			span.RecordError(err, trace.WithStackTrace(true))
		}
		return err
	}
}
