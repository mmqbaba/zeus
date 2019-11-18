package generator

import (
	"fmt"
	"strings"
)

func GenerateHandler(PD *Generator, rootdir string) (err error) {
	err = genHandleComm(PD, rootdir)
	if err != nil {
		return
	}

	return genHandleFun(PD, rootdir)
}

func genHandleComm(PD *Generator, rootdir string) (err error) {
	header := ``
	tmpContext := `package handler

type %s struct{}

func New%s()*%s{
	return &%s{}
}

`
	camelSrvName := CamelCase(PD.SvrName)
	context := fmt.Sprintf(tmpContext, camelSrvName, camelSrvName, camelSrvName, camelSrvName)
	fn := GetTargetFileName(PD, "handler.comm", rootdir)
	//fmt.Println(fn)
	return writeContext(fn, header, context, false)
}

func genHandleFun(PD *Generator, rootdir string) (err error) {
	header := ``
	tmpContext := `package handler

import (
	"context"

	gomicro "%s/proto/%spb"
)

`

	postFunc := `
func (h *%s) %s(ctx context.Context, req *gomicro.%s, rsp *gomicro.%s) (err error) {

	return
}
`
	streamFunc := `
func (h *%s) %s(ctx context.Context, stream gomicro.%s_%sStream) (err error) {
	
	return
}
`
	camelSrvName := CamelCase(PD.SvrName)
	for _, v := range PD.Rpcapi {
		context := fmt.Sprintf(tmpContext, _defaultPkgPrefix+PD.PackageName, PD.PackageName)
		if v.IsStreamsRequest {
			funtext := fmt.Sprintf(streamFunc, camelSrvName, v.Name, camelSrvName, v.Name)
			context += funtext
		} else if v.IsPost {
			funtext := fmt.Sprintf(postFunc, camelSrvName, v.Name, v.RequestType, v.ReturnsType)
			context += funtext
		} else {
			funtext := fmt.Sprintf(postFunc, camelSrvName, v.Name, v.RequestType, v.ReturnsType)
			context += funtext
		}

		fn := GetTargetFileName(PD, "handler", rootdir, strings.ToLower(v.Name))
		if FileExists(fn) {
			continue
		}
		if err := writeContext(fn, header, context, false); err != nil {
			fmt.Println(err)
			continue
		}

	}

	return
}
