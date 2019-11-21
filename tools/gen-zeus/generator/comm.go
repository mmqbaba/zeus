package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

	"github.com/emicklei/proto"
)

const (
	_defaultHeader = `// Code generated by zeus-gen. DO NOT EDIT.
`
	_errdefFileName = "errdef.proto"
)

var (
	projectBasePrefix = "zeus_app/"
)

func SetProjectBasePrefix(base string) {
	projectBasePrefix = base
}

type ErrCodeDef struct {
	ErrCodeSetName string
	ErrCodeEnums   []*proto.EnumField
}

type RpcItem struct {
	Name             string
	RequestType      string
	ReturnsType      string
	IsPost           bool
	Method           string
	IsStreamsRequest bool
	ApiPath          string
}

type Generator struct {
	PackageName   string
	SvrName       string
	GopackageName string
	Rpcapi        []RpcItem
	ErrCodes      []ErrCodeDef
	SvrPort       int
	Imports       []string

	SvrDef map[string]string

	FuncSvr   map[string]int
	FuncCli   map[string]int
	FuncLogic map[string]int
}

type ServerComposement struct {
	mainL   token.Pos
	mainR   token.Pos
	imports []string
}

func GetTargetFileName(PD *Generator, objtype string, rootdir string, opts ...string) string {
	var fn string
	switch objtype {
	case "errdef":
		dirname := fmt.Sprintf("%s/%s/errdef", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = dirname + "/errdef.go"
	case "errdef.enum":
		//dirname := fmt.Sprintf("%s/%s/errdef", rootdir, PD.PackageName)
		dirname := fmt.Sprintf("%s/errdef", rootdir)
		CheckPrepareDir(dirname)
		fn = dirname + "/errdef.go"
	case "cmd.init":
		dirname := fmt.Sprintf("%s/%s/cmd/app", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/init.go", dirname)
	case "cmd.main":
		dirname := fmt.Sprintf("%s/%s/cmd/app", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/main.go", dirname)
	case "global.init":
		dirname := fmt.Sprintf("%s/%s/global", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/init.go", dirname)
	case "global.enum":
		dirname := fmt.Sprintf("%s/%s/global", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/enum.go", dirname)
	case "global":
		dirname := fmt.Sprintf("%s/%s/global", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/global.go", dirname)
	case "handler.comm":
		dirname := fmt.Sprintf("%s/%s/handler", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/%s.go", dirname, PD.PackageName)
	case "handler":
		dirname := fmt.Sprintf("%s/%s/handler", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		if len(opts) > 0 {
			fn = fmt.Sprintf("%s/%s_%s.go", dirname, PD.PackageName, opts[0])
		} else {
			fn = fmt.Sprintf("%s/%s.go", dirname, PD.PackageName)
		}
	case "http.init":
		dirname := fmt.Sprintf("%s/%s/http", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/init.go", dirname)
	case "http":
		dirname := fmt.Sprintf("%s/%s/http", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/http.go", dirname)
	case "resource":
		dirname := fmt.Sprintf("%s/%s/resource/rpc", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/rpc.go", dirname)
	case "rpc.init":
		dirname := fmt.Sprintf("%s/%s/rpc", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/init.go", dirname)
	case "proto.gw":
		dirname := fmt.Sprintf("%s/%s/proto/%spb", rootdir, PD.PackageName, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = dirname
	case "proto.gomicro":
		dirname := fmt.Sprintf("%s/%s/proto/%spb", rootdir, PD.PackageName, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = dirname
	case "proto.errdef":
		dirname := fmt.Sprintf("%s/proto", rootdir)
		CheckPrepareDir(dirname)
		fn = dirname + "/errdef.proto"
	case "proto":
		dirname := fmt.Sprintf("%s/%s/proto", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = dirname
	case "build.proto":
		dirname := fmt.Sprintf("%s/%s", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/build-proto.sh", dirname)
	case "go.mod":
		dirname := fmt.Sprintf("%s", rootdir)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/go.mod", dirname)
	case "readme":
		dirname := fmt.Sprintf("%s/%s", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/README.md", dirname)
	case "makefile":
		dirname := fmt.Sprintf("%s/%s", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/Makefile", dirname)
	case "logic":
		dirname := fmt.Sprintf("%s/%s/logic", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = dirname
	case "resource.dao":
		dirname := fmt.Sprintf("%s/%s/resource/dao", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = dirname + "/dao.go"
	case "resource.cache":
		dirname := fmt.Sprintf("%s/%s/resource/cache", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = dirname + "/cache.go"
	case "resource.rpcclient":
		dirname := fmt.Sprintf("%s/%s/resource/rpcclient", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = dirname + "/rpcclient.go"
	case "resource.httpclient":
		dirname := fmt.Sprintf("%s/%s/resource/httpclient", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = dirname + "/httpclient.go"
	case "dockerfile":
		dirname := fmt.Sprintf("%s/%s", rootdir, PD.PackageName)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/Dockerfile", dirname)
	case "zeus.errdef":
		dirname := fmt.Sprintf("%s/errors", rootdir)
		CheckPrepareDir(dirname)
		fn = fmt.Sprintf("%s/errdef.go", dirname)
	default:
		fmt.Print("Can not support object type")
		return ""
	}

	return fn
}

func ParseGoCode(PD *Generator, fn string, objtype string) {

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fn, nil, 0)
	if err != nil {
		fmt.Printf("Parse %s name , error : %v\n", objtype, err)
		return
	}

	for _, decl := range f.Decls {
		switch t := decl.(type) {
		// That's a func decl !
		case *ast.FuncDecl:
			if objtype == "client" {
				PD.FuncCli[t.Name.Name] = 1
			} else if objtype == "logic" {
				PD.FuncLogic[t.Name.Name] = 1
			} else if objtype == "server" {
				PD.FuncSvr[t.Name.Name] = 1
			}
		case *ast.GenDecl:
			if objtype == "def" {
				for _, spec := range t.Specs {
					switch spec := spec.(type) {
					case *ast.ImportSpec:
						fmt.Println("Import", spec.Path.Value)
					case *ast.TypeSpec:
						fmt.Println("Type", spec.Name.String())
					case *ast.ValueSpec:
						for _, id := range spec.Names {
							if id.Name == "SetSvrPort" || id.Name == "SvrAddrPort" {
								continue
							}

							PD.SvrDef[id.Name] = id.Obj.Decl.(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value
						}
					default:
						fmt.Printf("Unknown token type: %s\n", t.Tok)
					}
				}
			}
		default:
			fmt.Printf("decl type,%v", t)

			// Now I would like to access to the test var and get its type
			// Must be int, because foo() returns an int
		}
	}
}

func ParseGoServer(s *ServerComposement, fn string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fn, nil, 0)
	if err != nil {
		fmt.Printf("Parse  error : %v\n", err)
		return
	}
	if s.imports == nil {
		s.imports = make([]string, 0)
	}
	for _, decl := range f.Decls {
		switch t := decl.(type) {
		// That's a func decl !
		case *ast.FuncDecl:
			if t.Name.Name == "main" {
				s.mainL = t.Body.Lbrace
				s.mainR = t.Body.Rbrace
			}
		case *ast.GenDecl:
			for _, spec := range t.Specs {
				switch v := spec.(type) {
				case *ast.ImportSpec:
					s.imports = append(s.imports, v.Path.Value)
				case *ast.ValueSpec:
				case *ast.TypeSpec:
				default:
					log.Println(v)
				}
			}

		default:
			fmt.Printf("decl type,%v", t)

			// Now I would like to access to the test var and get its type
			// Must be int, because foo() returns an int
		}
	}
}

// FileExists 检测文件是否存在
func FileExists(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func CheckPrepareDir(dir string) {
	stat, err := os.Stat(dir)
	if err != nil || !stat.IsDir() {
		err := os.MkdirAll(dir, 0744)
		if err != nil {
			fmt.Printf("can not create server directory(%s)(err:%v)\n", dir, err)
			return
		}
	}
	return
}

func writeContext(fn string, header string, context string, override bool) error {

	if !override && FileExists(fn) {
		return nil
	}
	//header += _defaultHeader

	var perm os.FileMode = 0644
	if strings.HasSuffix(fn, ".sh") {
		perm = 0777
	}
	f, err := os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		fmt.Printf("can not generate file %s,Error :%v\n", fn, err)
		return err
	}
	if len(header) > 0 {
		if _, err := f.Write([]byte(header)); err != nil {
			return err
		}
	}
	if _, err := f.Write([]byte(context)); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func baseName(name string) string {
	// First, find the last element
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	// Now drop the suffix
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[0:i]
	}
	return name
}

// Is c an ASCII lower-case letter?
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

// In short, _my_field_name_2 becomes XMyFieldName_2.
func CamelCase(s string) string {
	if s == "" {
		return ""
	}
	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		// Need a capital letter; drop the '_'.
		t = append(t, 'X')
		i++
	}
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		// Assume we have a letter now - if not, it's a bogus identifier.
		// The next word is a sequence of characters that must start upper case.
		if isASCIILower(c) {
			c ^= ' ' // Make it a capital letter.
		}
		t = append(t, c) // Guaranteed not lower case.
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}
