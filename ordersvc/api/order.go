package api

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/hedon-go-road/go-apm/dogapm"
	"github.com/hedon-go-road/go-apm/ordersvc/grpcclient"
	"github.com/hedon-go-road/go-apm/ordersvc/metric"
	"github.com/hedon-go-road/go-apm/protos"
	"github.com/spf13/cast"
)

type order struct {
}

var Order = &order{}

func (o *order) Add(w http.ResponseWriter, request *http.Request) {
	ctx, span := dogapm.Tracer.Start(request.Context(), "order.Add-Start")
	defer span.End()

	// get request body
	values := request.URL.Query()
	var (
		uid, _   = strconv.Atoi(values.Get("uid"))
		skuID, _ = strconv.Atoi(values.Get("sku_id"))
		num      = cast.ToInt32(values.Get("num"))
	)

	// check user info
	userInfo, err := grpcclient.UserClient.GetUser(ctx, &protos.User{
		Id: int64(uid),
	})
	if err != nil {
		dogapm.Logger.Error(ctx, "get user info from user service", map[string]any{
			"uid":    uid,
			"sku_id": skuID,
			"num":    num,
		}, err)
		dogapm.HttpStatus.Error(w, err.Error(), nil)
		return
	}
	if userInfo.Id == 0 {
		dogapm.HttpStatus.Error(w, "user not found from user service", nil)
		return
	}

	// deduct stock
	res, err := grpcclient.SkuClient.DecreaseStock(ctx, &protos.Sku{
		Id:  int64(skuID),
		Num: num,
	})
	if err != nil {
		dogapm.Logger.Error(ctx, "createOrder", map[string]any{
			"sku_id": skuID,
			"num":    num,
		}, err)
		dogapm.HttpStatus.Error(w, err.Error(), nil)
		return
	}

	// create order
	_, err = dogapm.Infra.DB.ExecContext(ctx,
		"INSERT INTO `t_order` (`order_id`, `sku_id`, `num`, `price`, `uid`) VALUES (?, ?, ?, ?, ?)",
		uuid.NewString(), skuID, num, res.Price, uid,
	)
	if err != nil {
		dogapm.Logger.Error(ctx, "createOrder", map[string]any{
			"uid":    uid,
			"sku_id": skuID,
			"num":    num,
		}, err)
		dogapm.HttpStatus.Error(w, err.Error(), nil)
		return
	}

	// add order-success metric
	metric.OrderSuccessCounter.WithLabelValues(strconv.Itoa(skuID)).Inc()

	// return
	dogapm.HttpStatus.Ok(w)
}
