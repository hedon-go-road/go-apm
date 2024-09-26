package api

import (
	"context"

	"github.com/hedon-go-road/go-apm/protos"
	"github.com/hedon-go-road/go-apm/skusvc/dao"

	"github.com/spf13/cast"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SkuService struct {
	protos.UnimplementedSkuServiceServer
}

func (s *SkuService) DecreaseStock(ctx context.Context, req *protos.Sku) (*protos.Sku, error) {
	// 获取商品信息
	info := dao.SkuDao.Get(ctx, req.Id)
	if len(info) == 0 {
		return nil, status.Errorf(codes.NotFound, "sku not found")
	}

	// 进行扣减看库存
	res, err := dao.SkuDao.Decr(ctx, req.Id, req.Num)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "decrease stock failed: %v", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "decrease stock failed: %v", err)
	}

	if affected == 0 {
		return nil, status.Errorf(codes.Aborted, "stock not enough")
	}

	return &protos.Sku{
		Id:    cast.ToInt64(info["id"]),
		Name:  cast.ToString(info["name"]),
		Price: cast.ToInt32(info["price"]),
		Num:   cast.ToInt32(info["num"]) - req.Num,
	}, nil
}
