package container

import (
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/plugin"
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/plugin/zcontainer"
)

func GetContainer() zcontainer.Container {
	cnt := plugin.NewContainer()
	return cnt
}
