package zcontainer

import (
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/httpclient/zhttpclient"
	"net/http"

	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/sirupsen/logrus"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/mongo/zmongo"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/redis/zredis"
	tracing "gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/trace"
)

// Container 组件的容器访问接口
type Container interface {
	Init(appcfg *config.AppConf)
	Reload(appcfg *config.AppConf)
	GetRedisCli() zredis.Redis
	SetGoMicroClient(cli client.Client)
	GetGoMicroClient() client.Client
	GetLogger() *logrus.Logger
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
