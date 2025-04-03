package service

import (
	"context"
	"review-b/internal/biz"

	pb "review-b/api/business/v1"
)

type BusinessService struct {
	pb.UnimplementedBusinessServer
	uc *biz.BusinessUseCase
}

func NewBusinessService(uc *biz.BusinessUseCase) *BusinessService {
	return &BusinessService{uc: uc}
}

func (s *BusinessService) ReplyReview(ctx context.Context, req *pb.ReplyReviewRequest) (*pb.ReplyReviewReply, error) {
	replyID, err := s.uc.CreateReply(ctx,
		&biz.ReplyParam{StoreID: req.StoreID,
			ReviewID:  req.ReviewID,
			Content:   req.Content,
			PicInfo:   req.PicInfo,
			VideoInfo: req.VideoInfo,
		})
	if err != nil {
		return nil, err
	}
	return &pb.ReplyReviewReply{ReplyID: replyID}, nil
}
