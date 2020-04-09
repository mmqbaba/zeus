package generator

import (
	"fmt"
	"strings"
)

func GenerateGitignore(PD *Generator, rootdir string) (err error) {
	header := ``
	tmpContext := `.vscode
*.a
*.exe
*.exe~
*.log
*.logger.Log
~$
%s_server

`

	context := fmt.Sprintf(tmpContext, strings.ToLower(PD.PackageName))
	fn := GetTargetFileName(PD, "gitignore", rootdir)
	return writeContext(fn, header, context, false)
}
