package generator

import "fmt"

func GenerateBuildProtoSh(PD *Generator, rootdir string) (err error) {
	header := ``
	tmpContext := `#!/bin/bash


service=%s # 服务名

# gen-zeus
gen-zeus --proto proto/${service}.proto --dest ../

# gen-gomicro gen-validator
cd proto
mkdir -p gomicro
protoc -I. \
   -I$GOPATH/src \
   -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
   -I${GOPATH}/src/github.com/google/protobuf/src \
   --go_out=./gomicro/ \
   --micro_out=./gomicro/ \
   --govalidators_out=./gomicro \
   ./$service.proto
protoc-go-inject-tag -input=./gomicro/$service.pb.go # inject tag

# gen-grpc gen-grpc-gateway gen-validator
mkdir -p gw
protoc -I. \
   -I$GOPATH/src \
   -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
   -I${GOPATH}/src/github.com/google/protobuf/src \
   --go_out=plugins=grpc:./gw \
   --grpc-gateway_out=logtostderr=true:./gw \
   --govalidators_out=./gw \
   ./$service.proto
protoc-go-inject-tag -input=./gw/$service.pb.go # inject tag

# gen-swagger
protoc -I. \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  -I${GOPATH}/src/github.com/google/protobuf/src \
  --swagger_out=logtostderr=true:. \
   ./$service.proto

cd -
`
	context := fmt.Sprintf(tmpContext, PD.PackageName)
	fn := GetTargetFileName(PD, "build.proto", rootdir)
	return writeContext(fn, header, context, false)
}
