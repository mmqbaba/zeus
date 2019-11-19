package generator

import "fmt"

func GenerateBuildProtoSh(PD *Generator, rootdir string) (err error) {
	header := ``
	tmpContext := `#!/bin/bash


projectpath=. # 具体的项目路径
service=%s # 服务名
pbout=${service}pb

test -f ../proto/${service}.proto || exit 1
# gen-zeus
gen-zeus --proto ../proto/${service}.proto --dest ../

mkdir -p $projectpath/proto/${service}pb
cd $projectpath/proto

# gen-gomicro gen-grpc-gateway gen-validator swagger

protoc -I../../proto \
   -I$GOPATH/src \
   -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
   -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
   --proto_path=${GOPATH}/src/github.com/google/protobuf/src \
   --go_out=plugins=grpc:./${service}pb \
   --grpc-gateway_out=logtostderr=true:./${service}pb \
   --micro_out=./${service}pb \
   --govalidators_out=./${service}pb \
   --swagger_out=logtostderr=true:. \
   $service.proto
protoc-go-inject-tag -input=./${service}pb/$service.pb.go # inject tag

sed -i 's/Register%sHandler(/Register%sHandlerGW(/g' ./${service}pb/$service.pb.gw.go
sed -i 's/ Register%sHandler / Register%sHandlerGW /g' ./${service}pb/$service.pb.gw.go

# gen-gomicro gen-grpc-gateway gen-validator swagger
# protoc -I. \
#    -I$GOPATH/src \
#    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
#    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway \
#    --proto_path=${GOPATH}/src/github.com/google/protobuf/src \
#    --go_out=plugins=grpc:./$service \
#    --grpc-gateway_out=logtostderr=true:./$service \
#    --micro_out=./$service \
#    --govalidators_out=./$service \
#    --swagger_out=logtostderr=true:. \
#    ./$service.proto
# protoc-go-inject-tag -input=./$service/$service.pb.go # inject tag

cd -
`
	context := fmt.Sprintf(tmpContext, PD.PackageName, PD.SvrName, PD.SvrName, PD.SvrName, PD.SvrName)
	fn := GetTargetFileName(PD, "build.proto", rootdir)
	return writeContext(fn, header, context, false)
}
