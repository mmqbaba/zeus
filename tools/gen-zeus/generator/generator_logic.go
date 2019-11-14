package generator

func GenerateLogic(PD *Generator, rootdir string) (err error) {
	GetTargetFileName(PD, "logic", rootdir)
	return
}
