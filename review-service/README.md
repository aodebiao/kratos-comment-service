
# 生成pb文件
kratos proto add api/review/v1/review.proto

# 生成客户端代码
kratos proto client api/review/v1/review.proto

# 生成服务端代码，-t 参数，指定存放目录
kratos proto server api/review/v1/review.proto -t internal/service

# 修改配置文件中的mysql和redis链接信息后，需要重新生成代码
```
protoc --proto_path=./internal --proto_path=./third_party --go_out=paths=source_relative:./internal internal/conf/conf.proto
```


## 接口开发流程
### 1、定义API文件
按照要求编写proto文件

### 2、生成客户端和服务端代码
- api
  protoc -I=./api -I=./third_party --go_out=paths=source_relative:./api --go-http_out=paths=source_relative:./api --go-grpc_out=paths=source_relative:./api \
  --openapi_out=fq_schema_naming=true,default_response=false:. api/business/v1/business.proto

- server
  kratos proto server api/review/v1/review.proto -t internal/service 
- client
  kratos proto client api/review/v1/review.proto 

### 填充业务逻辑
internal目录下
server -> service -> biz -> data


### GORM GEN生成
- 在cmd目录下新建generate.go文件
- 在generate.go中完善生成逻辑
- go mod tidy
- go run generate.go
### 完善ProviderSet执行Wire实现依赖注入
- 切换到程序main函数所在目录中，执行wire命令(直接命令行输入wire即可)

### 启动,项目根目录下执行kratos run
- 如果添加数据时，部分可以为空的字段在不伟值时，报错，需要配置GROM GEN 生成时		FieldNullable: true, // 当字段可为空时，生成指针



### 新增配置
- 同步修改config.yaml和config.proto后，需要重新生成配置文件相关的pb
> protoc --proto_path=./internal \
--proto_path=./third_party \
--go_out=paths=source_relative:./internal \
internal/conf/conf.proto



### 参数校验
- 1.go install github.com/envoyproxy/protoc-gen-validate@latest
- 2.在pb文件中按要求编写字段校验规则，在我们的项目的`api/review/v1/review.proto`中编写规则
- 3.规则编写好后，执行 `protoc -I=. -I=./third_party --go_out=paths=source_relative:. --validate_out=paths=source_relative,lang=go:. api/review/v1/review.proto`
- 4.在server目录中应用对应的中间件
```
  // http
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			validate.Validator(),
		),
	}
	
	// grpc
		var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			validate.Validator(),
			//v2.ProtoValidate(),
		),
	}

```


### 利用proto文件定义错误枚举
```
syntax = "proto3";
package api.review.v1;

option go_package = "review-service/api/review/v1;v1";
option java_multiple_files = true;
option java_package = "api.review.v1";

import "errors/errors.proto";

enum ErrorReason {

  // 设置缺省的错误码
  option(errors.default_code) = 500;

  // 为某个枚举值单独设置错误码
  NEED_LOGIN = 0 [(errors.code) = 401];

  DB_FAILED = 1 [(errors.code) = 500];

  ORDER_REVIEWED = 100 [(errors.code) = 400];

}


```

- 1.生成错误枚举 `protoc -I=. -I=./third_party --go_out=paths=source_relative:. --go-errors_out=paths=source_relative:. api/review/v1/*.proto `

### GEN事务
```
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
```


### 项目中如何管理pb文件
- proto文件要用一个(保证同一协议)
- protoc要使用同一个版本

> 通常在公司中都是把proto文件和生成的不同语言的代码都放在一个单独的公用代码库中，别的项目直接引用这个公用代码库

### git submodule
- 1. kratos new test-submodule
- 2. 删除直接删除该项目目录的api文件
- 3. 运行 `git submodule add git@github.com:Q1mi/reviewapis.git ./api  # 也就是说，当前项目下的api目录是引用了另一个git仓库
- 3.1 如果步骤报错，goland  git submodule add git@github.com:xxx/xxx.git ./api
  报错提示fatal: 'x-x/api' already exists in the index，类似这种，解决git rm -r --cached api

- 4. 常用命令
  # 用来初始化本地配置文件
  git submodule init
  # 从该项目中抓取所有数据并检出父项目中列表的合适的提交
  git submodule update
`

### data层注入rpc客户端，调用其它rpc服务
```
// data层
type Data struct {
	// TODO wrapped database client
	// 嵌入一个grpc Client端，去调用review-service服务
	rc  v1.ReviewClient
	log *log.Helper
}

func NewReviewServiceClient() v1.ReviewClient {
	// 	"github.com/go-kratos/kratos/v2/transport/grpc"
	conn, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("127.0.0.1:9000"),
		grpc.WithMiddleware(recovery.Recovery(), validate.Validator()),
	)
	if err != nil {
		panic(err)
	}
	return v1.NewReviewClient(conn)
}

// NewData .
func NewData(c *conf.Data, rc v1.ReviewClient, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		rc:  rc,
		log: log.NewHelper(logger),
	}, cleanup, nil
}



--------------------------
// 使用rc客户端
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

```

### 服务注册
#### 新增配置
```


// config.proto文件 注册中心相关配置
message Registry {
  message Consul {
    string address = 1;
    string scheme =2;
  }
  Consul consul = 1;
}

// register.yaml 配置文件
consul:
  address: 127.0.0.1:8500
  scheme: http
```

#### 修改完config.proto文件后，重新生成文件
` protoc --proto_path=./internal \
--proto_path=./third_party \
--go_out=paths=source_relative:./internal \
internal/conf/conf.proto
`

- main.go
```
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	// 解析register.yaml配置文件
	var rc conf.Registry
	if err := c.Scan(&rc); err != nil {
		panic(err)
	}
    // 如果像这样单独解析regsiter.yaml(未合并到config.yaml)，需要修改wire.go文件
	app, cleanup, err := wireApp(bc.Server, &rc, bc.Data, logger)
	
	
	//wire.go
	
	// 单独解析register，需要在wire.go函数中手动新增一个形参，*conf.Registry
    func wireApp(*conf.Server, *conf.Registry, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	  panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
    }

```

### 服务发现(参考review-o,review-b中的data层)


### canal使用

#### 修改mysql配置
- 修改mysql的配置文件(my.conf)
```
[mysqld]
log-bin=mysql-bin
binlog-format=ROW
server_id=1 # 配置mysql replaction需要定义，不要和canal的slaveId重复
```

- 添加授权
```
create user canal identified by 'canal'; # 创建一个用户名和密码都为canal的账号

grant select,replication slave,replication client on *.* to 'canal'@'%'; # 赋予权限

flush privileges; # 刷新权限
```


### docker pull canal/canal-server:latest

### docker exec -it canal-server /bin/bash

### json tag中的`,string`选项，可以用来指定从什么类型序列化
```
    {
          "id": 1,
          "userID": "147982601",
          "score": "5",
          "status": "2",
          "publishTime": "2023-09-09T16:07:42.499144+08:00",
          "content": "这是一个好评！",
          "orderID": "1231231",
          "store_id": "7890"
        }
上面的数据全是字符串类型
type test struct {
  score int64 `json:"score,string"`      
  orderID int64 `json:"orderID,string"`      
  store_id int64 `json:"store_id,string"`      

}

// 加了,string选项后，才能正确的反序列化，且序列化时，也能将score变成字符串返回

```


#### kratos openApi swagger使用
- go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
- protoc -I=. -I=./third_party --openapi_out=fq_schema_naming=true,default_response=false:. api/hellworld/v1/greeter.proto
