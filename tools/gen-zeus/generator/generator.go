package generator

import (
	"github.com/emicklei/proto"
	"io"
	"math/rand"
	"strings"
	"time"
)

func New(reader io.Reader) (*Generator, error) {
	t := &Generator{
		FuncCli:   make(map[string]int),
		FuncSvr:   make(map[string]int),
		FuncLogic: make(map[string]int),
		SvrPort:   generateRangeNum(30000, 39999),
	}

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		return t, err
	}
	proto.Walk(definition,
		proto.WithService(t.handleService),
		proto.WithPackage(t.handlePackage),
		proto.WithOption(t.handleOption),
		proto.WithRPC(t.handleRpc),
		proto.WithEnum(t.handleEnum))

	if len(t.GopackageName) == 0 {
		t.GopackageName = t.PackageName
	}
	return t, nil
}

func (t *Generator) handlePackage(p *proto.Package) {
	t.PackageName = p.Name
}

func (t *Generator) handleService(s *proto.Service) {
	t.SvrName = s.Name
}

func (t *Generator) handleOption(o *proto.Option) {
	if o.Name == "go_package" {
		t.GopackageName = o.Constant.Source
	}
}

func (t *Generator) handleRpc(m *proto.RPC) {
	var p = RpcItem{
		Name:             m.Name,
		RequestType:      m.RequestType,
		ReturnsType:      m.ReturnsType,
		IsStreamsRequest: m.StreamsRequest,
		IsPost:           true,
	}
	for _, v := range m.Options {
		//fmt.Println(v.Name)
		//fmt.Println(v.Constant)
		if v.Name == "(google.api.http)" {
			for k2, _ := range v.Constant.Map {
				if k2 == "get" {
					p.IsPost = false
					break
				}
			}
			if !p.IsPost {
				break
			}
		}
	}
	t.Rpcapi = append(t.Rpcapi, p)
}

func (t *Generator) handleEnum(e *proto.Enum) {
	if strings.HasSuffix(e.Name, "ErrCode") {
		//log.Info(*e)
		var pv ProtoVistor
		for _, ei := range e.Elements {
			ei.Accept(&pv)
		}
		t.ErrCodes = append(t.ErrCodes, ErrCodeDef{ErrCodeSetName: e.Name, ErrCodeEnums: pv.EnumFields})
	}
}

func generateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max-min) + min
	return randNum
}
