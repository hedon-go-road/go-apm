package api

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/hedon-go-road/go-apm/dogapm"
	"github.com/hedon-go-road/go-apm/ordersvc/grpcclient"
	"github.com/hedon-go-road/go-apm/protos"
	"github.com/spf13/cast"
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
		num      = cast.ToInt32(values.Get("num"))
	)

	// check user info
	userInfo, err := grpcclient.UserClient.GetUser(context.TODO(), &protos.User{
		Id: int64(uid),
	})
	if err != nil {
		dogapm.HttpStatus.Error(w, err.Error(), nil)
		return
	}
	if userInfo.Id == 0 {
		dogapm.HttpStatus.Error(w, "user not found from user service", nil)
		return
	}

	// deduct stock
	res, err := grpcclient.SkuClient.DecreaseStock(context.TODO(), &protos.Sku{
		Id:  int64(skuID),
		Num: num,
	})
	if err != nil {
		dogapm.Logger.Error(context.TODO(), "createOrder", map[string]any{
			"sku_id": skuID,
			"num":    num,
		}, err)
		dogapm.HttpStatus.Error(w, err.Error(), nil)
		return
	}

	// create order
	_, err = dogapm.Infra.DB.ExecContext(context.TODO(),
		"INSERT INTO `t_order` (`order_id`, `sku_id`, `num`, `price`, `uid`) VALUES (?, ?, ?, ?, ?)",
		uuid.NewString(), skuID, num, res.Price, uid,
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
