package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gen"
	"gorm.io/gorm"
	"review-service/internal/conf"
	"strings"
)

// GORM GEN生成代码配置

func connectDB(cfg *conf.Data_Database) *gorm.DB {
	if cfg == nil {
		panic(errors.New("GET:connectDB fail"))
	}

	switch strings.ToLower(cfg.Driver) {
	case "mysql":
		db, err := gorm.Open(mysql.Open(cfg.Source))
		if err != nil {
			panic(fmt.Errorf("failed to connect database: %w", err))
		}
		return db
	case "sqlite":
		db, err := gorm.Open(sqlite.Open(cfg.Source))
		if err != nil {
			panic(fmt.Errorf("failed to connect database: %w", err))
		}
		return db
	}
	panic(errors.New("GET:connectDB fail unsupported db driver: " + cfg.Driver))
}

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "../../configs", "config path,eg: -conf config.yaml")
}

// GEN 框架的生成配置
func main() {
	// 从配置文件读取数据库相关信息
	flag.Parse()
	c := config.New(
		config.WithSource(
			file.NewSource(flagConf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	// 指定生成代码的具体相对目录(相对当前文件夹),默认为:./query
	// 默认生成需要使用WithContext之后才可以查询代码，但可以通过设计gen.WithoutContext禁用该模式
	g := gen.NewGenerator(gen.Config{
		// 默认会在 OutPath目录生成CRUD代码，并且同目录下生成model包
		// 所以OutPath最后package不能设置为model,在有数据库表同步的情况下会产生冲突
		// 如果一定要使用,可以通过ModelPkgPath单独指定model package包名称
		OutPath: "../../internal/data/query",
		//ModelPkgPath: "dao/model",
		//Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
		Mode:          gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable: true, // 当字段可为空时，生成指针
	})

	// gormdb, _ := gorm.Open(mysql.Open("root:@(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local"))
	g.UseDB(connectDB(bc.Data.Database)) // reuse your gorm db

	// Generate basic type-safe DAO API for struct `model.User` following conventions
	g.ApplyBasic(g.GenerateAllTable()...)

	// 自定义查询逻辑

	// Generate the code
	g.Execute()
}
