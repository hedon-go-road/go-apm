package dao

import (
	"context"
	"database/sql"

	"github.com/hedon-go-road/go-apm/dogapm"
)

type skuDao struct{}

var SkuDao = new(skuDao)

func (s *skuDao) Get(ctx context.Context, id int64) map[string]any {
	raw, err := dogapm.Infra.DB.QueryContext(ctx,
		"select * from `t_sku` where id = ?", id)
	if err != nil {
		return nil
	}
	info := dogapm.DBUtils.QueryFirst(raw, raw.Err())
	return info
}

func (s *skuDao) Decr(ctx context.Context, id int64, num int32) (sql.Result, error) {
	return dogapm.Infra.DB.ExecContext(ctx,
		"update `t_sku` set num = num - ? where id = ? and (num - ? >= 0)", num, id, num)
}
