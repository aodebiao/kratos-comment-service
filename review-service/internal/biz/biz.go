package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
// var ProviderSet = wire.NewSet(NewGreeterUsecase, NewReviewUsecase)
var ProviderSet = wire.NewSet(NewReviewUsecase)
