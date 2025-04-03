package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	v1 "review-b/api/review/v1"
	"review-b/internal/biz"
)

type businessRepo struct {
	data *Data
	log  *log.Helper
}

func (b *businessRepo) Reply(ctx context.Context, param *biz.ReplyParam) (int64, error) {
	b.log.WithContext(ctx).Infof("[data] Reply: params:%v", param)
	// 之前都是操作数据库，现在调用rpc服务实现
	ret, err := b.data.rc.ReplyReview(ctx, &v1.ReplyReviewRequest{
		ReviewID:  param.ReviewID,
		StoreID:   param.StoreID,
		Content:   param.Content,
		PicInfo:   param.PicInfo,
		VideoInfo: param.VideoInfo,
	})
	b.log.WithContext(ctx).Debugf("[data] Reply: ret:%v, err:%v", ret, err)
	if err != nil {
		b.log.WithContext(ctx).Infof("[data] Reply: err:%v", err)
		return 0, err
	}
	return ret.ReplyID, nil
}

func NewBusinessRepo(data *Data, logger log.Logger) biz.BusinessRepo {
	return &businessRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (b *businessRepo) Save(ctx context.Context) error {
	return nil
}
