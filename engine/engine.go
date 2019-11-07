package engine

import (
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/plugin"
)

// Engine for configuration and plugin container
type Engine interface {
	// Init 初始化
	// events 触发的事件，提供扩展
	// cancel 接收停止监听信号
	// errorC 通知外部，出现异常
	Init(events chan interface{}, cancel chan struct{}, errorC chan struct{}) error

	// GetConfiger 配置器
	GetConfiger() (config.Configer, error)

	// GetContainer 组件容器
	GetContainer() *plugin.Container
}

type NewEngineFn func() (Engine, error)

// func InitConfig(c Config, v interface{}) error {
// 	if err := c.init(v); err != nil {
// 		return err
// 	}
// 	return nil
// }

// //func WatchConfig(c Config, ch... chan interface{}) {
// //	c.watch(ch...)
// //}

// func WatchConfig(c Config, f ...func(v interface{})) {
// 	c.watch(f...)
// }

	// init(v interface{}) error
	// //watch(c... chan interface{})
	// watch(f ...func(v interface{}))