package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"review-service/internal/biz"
	"review-service/internal/data/model"
	"review-service/internal/data/query"
	"review-service/pkg/snowflake"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type reviewRepo struct {
	data *Data
	log  *log.Helper
}

// NewReviewRepo .
func NewReviewRepo(data *Data, logger log.Logger) biz.ReviewRepo {
	return &reviewRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *reviewRepo) SaveReview(ctx context.Context, review *model.ReviewInfo) (*model.ReviewInfo, error) {
	err := r.data.query.ReviewInfo.
		WithContext(ctx).
		Save(review)
	return review, err
}

// GetReviewByOrderID 根据订单ID查询评价
func (r *reviewRepo) GetReviewByOrderID(ctx context.Context, orderID int64) ([]*model.ReviewInfo, error) {
	return r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.OrderID.Eq(orderID)).
		Find()
}

func (r *reviewRepo) GetReview(ctx context.Context, reviewID int64) (*model.ReviewInfo, error) {
	return r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.ReviewID.Eq(reviewID)).
		First()
}

// SaveReply 保存评价回复
func (r *reviewRepo) SaveReply(ctx context.Context, reply *model.ReviewReplyInfo) (*model.ReviewReplyInfo, error) {
	// 1. 数据校验
	// 1.1 数据合法性校验（已回复的评价不允许商家再次回复）
	// 先用评价ID查库,看下是否已回复
	review, err := r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.ReviewID.Eq(reply.ReviewID)).
		First()
	if err != nil {
		return nil, err
	}
	if review.HasReply == 1 {
		return nil, errors.New("该评价已回复")
	}
	// 1.2 水平越权校验（A商家只能回复自己的不能回复B商家的）
	// 举例子：用户A删除订单，userID + orderID 当条件去查询订单然后删除
	if review.StoreID != reply.StoreID {
		return nil, errors.New("水平越权")
	}
	// 2. 更新数据库中的数据（评价回复表和评价表要同时更新，涉及到事务操作）
	// 事务操作
	err = r.data.query.Transaction(func(tx *query.Query) error {
		// 回复表插入一条数据
		if err := tx.ReviewReplyInfo.
			WithContext(ctx).
			Save(reply); err != nil {
			r.log.WithContext(ctx).Errorf("SaveReply create reply fail, err:%v", err)
			return err
		}
		// 评价表更新hasReply字段
		if _, err := tx.ReviewInfo.
			WithContext(ctx).
			Where(tx.ReviewInfo.ReviewID.Eq(reply.ReviewID)).
			Update(tx.ReviewInfo.HasReply, 1); err != nil {
			r.log.WithContext(ctx).Errorf("SaveReply update review fail, err:%v", err)
			return err
		}
		return nil
	})
	// 3. 返回
	return reply, err
}

// GetReviewReply 获取评价回复
func (r *reviewRepo) GetReviewReply(ctx context.Context, reviewID int64) (*model.ReviewReplyInfo, error) {
	return r.data.query.ReviewReplyInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewReplyInfo.ReviewID.Eq(reviewID)).
		First()
}

// AuditReview 审核评价（运营对用户的评价进行审核）
func (r *reviewRepo) AuditReview(ctx context.Context, param *biz.AuditParam) error {
	_, err := r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.ReviewID.Eq(param.ReviewID)).
		Updates(map[string]interface{}{
			"status":     param.Status,
			"op_user":    param.OpUser,
			"op_reason":  param.OpReason,
			"op_remarks": param.OpRemarks,
		})
	return err
}

// AppealReview 申诉评价（商家对用户评价进行申诉）
func (r *reviewRepo) AppealReview(ctx context.Context, param *biz.AppealParam) (*model.ReviewAppealInfo, error) {
	// 先查询有没有申诉
	ret, err := r.data.query.ReviewAppealInfo.
		WithContext(ctx).
		Where(
			query.ReviewAppealInfo.ReviewID.Eq(param.ReviewID),
			query.ReviewAppealInfo.StoreID.Eq(param.StoreID),
		).First()
	r.log.Debugf("AppealReview query, ret:%v err:%v", ret, err)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// 其他查询错误
		return nil, err
	}
	if err == nil && ret.Status > 10 {
		return nil, errors.New("该评价已有审核过的申诉记录")
	}
	// 查询不到审核过的申诉记录
	// 1. 有申诉记录但是处于待审核状态，需要更新
	// if ret != nil{
	// 	// update
	// }else{
	// 	// insert
	// }
	// 2. 没有申诉记录，需要创建
	appeal := &model.ReviewAppealInfo{
		ReviewID:  param.ReviewID,
		StoreID:   param.StoreID,
		Status:    10,
		Reason:    param.Reason,
		Content:   param.Content,
		PicInfo:   param.PicInfo,
		VideoInfo: param.VideoInfo,
	}
	if ret != nil {
		appeal.AppealID = ret.AppealID
	} else {
		appeal.AppealID = snowflake.GenID()
	}
	err = r.data.query.ReviewAppealInfo.
		WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "review_id"}, // ON DUPLICATE KEY，唯一索引
			},
			// 在review_id冲突时，更新，否则创建
			DoUpdates: clause.Assignments(map[string]interface{}{ // UPDATE
				"status":     appeal.Status,
				"content":    appeal.Content,
				"reason":     appeal.Reason,
				"pic_info":   appeal.PicInfo,
				"video_info": appeal.VideoInfo,
			}),
		}).
		Create(appeal) // INSERT
	r.log.Debugf("AppealReview, err:%v", err)
	return appeal, err
}

// AuditAppeal 审核申诉（运营对商家的申诉进行审核，审核通过会隐藏该评价）
func (r *reviewRepo) AuditAppeal(ctx context.Context, param *biz.AuditAppealParam) error {
	err := r.data.query.Transaction(func(tx *query.Query) error {
		// 申诉表
		if _, err := tx.ReviewAppealInfo.
			WithContext(ctx).
			Where(r.data.query.ReviewAppealInfo.AppealID.Eq(param.AppealID)).
			Updates(map[string]interface{}{
				"status":  param.Status,
				"op_user": param.OpUser,
			}); err != nil {
			return err
		}
		// 评价表
		if param.Status == 20 { // 申诉通过则需要隐藏评价
			if _, err := tx.ReviewInfo.WithContext(ctx).
				Where(tx.ReviewInfo.ReviewID.Eq(param.ReviewID)).
				Update(tx.ReviewInfo.Status, 40); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// ListReviewByUserID 根据userID查询所有评价
func (r *reviewRepo) ListReviewByUserID(ctx context.Context, userID int64, offset, limit int) ([]*model.ReviewInfo, error) {
	return r.data.query.ReviewInfo.
		WithContext(ctx).
		Where(r.data.query.ReviewInfo.UserID.Eq(userID)).
		Order(r.data.query.ReviewInfo.ID.Desc()).
		Limit(limit).
		Offset(offset).
		Find()
}

// ListReviewByStoreID 根据storeID分页查询
func (r *reviewRepo) ListReviewByStoreID(ctx context.Context, storeID int64, offset, limit int) ([]*biz.MyReviewInfo, error) {

	return r.getData2(ctx, storeID, offset, limit)
	//去es中查询
	//return r.getData1(ctx, storeID, offset, limit)

}

func (r *reviewRepo) getData1(ctx context.Context, storeID int64, offset int, limit int) ([]*biz.MyReviewInfo, error) {
	resp, err := r.data.es.Search().Index("review").
		From(offset).
		Size(limit).
		Query(&types.Query{
			Bool: &types.BoolQuery{
				Filter: []types.Query{
					{
						Term: map[string]types.TermQuery{
							"store_id": {
								Value: storeID,
							},
						},
					},
				},
			},
		}).Do(ctx)

	fmt.Printf("---> es search resp:%v ,%v\n", resp, err)
	if err != nil {
		return nil, err
	}
	list := make([]*biz.MyReviewInfo, 0, resp.Hits.Total.Value)
	for _, hit := range resp.Hits.Hits {
		temp := &biz.MyReviewInfo{}
		if err := json.Unmarshal(hit.Source_, temp); err != nil {
			r.log.Errorf("ReviewInfoJson Unmarshal err:%v", err)
			continue
		}
		list = append(list, temp)
	}

	fmt.Printf("es reuslt:Resp:%v\n", resp.Hits.Total.Value)
	fmt.Printf("es reuslt:%+v\n", resp.Hits.Hits)
	return list, nil
}

var g singleflight.Group

// 升级版，带缓存
func (r *reviewRepo) getData2(ctx context.Context, storeID int64, offset, limit int) ([]*biz.MyReviewInfo, error) {
	// 1.先查询缓存

	// 2.缓存没有则查es
	// 3.通过singleflight合并短时间内大量的并发请求
	key := fmt.Sprintf("review:%d:%d:%d", storeID, offset, limit)
	data, err := r.getDataBySingleflight(ctx, key)
	if err != nil {
		return nil, err
	}
	hm := new(types.HitsMetadata)
	if err := json.Unmarshal(data, hm); err != nil {
		return nil, err
	}
	// 反序列化
	list := make([]*biz.MyReviewInfo, 0, hm.Total.Value)
	for _, hit := range hm.Hits {
		temp := &biz.MyReviewInfo{}
		if err := json.Unmarshal(hit.Source_, temp); err != nil {
			r.log.Errorf("ReviewInfoJson Unmarshal err:%v", err)
			continue
		}
		list = append(list, temp)
	}
	return list, nil
}

// key review:231231:1:10
func (r *reviewRepo) getDataBySingleflight(ctx context.Context, key string) ([]byte, error) {
	v, err, share := g.Do(key, func() (interface{}, error) {
		data, err := r.getDataFromCache(ctx, key)
		r.log.Debugf("getDataFromCache,data:%s, err:%v", data, err)
		if err == nil {
			return data, nil
		}
		// 只有在缓存中没有这个key时，才查询
		if errors.Is(err, redis.Nil) {
			// 缓存中没有这个key,查es
			data, err := r.getDataFromES(ctx, key)
			r.log.Debugf("getDataFromES,data:%v, err:%v", data, err)
			if err == nil {
				// 设置缓存

				return data, r.setCache(ctx, key, data)
			}
			return nil, err
		}
		// 查缓存失败,直接返回，不继续向下传导压力
		return nil, err
	})
	r.log.Debugf("getDataBySingleflight,v:%v,err:%v share:%v", v, err, share)
	if err != nil {
		return nil, err
	}
	return v.([]byte), nil
}

func (r *reviewRepo) setCache(ctx context.Context, key string, data []byte) error {
	return r.data.rdb.Set(ctx, key, data, time.Second*20).Err()
}

func (r *reviewRepo) getDataFromCache(ctx context.Context, key string) ([]byte, error) {
	r.log.Debugf("getDataFromCache, key:%v", key)
	return r.data.rdb.Get(ctx, key).Bytes()
}

func (r *reviewRepo) getDataFromES(ctx context.Context, key string) ([]byte, error) {
	r.log.Debugf("getDataFromES, key:%v", key)
	values := strings.Split(key, ":")
	if len(values) < 4 {
		return nil, errors.New("invalid key")
	}
	index, storeID, offsetStr, limitStr := values[0], values[1], values[2], values[3]
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return nil, err
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return nil, err
	}
	resp, err := r.data.es.Search().Index(index).
		From(offset).
		Size(limit).
		Query(&types.Query{
			Bool: &types.BoolQuery{
				Filter: []types.Query{
					{
						Term: map[string]types.TermQuery{
							"store_id": {
								Value: storeID,
							},
						},
					},
				},
			},
		}).Do(ctx)

	if err != nil {
		return nil, err
	}
	return json.Marshal(resp.Hits)
}
