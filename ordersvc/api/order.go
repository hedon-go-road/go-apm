package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/hedon-go-road/go-apm/dogapm"
)

type order struct {
}

var Order = &order{}

func (o *order) Add(w http.ResponseWriter, request *http.Request) {
	// get request body
	values := request.URL.Query()
	var (
		uid, _   = strconv.Atoi(values.Get("uid"))
		skuID, _ = strconv.Atoi(values.Get("sku_id"))
		num, _   = strconv.Atoi(values.Get("num"))
	)
	// check user info

	// deduct stock

	// create order
	_, err := dogapm.Infra.DB.ExecContext(context.TODO(),
		"INSERT INTO `t_order` (`order_id`, `sku_id`, `num`, `price`, `uid`) VALUES (?, ?, ?, ?, ?)",
		uuid.NewString(), skuID, num, 1, uid,
	)
	if err != nil {
		dogapm.Logger.Error(context.TODO(), "createOrder", map[string]any{
			"uid":    uid,
			"sku_id": skuID,
			"num":    num,
		}, err)
		dogapm.HttpStatus.Error(w, err.Error(), nil)
		return
	}

	// return
	dogapm.HttpStatus.Ok(w)
}
