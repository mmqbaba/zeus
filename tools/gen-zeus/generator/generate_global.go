package generator

func GenerateGlobal(PD *Generator, rootdir string) (err error) {
	err = genGlobal(PD, rootdir)
	if err != nil {
		return
	}
	err = genGlobalInit(PD, rootdir)
	if err != nil {
		return
	}
	err = genGlobalEnum(PD, rootdir)
	if err != nil {
		return
	}
	return
}

func genGlobalInit(PD *Generator, rootdir string) error {
	header := _defaultHeader
	context := `package global

import (
	"log"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/service"
)

var ng engine.Engine
var ServiceOpts []service.Option

func init() {
	// load engine
	loadEngineFnOpt := service.WithLoadEngineFnOption(func(ng engine.Engine) {
		log.Println("WithLoadEngineFnOption: SetNG success.")
		SetNG(ng)
	})
	ServiceOpts = append(ServiceOpts, loadEngineFnOpt)
	// // server wrap
	// ServiceOpts = append(ServiceOpts, service.WithGoMicroServerWrapGenerateFnOption(gomicro.GenerateServerLogWrap))
}

// GetNG ...
func GetNG() engine.Engine {
	return ng
}

// SetNG ...
func SetNG(n engine.Engine) {
	ng = n
}

// GetConfig ...
func GetConfig() (conf *config.AppConf) {
	c, err := ng.GetConfiger()
	if err != nil {
		log.Println("global.GetConfig err:", err)
		return
	}
	conf = c.Get()
	return
}

`
	fn := GetTargetFileName(PD, "global.init", rootdir)
	return writeContext(fn, header, context, false)
}

func genGlobal(PD *Generator, rootdir string) error {
	header := ``
	context := `package global

`
	fn := GetTargetFileName(PD, "global", rootdir)
	return writeContext(fn, header, context, false)
}

func genGlobalEnum(PD *Generator, rootdir string) error {
	header := ``
	context := `package global

import (
	"net/http"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/enum"
)

// 每个子项目特有的错误码定义，避免使用 0 ~ 19999，与公共库冲突
const (
	ECodeSampleServiceOK enum.ErrorCode = iota + 20000
	ECodeSampleServiceErr
)

func init() {
	// ECodeMsg and ECodeStatus
	enum.ECodeMsg[ECodeSampleServiceOK] = "ECodeSampleServiceOK"
	enum.ECodeStatus[ECodeSampleServiceOK] = http.StatusOK

	enum.ECodeMsg[ECodeSampleServiceErr] = "ECodeSampleServiceErr"
	enum.ECodeStatus[ECodeSampleServiceErr] = http.StatusInternalServerError
}

`
	fn := GetTargetFileName(PD, "global.enum", rootdir)
	return writeContext(fn, header, context, false)
}
