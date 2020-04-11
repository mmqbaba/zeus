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
	github.com/coreos/etcd v3.3.17+incompatible
	github.com/gin-gonic/gin v1.5.0
	github.com/golang/protobuf v1.3.2
	github.com/grpc-ecosystem/grpc-gateway v1.12.0
	github.com/micro/go-micro v1.18.0
	github.com/mwitkow/go-proto-validators v0.2.0
	github.com/sirupsen/logrus v1.4.2
	gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus v0.0.0
	google.golang.org/genproto v0.0.0-20191108220845-16a3f7862a1a
	google.golang.org/grpc v1.25.1
)

`
	fullPkg := strings.TrimRight(projectBasePrefix, "/")
	context := fmt.Sprintf(tmpContext, fullPkg)
	fn := GetTargetFileName(PD, "go.mod", rootdir)
	return writeContext(fn, header, context, false)
}
