package swagger

import "strings"

// SetServer 设置默认swagger文件名
func SetServer(name string) {
	s := strings.Replace(string(_third_partySwaggerUiIndexHtml), "{DEFAULT_SERVER}", name, 1)
	_third_partySwaggerUiIndexHtml = []byte(s)
}