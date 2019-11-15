package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/tools/gen-zeus/generator"
)

func main() {

	sourceRoot := flag.String("dest", ".", "生成工程存储路径，不需要工程目录名称，例如/home/xxx/zeus_app_project/src")
	protoFile := flag.String("proto", "", "server proto file.")

	flag.Parse()
	var err error
	if len(*protoFile) <= 0 || !generator.FileExists(*protoFile) {
		fmt.Printf("can not find protofile(%s)\n", *protoFile)
		flag.Usage()
		return
	}

	reader, err := os.Open(*protoFile)
	if err != nil {
		log.Fatalf("Can not open proto file %s,error is %v", *protoFile, err)
		return
	}
	defer reader.Close()

	g, err := generator.New(reader)
	if err != nil {
		log.Fatal(err)
		return
	}
	var errcount int = 0

	err = generator.GenerateCmd(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate cmd file failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateGlobal(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate global file failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateHttp(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate http file failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateHandler(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate handler file failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateRpc(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate rpc file failed, error is %v\n", err)
		errcount++
	}

	//err = generator.GenerateProto(g, *sourceRoot, *protoFile)
	//if err != nil {
	//	fmt.Printf("Generate proto buffer file failed, error is %v\n", err)
	//	errcount++
	//}

	err = generator.GenerateProtoCopy(g, *sourceRoot, *protoFile)
	if err != nil {
		fmt.Printf("Generate build-proto.sh file failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateBuildProtoSh(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate build-proto.sh file failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateGoMod(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate go.mod file failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateMakefile(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate Makefile file failed, error is %v\n", err)
		errcount++
	}

	//err = generator.GenerateLogic(g, *sourceRoot)
	//if err != nil {
	//	fmt.Printf("Generate logic dir failed, error is %v\n", err)
	//	errcount++
	//}

	err = generator.GenerateResource(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate resource files failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateErrdef(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate errdef file failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateDockerfile(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate errdef file failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateReadme(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate errdef file failed, error is %v\n", err)
		errcount++
	}

	if errcount == 0 {
		fmt.Println("Generate rpc engin success!")
	} else {
		fmt.Println("Generate rpc engin have some error, please check error information!")
	}
	return
}
