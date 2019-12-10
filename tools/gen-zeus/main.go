package main

import (
	"flag"
	"fmt"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/tools/gen-zeus/generator"
	"log"
	"os"
	"path/filepath"
)

var (
	BuildDate = ""
	Version   = ""
	GoVersion = ""
)

func main() {
	log.Println("----------version info----------")
	log.Printf("Git Commit Hash: %s\n", Version)
	log.Printf("Build Date     : %s\n", BuildDate)
	log.Printf("Golang Version : %s\n", GoVersion)
	log.Println("--------------------------------")
	fmt.Print("\n")

	args := os.Args
	if len(args) == 2 && (args[1] == "--version" || args[1] == "-version" || args[1] == "-v") {
		return
	}

	sourceRoot := flag.String("dest", ".", "生成工程存储路径")
	protoFile := flag.String("proto", "", "server proto file.")
	projectBase := flag.String("base", "", "project base prefix.")
	errdefProto := flag.String("errdef", "proto/errdef.proto", "errdef.proto path")
	isZeuserr := flag.Bool("onlyzeuserr", false, "gen-zeus -onlyzeuserr -errdef=errors/errdef.proto")

	flag.Parse()
	var err error
	var reader *os.File
	if !*isZeuserr {
		if len(*protoFile) <= 0 || !generator.FileExists(*protoFile) {
			fmt.Printf("can not find protofile(%s)\n", *protoFile)
			flag.Usage()
			return
		}

		if len(*projectBase) > 0 && (*projectBase)[len(*projectBase)-1] != '/' {
			*projectBase += "/"
			generator.SetProjectBasePrefix(*projectBase)
		} else if fullPath, err := filepath.Abs(filepath.Dir(*sourceRoot + "/")); err != nil {
			log.Fatalf("Can not get full path %s, %s", *sourceRoot, err)
			return
		} else {
			baseName := filepath.Base(fullPath)
			if baseName != "" {
				generator.SetProjectBasePrefix(baseName + "/")
			}
		}
		reader, err = os.Open(*protoFile)
	} else {
		reader, err = os.Open(*errdefProto)
	}

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
	if *isZeuserr {
		generator.GenerateZeusErrdef(g, *sourceRoot)
		return
	}
	generator.WalkErrDefProto(*sourceRoot, g, g.Imports, *errdefProto)

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

	//err = generator.GenerateProtoCopy(g, *sourceRoot, *protoFile)
	//if err != nil {
	//	fmt.Printf("Generate build-proto.sh file failed, error is %v\n", err)
	//	errcount++
	//}

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

	err = generator.GenerateConf(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate conf/zeus.josn file failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateEditorconfig(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate .editorconfig file failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateGitignore(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate .gitignore file failed, error is %v\n", err)
		errcount++
	}

	err = generator.GenerateUnittest(g, *sourceRoot)
	if err != nil {
		fmt.Printf("Generate unittest files failed, error is %v\n", err)
		errcount++
	}

	if errcount == 0 {
		fmt.Printf("\n\nGenerate zeus engin success!\n")
	} else {
		fmt.Printf("\n\nGenerate zeus engin have some error, please check error information!\n")
		os.Exit(1)
	}
	return
}
