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

	return
}

func genGlobalInit(PD *Generator, rootdir string) error {
	header := _defaultHeader
	context := `package global

import (
	"log"

	"gitlab.dg.com/BackEnd/deliver/tif/zeus/config"
	"gitlab.dg.com/BackEnd/deliver/tif/zeus/engine"
	"gitlab.dg.com/BackEnd/deliver/tif/zeus/service"
)

var ng engine.Engine
var ServiceOpts = []service.Option{
	service.WithLoadEngineFnOption(func(ng engine.Engine) {
		log.Println("WithLoadEngineFnOption: SetNG success.")
		SetNG(ng)
        loadEngineSuccess(ng)
	}),
}

func init() {
	// load engine
	//loadEngineFnOpt := service.WithLoadEngineFnOption(func(ng engine.Engine) {
	//	log.Println("WithLoadEngineFnOption: SetNG success.")
	//	SetNG(ng)
	//	loadEngineSuccess(ng)
	//})
	processChangeFnOpt := service.WithProcessChangeFnOption(func(event interface{}) {
		processChange(event)
	})
	ServiceOpts = append(ServiceOpts, processChangeFnOpt)

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

type BaseExtConfig struct{

}
`
	fn := GetTargetFileName(PD, "global.init", rootdir)
	return writeContext(fn, header, context, true)
}

func genGlobal(PD *Generator, rootdir string) error {
	header := ``
	context := `package global
import (
    "gitlab.dg.com/BackEnd/deliver/tif/zeus/config"
    "gitlab.dg.com/BackEnd/deliver/tif/zeus/engine"
    "encoding/json"
	"log"
)

var CustomExtConfig = new(ExtConfig)

type ExtConfig struct {
    BaseExtConfig
}

func loadConfig(conf *config.AppConf) {
    // 加载配置
	tempExtBytes, err := json.Marshal(conf.Ext)
	if err != nil {
		log.Printf("[loadConfig] json.Marshal(conf.Ext) err: %s", err.Error())
	} else {
		err := json.Unmarshal(tempExtBytes, CustomExtConfig)
		if err != nil {
			log.Printf("[loadConfig] json.Unmarshal(tempExtBytes, CustomExtConfig) err: %s", err.Error())
		}
	}
	log.Printf("CustomExtConfig:%+v", CustomExtConfig)
}

func loadEngineSuccess(ng engine.Engine) {
    loadConfig(GetConfig())
    // 加载engine成功
    // TODO: do something here
}

func processChange(event interface{}) {
    loadConfig(GetConfig())
    // 配置变更
    // TODO: do something here
}

`
	fn := GetTargetFileName(PD, "global", rootdir)
	return writeContext(fn, header, context, false)
}
