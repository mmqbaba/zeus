package generator

import (
	"fmt"
	"strings"
)

func GenerateCmd(PD *Generator, rootdir string) (err error) {
	if err := genCmdInit(PD, rootdir); err != nil {
		return err
	}
	if err := genCmdMain(PD, rootdir); err != nil {
		return err
	}
	if err := genCmdSwaggerUI(PD, rootdir); err != nil {
		return err
	}
	return
}

func genCmdInit(PD *Generator, rootdir string) error {
	header := _defaultHeader
	tmpContext := `package main

import (
	_ "%s/http"
	_ "%s/rpc"
	_ "%s/proto"
)

`
	context := fmt.Sprintf(tmpContext, projectBasePrefix+PD.PackageName, projectBasePrefix+PD.PackageName, projectBasePrefix+PD.PackageName)
	fn := GetTargetFileName(PD, "cmd.init", rootdir)
	return writeContext(fn, header, context, true)
}

func genCmdMain(PD *Generator, rootdir string) error {
	header := ``
	tmpContext := `package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/mmqbaba/zeus/plugin/container"
	"github.com/mmqbaba/zeus/service"

	"{PRJ}/global"
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

	num := runtime.NumCPU()
	log.Printf("[NumCPU] %v\n", num)
	gmp := os.Getenv("GOMAXPROCS")
	if gmp != "" {
		r, e := strconv.Atoi(gmp)
		// 限制线程数在cpu核数范围内
		if e == nil && r < num && r > 0 {
			num = r
		}
	}
	log.Printf("[GOMAXPROCS] %v\n", num)
	runtime.GOMAXPROCS(num)

	log.Println("service run ...")
	if err := service.Run(container.GetContainer(), nil, global.ServiceOpts...); err != nil {
		log.Printf("Service exited with error: %s\n", err)
		os.Exit(255)
	} else {
		log.Println("Service exited gracefully")
	}
}

`
	context := strings.Replace(tmpContext, "{PRJ}", projectBasePrefix+PD.PackageName, 1)
	fn := GetTargetFileName(PD, "cmd.main", rootdir)
	return writeContext(fn, header, context, false)
}

func genCmdSwaggerUI(PD *Generator, rootdir string) error {
	header := _defaultHeader
	tmpContext := `package main

import (
	_ "github.com/mmqbaba/zeus/swagger/third_party"
)

`
	context := fmt.Sprintf(tmpContext)
	fn := GetTargetFileName(PD, "cmd.swaggerui.plugin", rootdir)
	return writeContext(fn, header, context, false)
}
