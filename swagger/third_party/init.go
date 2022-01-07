package third_party

import (
	"embed"

	swaggerhelper "github.com/mmqbaba/zeus/swagger"
)

//go:embed swagger-ui
var swaggerUI embed.FS

func init() {
	swaggerhelper.SetSwaggerUI(swaggerUI)
	// zservice.CommonServiceOptions = append(zservice.CommonServiceOptions, zservice.WithSetSwaggerServiceFn(setService))
}

// // setService 设置默认swagger文件名
// func setService(name string) {
// }
