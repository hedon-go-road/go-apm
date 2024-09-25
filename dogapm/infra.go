package dogapm

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

type infra struct {
	DB  *sql.DB
	RDB *redis.Client
}

var Infra = &infra{}

type InfraOption func(i *infra)

func WithMySQL(url string) InfraOption {
	return func(i *infra) {
		db, err := sql.Open("mysql", url)
		if err != nil {
			panic(err)
		}
		err = db.Ping()
		if err != nil {
			panic(err)
		}
		i.DB = db
	}
}

func WithRedis(url string) InfraOption {
	return func(i *infra) {
		client := redis.NewClient(&redis.Options{
			Addr:     url,
			DB:       0,
			Password: "",
		})
		res, err := client.Ping(context.TODO()).Result()
		if err != nil {
			panic(err)
		}
		if res != "PONG" {
			panic("redis ping failed")
		}
		i.RDB = client
	}
}

func (i *infra) Init(opts ...InfraOption) {
	for _, opt := range opts {
		opt(i)
	}
}
