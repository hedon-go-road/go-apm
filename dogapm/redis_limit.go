package dogapm

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisLimit struct{}

var RedisLimiter = &redisLimit{}

// IsLimit returns true if the key is limited.
func (r *redisLimit) IsLimit(client *redis.Client, key string, limit int, expire time.Duration) bool {
	sha, err := client.ScriptLoad(context.Background(), limitSscript).Result()
	if err != nil {
		return false
	}
	result, err := client.EvalSha(context.Background(), sha, []string{key}, limit, expire.Seconds()).Result()
	if err != nil {
		return false
	}
	return result == int64(1)
}

const limitSscript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local expire = tonumber(ARGV[2])
local current = tonumber(redis.call("GET", key) or "0")
if current + 1 > limit then
	return 1
else
	redis.call("INCR", key)
	redis.call("EXPIRE", key, expire)
	return 0
end
`
