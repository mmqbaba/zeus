package file

import (
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/config"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/engine"
)

type ng struct {
}

func (n *ng) Init() (err error) {
	return nil
}

func (n *ng) Subscribe(changes chan interface{}, cancelC chan struct{}) error {
	return nil
}

func (n *ng) GetConfiger() (config.Configer, error) {
	return nil, nil
}

func (n *ng) GetContainer() *engine.Container {
	return nil
}
