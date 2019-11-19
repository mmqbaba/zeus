package swagger

import "strings"

// SetService 设置默认swagger文件名
func SetService(name string) {
	s := strings.Replace(string(_third_partySwaggerUiIndexHtml), "{DEFAULT_SERVICE}", name, 1)
	_third_partySwaggerUiIndexHtml = []byte(s)
}