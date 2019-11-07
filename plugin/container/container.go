package container

import (
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/plugin"
)

func GetContainer() *plugin.Container {
	cnt := plugin.NewContainer()
	// specs := []*plugin.MiddlewareSpec{
	// 	&plugin.MiddlewareSpec{Type: "oauth", MW: authtoken.New},
	// }
	// for _, spec := range specs {
	// 	if err := r.AddMW(spec); err != nil {
	// 		panic(err)
	// 	}
	// }
	return cnt
}