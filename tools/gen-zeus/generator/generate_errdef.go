package generator

import (
	"fmt"
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
	for _, errSet := range PD.ErrCodes {
		if errSet.ErrCodeEnums != nil {
			for _, e := range errSet.ErrCodeEnums {
				if e.Integer < 20000 {
					continue
				}
				errConstBlock += fmt.Sprintf("	%s errors.ErrorCode = %d\n", e.Name, e.Integer)
				errMsg := ""
				if e.InlineComment != nil {
					errMsg = strings.TrimSpace(e.InlineComment.Message())
				}
				errInitBlock += fmt.Sprintf(`	errors.ECodeMsg[%s] = "%s"`, e.Name, errMsg) + "\n"
			}
		}
	}
	context := fmt.Sprintf(tmpContext, errConstBlock, errInitBlock)
	fn := GetTargetFileName(PD, "errdef.enum", rootdir)
	return writeContext(fn, header, context, true)
}
