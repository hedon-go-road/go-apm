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

	dogapm.Logger.Debug(ctx, "userDao.Get", map[string]any{
		"uid": uid,
	})

	raw, err := dogapm.Infra.DB.QueryContext(ctx,
		"select * from `t_user` where id = ?", uid)
	if err != nil {
		dogapm.Logger.Error(ctx, "userDao.Get", map[string]any{
			"uid": uid,
		}, err)
		return nil
	}
	info := dogapm.DBUtils.QueryFirst(raw, raw.Err())

	dogapm.Logger.Debug(ctx, "userDao.Get", map[string]any{
		"info": info,
	})

	if info != nil {
		userInfo, err := json.Marshal(info)
		if err == nil {
			dogapm.Infra.RDB.Set(ctx, userKey(uid), string(userInfo), time.Minute*10)
		}
	}

	fmt.Println("info", info)

	return info
}

func userKey(uid int64) string {
	return fmt.Sprintf("%s:%s:%d", "usersc", "uinfo", uid)
}
