package generator

func GenerateEditorconfig(PD *Generator, rootdir string) (err error) {
	header := ``
	tmpContext := `root = true

[*]
indent_style = tab
indent_size = 4
end_of_line = lf
charset = utf-8
trim_trailing_whitespace = true
insert_final_newline = true

[*.md]
trim_trailing_whitespace = false

[Makefile]
indent_style = tab
indent_size = 4

`

	context := tmpContext
	fn := GetTargetFileName(PD, "editor", rootdir)
	return writeContext(fn, header, context, false)
}
