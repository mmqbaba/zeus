package swagger

import (
	"embed"

	assetfs "github.com/elazarl/go-bindata-assetfs"
)

// for swagger json file
var Swaggerfile embed.FS

// for swaggerui (use embed)
var SwaggerUI embed.FS

// for swaggerui (use go-bindata)
var SwaggerAssetFS *assetfs.AssetFS

func SetSwaggerfile(f embed.FS) {
	Swaggerfile = f
}

func SetSwaggerUI(f embed.FS) {
	SwaggerUI = f
}

func SetSwaggerAssetFS(f *assetfs.AssetFS) {
	SwaggerAssetFS = f
}
