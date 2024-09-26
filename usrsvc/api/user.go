package api

import (
	"context"

	"github.com/hedon-go-road/go-apm/protos"
	"github.com/hedon-go-road/go-apm/usrsvc/dao"
	"github.com/spf13/cast"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type User struct {
	protos.UnimplementedUserServiceServer
}

func (u *User) GetUser(ctx context.Context, req *protos.User) (*protos.User, error) {
	info := dao.UserDao.Get(ctx, req.Id)
	if info == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}
	return &protos.User{
		Name: info["name"].(string),
		Id:   cast.ToInt64(info["id"]),
	}, nil
}
