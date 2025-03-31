
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
- 4. 常用命令
  # 用来初始化本地配置文件
  git submodule init
  # 从该项目中抓取所有数据并检出父项目中列表的合适的提交
  git submodule update
`