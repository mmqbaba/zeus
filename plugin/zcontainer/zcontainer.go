package zcontainer

import (
	"net/http"

	"github.com/mmqbaba/zeus/httpclient/zhttpclient"

	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/sirupsen/logrus"

	"github.com/mmqbaba/zeus/config"
	"github.com/mmqbaba/zeus/mongo/zmongo"
	"github.com/mmqbaba/zeus/redis/zredis"
	tracing "github.com/mmqbaba/zeus/trace"
	"github.com/mmqbaba/zeus/utils"
)

// Container 组件的容器访问接口
type Container interface {
	utils.Releaser
	Init(appcfg *config.AppConf)
	Reload(appcfg *config.AppConf)
	GetRedisCli() zredis.Redis
	SetGoMicroClient(cli client.Client)
	GetGoMicroClient() client.Client
	GetLogger() *logrus.Logger
	GetAccessLogger() *logrus.Logger
	GetTracer() *tracing.TracerWrap
	SetServiceID(id string)
	GetServiceID() string
	SetHTTPHandler(h http.Handler)
	GetHTTPHandler() http.Handler
	SetGoMicroService(s micro.Service)
	GetGoMicroService() micro.Service
	GetMongo() zmongo.Mongo
	GetHttpClient() zhttpclient.HttpClient
}
