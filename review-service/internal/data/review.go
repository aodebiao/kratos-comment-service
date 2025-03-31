package data

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"review-service/internal/biz"
	"review-service/internal/data/model"
	"review-service/internal/data/query"
)

type reviewRepo struct {
	data *Data
	log  *log.Helper
}

func (r *reviewRepo) GetReview(ctx context.Context, reviewID int64) (*model.ReviewInfo, error) {
	return r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.ReviewID.Eq(reviewID)).
		First()
}

// SaveReply 保存评价回复
func (r *reviewRepo) SaveReply(ctx context.Context, info *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error) {
	// 1.数据校验
	// 1.1数据校验合法性（已 回复的评价不允许商家再次回复）
	// 先用评价id查库，看下是否已经回复
	review, err := r.data.query.ReviewInfo.WithContext(ctx).
		Where(r.data.query.ReviewInfo.ReviewID.Eq(info.ReplyID)).First()
	if err != nil {
		return nil, err
	}
	if review.HasReply == 1 {
		return nil, errors.New("该评价已经回复")
	}
	// 1.2水平越权检验（A商家只能回复自己的，不能回复B商家的)
	if review.StoreID != info.StoreID {
		return nil, errors.New("水平越权")
	}
	// 2.更新数据库中的数据(评价回复表和评价表需要同时更新，事务）
	// 事务操作
	err = r.data.query.Transaction(func(tx *query.Query) error {
		if err := tx.ReviewReplyInfo.WithContext(ctx).Save(info); err != nil {
			r.log.WithContext(ctx).Errorf("save review reply error: %v", err)
			return err
		}
		if _, err := tx.ReviewInfo.WithContext(ctx).
			Where(tx.ReviewInfo.ReviewID.Eq(info.ReviewID)).
			Update(tx.ReviewInfo.HasReply, 1); err != nil {
			r.log.WithContext(ctx).Errorf("update review  error: %v", err)
			return err
		}
		return nil
	})
	// 3.返回

	return info, err
}

func (r *reviewRepo) GetReviewReply(ctx context.Context, i int64) (*model.ReviewReplyInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (r *reviewRepo) AuditReview(ctx context.Context, param *biz.AuditParam) error {
	//TODO implement me
	panic("implement me")
}

func (r *reviewRepo) AppealReview(ctx context.Context, param *biz.AppealParam) error {
	//TODO implement me
	panic("implement me")
}

func (r *reviewRepo) AuditAppeal(ctx context.Context, param *biz.AuditAppealParam) error {
	//TODO implement me
	panic("implement me")
}

func (r *reviewRepo) ListReviewByUserID(ctx context.Context, userID int64, offset, limit int) ([]*model.ReviewInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (r *reviewRepo) GetReviewByOrderID(ctx context.Context, orderId int64) ([]*model.ReviewInfo, error) {
	return r.data.query.ReviewInfo.WithContext(ctx).Where(r.data.query.ReviewInfo.OrderID.Eq(orderId)).Find()
}

func (r *reviewRepo) SaveReview(ctx context.Context, info *model.ReviewInfo) (*model.ReviewInfo, error) {
	err := r.data.query.ReviewInfo.WithContext(ctx).Save(info)
	return info, err
}

// NewReviewRepo .
func NewReviewRepo(data *Data, logger log.Logger) biz.ReviewRepo {
	return &reviewRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
