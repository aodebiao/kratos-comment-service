package data

import (
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"review-service/internal/conf"
	"review-service/internal/data/query"
	"strings"
)

// ProviderSet is data providers.
// var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewReviewRepo, NewDB)
var ProviderSet = wire.NewSet(NewData, NewReviewRepo, NewDB, NewESClient, NewRedisClient)

// Data .
type Data struct {
	// TODO wrapped database client
	query *query.Query
	log   *log.Helper
	es    *elasticsearch.TypedClient
	rdb   *redis.Client
}

// NewData .
func NewData(db *gorm.DB, esClient *elasticsearch.TypedClient, rclient *redis.Client, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	// 非常重要，为GEN生成的query代码设置数据库对象
	query.SetDefault(db)
	return &Data{
		query: query.Q,
		es:    esClient,
		rdb:   rclient,
		log:   log.NewHelper(logger),
	}, cleanup, nil
}

func NewRedisClient(cfg *conf.Data) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		ReadTimeout:  cfg.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: cfg.Redis.WriteTimeout.AsDuration(),
	})
}

func NewESClient(cfg *conf.Elasticsearch) (*elasticsearch.TypedClient, error) {
	escfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
	}
	return elasticsearch.NewTypedClient(escfg)
}

func NewDB(c *conf.Data) (*gorm.DB, error) {
	if c == nil {
		panic(errors.New("GET:connectDB fail"))
	}

	switch strings.ToLower(c.Database.Driver) {
	case "mysql":
		db, err := gorm.Open(mysql.Open(c.Database.Source))
		if err != nil {
			panic(fmt.Errorf("failed to connect database: %w", err))
		}
		return db, nil
	case "sqlite":
		db, err := gorm.Open(sqlite.Open(c.Database.Source))
		if err != nil {
			panic(fmt.Errorf("failed to connect database: %w", err))
		}
		return db, nil
	}
	panic(errors.New("NewDB:connectDB fail unsupported db driver: " + c.Database.Driver))
}
