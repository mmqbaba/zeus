package generator

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func GenerateProto(PD *Generator, rootdir string, pbFileName string) error {

	//pbFileName = path.Base(pbFileName)

	protoDir := path.Dir(pbFileName)
	protoFile := pbFileName[len(protoDir)+1:]
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Fatal("the GOPATH don't found")
		return nil
	}
	log.Println(gopath)

	var err error
	//curDir, err := os.Getwd()
	//if err != nil {
	//	log.Fatal(err)
	//	return err
	//}
	paths := strings.Split(gopath, ":")
	if len(paths) > 1 {
		gopath = paths[0]
	}
	// include
	var incsPath []string
	incsPath = append(incsPath, "--proto_path=./")

	if len(protoDir) > 1 && protoDir[0] != '/' {
		incsPath = append(incsPath, fmt.Sprintf("--proto_path=%s/%s", ".", protoDir))
	} else {
		incsPath = append(incsPath, fmt.Sprintf("--proto_path=%s", protoDir))
	}
	incsPath = append(incsPath, fmt.Sprintf("--proto_path=%s/src", gopath))
	incsPath = append(incsPath, fmt.Sprintf("--proto_path=%s/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis", gopath))
	incsPath = append(incsPath, fmt.Sprintf("--proto_path=%s/src/github.com/google/protobuf/src", gopath))

	// gw
	gwOutPath := GetTargetFileName(PD, "proto.gw", rootdir)
	microOutPath := GetTargetFileName(PD, "proto.gomicro", rootdir)
	outPath := GetTargetFileName(PD, "proto", rootdir)

	var gwsOut []string
	gwsOut = append(gwsOut, fmt.Sprintf("--go_out=plugins=grpc:%s", gwOutPath))
	gwsOut = append(gwsOut, fmt.Sprintf("--grpc-gateway_out=logtostderr=true:%s", gwOutPath))
	gwsOut = append(gwsOut, fmt.Sprintf("--govalidators_out=%s", gwOutPath))

	// micro
	var microOut []string
	microOut = append(microOut, fmt.Sprintf("--go_out=%s", microOutPath))
	microOut = append(microOut, fmt.Sprintf("--micro_out=%s", microOutPath))
	microOut = append(microOut, fmt.Sprintf("--govalidators_out=%s", microOutPath))
	microOut = append(microOut, fmt.Sprintf("--swagger_out=logtostderr=true:%s", outPath))

	log.Println("protoc", strings.Join(incsPath, " "), strings.Join(gwsOut, " "), protoFile)
	var gwArgs []string
	gwArgs = append(gwArgs, incsPath...)
	gwArgs = append(gwArgs, gwsOut...)
	gwArgs = append(gwArgs, protoFile)
	cmd := exec.Command("protoc", gwArgs...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		log.Println("create protobuf gw file error", cmd.Stderr)
		return err
	}
	log.Println("create protobuf gw file ok.")

	//////////////
	log.Println("protoc", strings.Join(incsPath, " "), strings.Join(microOut, " "), protoFile)
	var microArgs []string
	microArgs = append(microArgs, incsPath...)
	microArgs = append(microArgs, microOut...)
	microArgs = append(microArgs, protoFile)

	cmd = exec.Command("protoc", microArgs...)
	err = cmd.Run()
	if err != nil {
		log.Println("create protobuf micro file error", cmd.Stderr)
		return err
	}
	log.Println("create protobuf micro file ok.")
	//dirname := fmt.Sprintf("%s/%s", rootdir, PD.GopackageName)
	//err = TagInject(PD, dirname)
	return nil
}
