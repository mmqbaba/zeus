package generator

import (
	"fmt"
	"strings"
)

func GenerateGoMod(PD *Generator, rootdir string) (err error) {
	header := ``
	tmpContext := `module %s

go 1.13

replace (
	github.com/golang/lint => golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
	github.com/testcontainers/testcontainer-go => github.com/testcontainers/testcontainers-go v0.0.4
	gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus => ../zeus
)

require (
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/gin-gonic/gin v1.3.0
	github.com/golang/protobuf v1.3.2
	github.com/grpc-ecosystem/grpc-gateway v1.12.0
	github.com/micro/go-micro v1.7.1-0.20190627135301-d8e998ad85fe
	github.com/mwitkow/go-proto-validators v0.2.0
	github.com/sirupsen/logrus v1.4.2
	gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus v0.0.0
	google.golang.org/genproto v0.0.0-20191028173616-919d9bdd9fe6
	google.golang.org/grpc v1.25.0
)

`
	fullPkg := strings.TrimRight(projectBasePrefix, "/")
	context := fmt.Sprintf(tmpContext, fullPkg)
	fn := GetTargetFileName(PD, "go.mod", rootdir)
	return writeContext(fn, header, context, false)
}
