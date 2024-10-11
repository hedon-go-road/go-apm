package dogapm

import "testing"

func TestSqlHook(t *testing.T) {
	Infra.Init(
		WithMySQL("root:root@tcp(127.0.0.1:23306)/ordersvc"),
		WithEnableAPM("127.0.0.1:4317"),
	)

	//nolint:all
	_, _ = Infra.DB.Query("select * sleep(1) from t_order limit ?;", 1)
	EndPoint.Shutdown()
}
