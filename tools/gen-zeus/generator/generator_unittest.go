package generator

import (
	"fmt"
	"strings"
)

func GenerateUnittest(PD *Generator, rootdir string) (err error) {
	err = genHandleUnittest(PD, rootdir)
	if err != nil {
		return err
	}
	err = genMainUnittest(PD, rootdir)
	if err != nil {
		return err
	}
	return
}

func genHandleUnittest(PD *Generator, rootdir string) error {
	header := ``
	tmpContext := `package handler

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/mmqbaba/zeus/plugin/container"
	"github.com/mmqbaba/zeus/service"

	"{PPKG}/global"
)

func TestMain(m *testing.M) {
	log.SetPrefix("[zeus_unittest] ")
	log.SetFlags(3)

	opt := service.Options{
		LogLevel:      "debug",
		ConfEntryPath: "../conf/zeus.json",
	}
	s := service.NewService(opt, container.GetContainer(), global.ServiceOpts...)
	if err := s.Init(); err != nil {
		log.Fatalf("Service s.Init exited with error: %s\n", err)
	}
	time.Sleep(1 * time.Second)
	code := m.Run()
	os.Exit(code)
}

`
	context := strings.ReplaceAll(tmpContext, "{PPKG}", projectBasePrefix+PD.PackageName)
	fn := GetTargetFileName(PD, "unittest.handler", rootdir)
	return writeContext(fn, header, context, false)
}

func genMainUnittest(PD *Generator, rootdir string) error {
	header := ``
	tmpContext := `package {PKG}

import (
    "bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mmqbaba/zeus/errors"
	"github.com/mmqbaba/zeus/plugin/container"
	"github.com/mmqbaba/zeus/service"
	"github.com/mmqbaba/zeus/utils"

	_ "{PPKG}/http"
	_ "{PPKG}/rpc"

	"{PPKG}/global"
    pb "{PPKG}/proto/{PKG}pb"
)

func TestMain(m *testing.M) {
	log.SetPrefix("[zeus_unittest] ")
	log.SetFlags(3)

	opt := service.Options{
		LogLevel:      "debug",
		ConfEntryPath: "./conf/zeus.json",
	}
	s := service.NewService(opt, container.GetContainer(), global.ServiceOpts...)
	if err := s.Init(); err != nil {
		log.Fatalf("Service s.Init exited with error: %s\n", err)
	}
	go func() {
		if err := s.RunServer(); err != nil {
			log.Fatalf("Service s.RunServer exited with error: %s\n", err)
		}
	}()
	time.Sleep(1 * time.Second)
	code := m.Run()
	time.Sleep(1 * time.Second)
	os.Exit(code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	global.GetNG().GetContainer().GetHTTPHandler().ServeHTTP(rr, req)
	return rr
}

func checkResponseStatusCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response statuscode %d. Got %d\n", expected, actual)
	}
}

func checkErrCode(t *testing.T, expected errors.ErrorCode, ret *errors.Error) {
	if ret.ErrCode != expected {
		t.Errorf("Expected response errcode %d. Got %d\n",
			expected, ret.ErrCode)
	}
}

func checkRsp(t *testing.T, expected, actual interface{}) {
	r, err := utils.Marshal(actual)
	if err != nil {
		t.Fatal(err)
	}
	w, err := utils.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	if string(r) != string(w) {
		t.Errorf("Expected response rsp %v. Got %v\n",
			expected, actual)
	}
}

func TestHTTP(t *testing.T) {
	// 路径带前缀'/api'
{HBLK}
}
func TestGoMicroGrpcGateway(t *testing.T) {
	// 路径不带前缀'/api'
{RBLK}
}

{FBLK}
`
	httpBlock := ""
	rpcBlock := ""
	funcBlock := ""
	runFuncFmt := `    run%s%s(t, %s, "%s")
`
	funcDefTmpl := `func run%s%s(t *testing.T, method, url string) {
	t.Run(url, func(t *testing.T) {
		type args struct {
			req interface{}
			ret *errors.Error
		}
		tests := []struct {
			name           string
			args           args
			wantStatusCode int
			wantErrCode    errors.ErrorCode
			wantRsp        interface{}
		}{
			// TODO: Add test cases.
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				jsonStr, _ := utils.Marshal(tt.args.req)
				req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonStr))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/json")
				rr := executeRequest(req)
				checkResponseStatusCode(t, tt.wantStatusCode, rr.Code)
				rsp := &pb.%s{}
				tt.args.ret.Data = rsp
				err = utils.Unmarshal(rr.Body.Bytes(), tt.args.ret)
				if err != nil {
					t.Fatal(err)
				}
				checkErrCode(t, tt.wantErrCode, tt.args.ret)
				checkRsp(t, tt.wantRsp, rsp)
			})
		}
	})
}

`
	camelSrvName := CamelCase(PD.SvrName)
	for _, v := range PD.Rpcapi {
		if v.ApiPath == "" {
			continue
		}
		httpBlock += fmt.Sprintf(runFuncFmt, camelSrvName, v.Name, v.Method, "/api"+v.ApiPath)
		rpcBlock += fmt.Sprintf(runFuncFmt, camelSrvName, v.Name, v.Method, v.ApiPath)

		funcBlock += fmt.Sprintf(funcDefTmpl, camelSrvName, v.Name, v.ReturnsType)
	}
	context := strings.ReplaceAll(tmpContext, "{PPKG}", projectBasePrefix+PD.PackageName)
	context = strings.ReplaceAll(context, "{PKG}", PD.PackageName)
	context = strings.ReplaceAll(context, "{HBLK}", httpBlock)
	context = strings.ReplaceAll(context, "{RBLK}", rpcBlock)
	context = strings.ReplaceAll(context, "{FBLK}", funcBlock)
	fn := GetTargetFileName(PD, "unittest.main", rootdir)
	return writeContext(fn, header, context, false)
}
