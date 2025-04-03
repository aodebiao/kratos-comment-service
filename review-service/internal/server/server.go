package server

import (
	consul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
	"review-service/internal/conf"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewRegister, NewGRPCServer, NewHTTPServer)

func NewRegister(conf *conf.Registry) registry.Registrar {
	c := api.DefaultConfig()
	// 使用配置文件中的配置
	c.Address = conf.Consul.Address
	c.Scheme = conf.Consul.Scheme
	client, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}
	reg := consul.New(client, consul.WithHealthCheck(true))
	kratos.Registrar(reg)
	return reg
}
