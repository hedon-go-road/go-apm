package dao

import "github.com/hedon-go-road/go-apm/dogapm"

type deployInfo struct {
}

var DeployInfo = &deployInfo{}

func (d *deployInfo) GetInfoByApp(app string) map[string]any {
	return dogapm.DBUtils.QueryFirst(dogapm.Infra.DB.Query("select * from t_deploy_info where app=?", app))
}
