package generator

import (
	"fmt"
	"io/ioutil"
)

func GenerateProtoCopy(PD *Generator, rootdir string, pbFileName string) (err error) {

	fn := GetTargetFileName(PD, "proto", rootdir)
	byts, err := ioutil.ReadFile(pbFileName)
	if err != nil {
		return
	}
	return writeContext(fmt.Sprintf("%s/%s.proto", fn, PD.PackageName), "", string(byts), false)
}
