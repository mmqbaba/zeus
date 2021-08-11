module gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus

go 1.16

replace (
	github.com/golang/lint => golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
	github.com/micro/go-micro => github.com/maidol/go-micro v1.18.1
	github.com/testcontainers/testcontainer-go => github.com/testcontainers/testcontainers-go v0.0.4
)

require (
	github.com/Shopify/sarama v1.24.1
	github.com/coreos/etcd v3.3.17+incompatible
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.0
	github.com/emicklei/proto v1.8.0
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/gin-gonic/gin v1.5.0
	github.com/go-playground/validator/v10 v10.2.0
	github.com/go-redis/redis v6.15.6+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-xorm/xorm v0.7.9
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/google/gops v0.3.7
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/grpc-ecosystem/grpc-gateway v1.12.0
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/json-iterator/go v1.1.8
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0
	github.com/lestrrat-go/file-rotatelogs v2.2.0+incompatible
	github.com/lestrrat-go/strftime v1.0.0 // indirect
	github.com/micro/go-micro v1.18.0
	github.com/micro/go-plugins v1.1.1
	github.com/onsi/ginkgo v1.10.3 // indirect
	github.com/onsi/gomega v1.7.1 // indirect
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.3.5 // indirect
	github.com/openzipkin/zipkin-go-opentracing v0.3.5
	github.com/prometheus/client_golang v1.2.1 // indirect
	github.com/prometheus/common v0.7.0
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/tebeka/strftime v0.1.3 // indirect
	go.mongodb.org/mongo-driver v1.5.0
	go.uber.org/ratelimit v0.1.0
	go.uber.org/zap v1.12.0
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	golang.org/x/tools v0.0.0-20200206050830-dd0d5d485177 // indirect
	google.golang.org/grpc v1.25.1
)
