package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type ReplyParam struct {
	ReviewID  int64
	StoreID   int64
	Content   string
	PicInfo   string
	VideoInfo string
}

type BusinessRepo interface {
	Reply(context.Context, *ReplyParam) (int64, error)
}

type BusinessUseCase struct {
	repo BusinessRepo
	log  *log.Helper
}

func NewBusinessUseCase(repo BusinessRepo, logger log.Logger) *BusinessUseCase {
	return &BusinessUseCase{repo: repo, log: log.NewHelper(logger)}
}

// CreateReply 创建一个回复，service层调用
func (r *BusinessUseCase) CreateReply(ctx context.Context, param *ReplyParam) (int64, error) {
	r.log.WithContext(ctx).Infof("CreateReply: params:%v", param)
	return r.repo.Reply(ctx, param)
}
