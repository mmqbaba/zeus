package generator

import (
	"fmt"
	"log"
	"strings"
)

func GenerateZeusErrdef(PD *Generator, rootdir string) (err error) {
	err = genZeusErrdef(PD, rootdir)
	if err != nil {
		return err
	}
	return
}

func genZeusErrdef(PD *Generator, rootdir string) error {
	header := _defaultHeader
	tmpContext := `package errors

import (
	"net/http"
)

// 每个子项目特有的错误码定义，避免使用 0 ~ 19999，与公共库冲突
const (
%s
)

// ECodeMsg error message
var ECodeMsg = map[ErrorCode]string{
%s
}

// ECodeStatus http status code
var ECodeStatus = map[ErrorCode]int{
%s
}

`
	errConstBlock := ""
	errCodeMsgBlock := ""
	errCodeStatusBlock := ""
	errcodeMap := make(map[int]string)
	for _, errSet := range PD.ErrCodes {
		if errSet.ErrCodeEnums != nil {
			for _, e := range errSet.ErrCodeEnums {
				if e.Integer >= 20000 {
					fmt.Printf("！Invalid errcode %d (0 ~ 19999)\n", e.Integer)
					continue
				}
				errConstBlock += fmt.Sprintf("	%s ErrorCode = %d\n", e.Name, e.Integer)
				errMsg := ""
				httpCode := ""
				if e.InlineComment != nil {
					msgs := strings.Split(strings.TrimSpace(e.InlineComment.Message()), "^")
					if len(msgs) > 0 {
						errMsg = strings.TrimSpace(msgs[0])
					}
					if len(msgs) > 1 {
						httpCode = strings.TrimSpace(msgs[1])
					}
				}
				if nm, ok := errcodeMap[e.Integer]; ok {
					err := fmt.Errorf("errcode %d：%s 与 %s 重复\n", e.Integer, nm, e.Name)
					log.Fatalln(err)
					return err
				}
				errCodeMsgBlock += fmt.Sprintf(`	%s: "%s",`, e.Name, errMsg) + "\n"
				if httpCode != "" {
					errCodeStatusBlock += fmt.Sprintf(`	%s: %s,`, e.Name, httpCode) + "\n"
				}
				errcodeMap[e.Integer] = e.Name
			}
		}
	}
	context := fmt.Sprintf(tmpContext, errConstBlock, errCodeMsgBlock, errCodeStatusBlock)
	fn := GetTargetFileName(PD, "zeus.errdef", rootdir)
	return writeContext(fn, header, context, true)
}
