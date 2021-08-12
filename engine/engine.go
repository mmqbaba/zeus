package engine

import (
	"gitlab.dg.com/BackEnd/deliver/tif/zeus/config"
	"gitlab.dg.com/BackEnd/deliver/tif/zeus/plugin/zcontainer"
)

// Engine for configuration and plugin container
type Engine interface {
	// Init 初始化
	Init() error

	// Subscribe
	// events 触发的事件，提供扩展
	// cancel 接收停止监听信号
	Subscribe(events chan interface{}, cancel chan struct{}) error

	// GetConfiger 配置器
	GetConfiger() (config.Configer, error)

	// GetContainer 组件容器
	GetContainer() zcontainer.Container
}

type NewEngineFn func(cnt zcontainer.Container) (Engine, error)
