package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hedon-go-road/go-apm/dogapm"
)

type userDao struct{}

var UserDao = new(userDao)

func (u *userDao) Get(ctx context.Context, uid int64) map[string]any {
	userCache := dogapm.Infra.RDB.Get(ctx, userKey(uid)).Val()
	if userCache != "" {
		userInfo := make(map[string]any)
		err := json.Unmarshal([]byte(userCache), &userInfo)
		if err == nil {
			return userInfo
		}
	}

	raw, err := dogapm.Infra.DB.QueryContext(ctx,
		"select * from `t_user` where uid = ?", uid)
	if err != nil {
		return nil
	}
	info := dogapm.DBUtils.QueryFirst(raw, raw.Err())

	if info != nil {
		userInfo, err := json.Marshal(info)
		if err == nil {
			dogapm.Infra.RDB.Set(ctx, userKey(uid), string(userInfo), time.Minute*10)
		}
	}

	return info
}

func userKey(uid int64) string {
	return fmt.Sprintf("%s:%s:%d", "usersc", "uinfo", uid)
}
