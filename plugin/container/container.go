package container

import (
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/plugin"
)

func GetContainer() *plugin.Container {
	cnt := plugin.NewContainer()
	return cnt
}
