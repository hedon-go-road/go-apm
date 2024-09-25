package dogapm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfra_Init(t *testing.T) {
	assert.NotPanics(t, func() {
		Infra.Init(
			WithMySQL("root:root@tcp(127.0.0.1:23306)/ordersvc?charset=utf8mb4&parseTime=True&loc=Local"),
			WithRedis("127.0.0.1:26379"),
		)
	})
	assert.NotNil(t, Infra.DB)
	assert.NotNil(t, Infra.RDB)
}
