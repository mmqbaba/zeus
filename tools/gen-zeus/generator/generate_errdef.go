package generator

import (
	"fmt"
	"log"
	"strings"
)

func GenerateErrdef(PD *Generator, rootdir string) (err error) {
	//err = genErrdef(PD, rootdir)
	//if err != nil {
	//	return
	//}

	err = genErrdefEnum(PD, rootdir)
	if err != nil {
		return
	}
	return
}

func genErrdef(PD *Generator, rootdir string) (err error) {
	header := ``
	context := `package errdef

import (
	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
)

var (
	ErrOK = errors.New(errors.ECodeSuccessed, "OK", "")
)

`
	fn := GetTargetFileName(PD, "errdef", rootdir)
	return writeContext(fn, header, context, false)
}

func genErrdefEnum(PD *Generator, rootdir string) error {
	header := _defaultHeader
	tmpContext := `package errdef

import (
	"net/http"

	"gitlab.dg.com/BackEnd/jichuchanpin/tif/zeus/errors"
)

// 每个子项目特有的错误码定义，避免使用 0 ~ 19999，与公共库冲突
const (
%s
)

func init() {
	// ECodeMsg and ECodeStatus
%s
}

`
	errConstBlock := ""
	errInitBlock := ""
	errcodeMap := make(map[int]string)
	for _, errSet := range PD.ErrCodes {
		if errSet.ErrCodeEnums != nil {
			for _, e := range errSet.ErrCodeEnums {
				if e.Integer <= 0 {
					continue
				}
				errConstBlock += fmt.Sprintf("	%s errors.ErrorCode = %d\n", e.Name, e.Integer)
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
				errInitBlock += fmt.Sprintf(`	errors.ECodeMsg[%s] = "%s"`, e.Name, errMsg) + "\n"
				if httpCode != "" {
					errInitBlock += fmt.Sprintf(`	errors.ECodeStatus[%s] = %s`, e.Name, httpCode) + "\n"
				}
				errcodeMap[e.Integer] = e.Name
			}
		}
	}
	context := fmt.Sprintf(tmpContext, errConstBlock, errInitBlock)
	fn := GetTargetFileName(PD, "errdef.enum", rootdir)
	return writeContext(fn, header, context, true)
}
