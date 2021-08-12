package swagger

import (
	"strings"

	assetfs "github.com/elazarl/go-bindata-assetfs"

	zservice "gitlab.dg.com/BackEnd/deliver/tif/zeus/service"
	swaggerhelper "gitlab.dg.com/BackEnd/deliver/tif/zeus/swagger"
)

func init() {
	swaggerhelper.SetSwaggerAssetFS(&assetfs.AssetFS{
		Asset:    Asset,
		AssetDir: AssetDir,
		Prefix:   "third_party/swagger-ui",
	})
	zservice.CommonServiceOptions = append(zservice.CommonServiceOptions, zservice.WithSetSwaggerServiceFn(setService))
}

// setService 设置默认swagger文件名
func setService(name string) {
	s := strings.Replace(string(_third_partySwaggerUiIndexHtml), "{DEFAULT_SERVICE}", name, 1)
	_third_partySwaggerUiIndexHtml = []byte(s)
}
