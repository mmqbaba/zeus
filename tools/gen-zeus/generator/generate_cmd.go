package generator

import (
	"fmt"
	"strings"
)

func GenerateCmd(PD *Generator, rootdir string) (err error) {
	err = genCmdInit(PD, rootdir)
	if err != nil {
		return err
	}
	err = genCmdMain(PD, rootdir)
	return
}

func genCmdInit(PD *Generator, rootdir string) error {
	header := _defaultHeader
	tmpContext := `package main

import (
	_ "%s/http"
	_ "%s/rpc"
)

`
	context := fmt.Sprintf(tmpContext, _defaultPkgPrefix+PD.PackageName, _defaultPkgPrefix+PD.PackageName)
	fn := GetTargetFileName(PD, "cmd.init", rootdir)
	return writeContext(fn, header, context, false)
}

func genCmdMain(PD *Generator, rootdir string) error {
	header := ``
	tmpContext := `package main

import (
	"log"
	"os"
	"runtime"
	"strconv"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/plugin/container"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/service"

	"{PRJ}/global"
)

func main() {
	num := runtime.NumCPU()
	log.Printf("[NumCPU] %v\n", num)
	gmp := os.Getenv("GOMAXPROCS")
	if gmp != "" {
		r, e := strconv.Atoi(gmp)
		if e == nil && r < num && r > 0 {
			num = r
		}
	}
	log.Printf("[GOMAXPROCS] %v\n", num)
	curr := runtime.GOMAXPROCS(num)
	log.Printf("[CURRENT GOMAXPROCS] %v\n", curr)
	log.Println("service run ...")
	if err := service.Run(container.GetContainer(), nil, global.ServiceOpts...); err != nil {
		log.Printf("Service exited with error: %s\n", err)
		os.Exit(255)
	} else {
		log.Println("Service exited gracefully")
	}
}

`
	context := strings.Replace(tmpContext, "{PRJ}", _defaultPkgPrefix+PD.PackageName, 1)
	fn := GetTargetFileName(PD, "cmd.main", rootdir)
	return writeContext(fn, header, context, false)
}
