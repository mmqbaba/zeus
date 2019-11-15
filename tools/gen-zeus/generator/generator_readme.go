package generator

import (
	"strings"
)

func GenerateReadme(PD *Generator, rootdir string) (err error) {
	header := ``
	tmpContext := `## 配置入口
{QUOTE}json
	// 默认路径/etc/tif/zeus.json
	{
		"engine_type": "etcd",
		"config_path": "/zeus/{PKG}", // 服务应用的配置路径
		"config_format": "json",     // 配置格式
		"endpoints": ["127.0.0.1:2379"],
		"username": "root",
		"password": "123456"
	}
	{QUOTE}

## 应用服务配置
{QUOTE}json
	// 路径/zeus/{PKG}
	{
		"redis": {
		"host": "127.0.0.1:6379",
			"pwd": ""
	},
		"go_micro": {
			"server_name": "zeus",
			"registry_plugin_type": "etcd",
			"registry_addrs": ["127.0.0.1:2379"],
			"registry_authuser": "root",
			"registry_authpwd": "123456"
	}
	}
	{QUOTE}

## gen-proto
{QUOTE}bash
	./build-proto.sh
	{QUOTE}

## run
{QUOTE}bash
	go run ./cmd/app
	{QUOTE}
`
	context := strings.ReplaceAll(tmpContext, "{QUOTE}", "```")
	context = strings.ReplaceAll(context, "{PKG}", strings.ToLower(PD.PackageName))
	fn := GetTargetFileName(PD, "readme", rootdir)
	return writeContext(fn, header, context, false)
}
