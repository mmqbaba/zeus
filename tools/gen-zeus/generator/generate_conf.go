package generator

import "fmt"

func GenerateConf(PD *Generator, rootdir string) (err error) {
	err = genConf(PD, rootdir)
	if err != nil {
		return err
	}
	return
}

func genConf(PD *Generator, rootdir string) error {
	header := ``
	tmpContext := `{
    "engine_type": "etcd",
    "config_path": "/zeus/%s",
    "config_format": "json",
	"endpoints": ["127.0.0.1:2379"],
	"username": "root",
	"password": "123456"
}

`
	context := fmt.Sprintf(tmpContext, projectBasePrefix+PD.PackageName)
	fn := GetTargetFileName(PD, "conf", rootdir)
	return writeContext(fn, header, context, false)
}
