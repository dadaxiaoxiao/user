package ioc

import (
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

func InitEtcd() *etcdv3.Client {
	var cfg etcdv3.Config
	err := viper.UnmarshalKey("etcd", &cfg)
	if err != nil {
		panic(err)
	}
	cli, err := etcdv3.New(cfg)
	if err != nil {
		panic(err)
	}
	return cli
}
