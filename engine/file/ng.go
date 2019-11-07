package file

import (
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
)

type ng struct {
}

func (n *ng) Init(changes chan interface{}, cancelC chan struct{}, errorC chan struct{}) (err error) {
	return nil
}

func (n *ng) GetConfiger() (config.Configer, error) {
	return nil, nil
}

func (n *ng) GetContainer() *engine.Container {
	return nil
}
