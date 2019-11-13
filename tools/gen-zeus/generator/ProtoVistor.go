package generator

import "github.com/emicklei/proto"

type ProtoVistor struct {
	EnumFields []*proto.EnumField
}

func (p *ProtoVistor) VisitMessage(m *proto.Message) {
}
func (p *ProtoVistor) VisitService(v *proto.Service) {
}
func (p *ProtoVistor) VisitSyntax(s *proto.Syntax) {
}
func (p *ProtoVistor) VisitPackage(pkg *proto.Package) {
}
func (p *ProtoVistor) VisitOption(o *proto.Option) {
}
func (p *ProtoVistor) VisitImport(i *proto.Import) {
}
func (p *ProtoVistor) VisitNormalField(i *proto.NormalField) {
}
func (p *ProtoVistor) VisitEnumField(i *proto.EnumField) {
	p.EnumFields = append(p.EnumFields, i)
}
func (p *ProtoVistor) VisitEnum(e *proto.Enum) {
}
func (p *ProtoVistor) VisitComment(e *proto.Comment) {}
func (p *ProtoVistor) VisitOneof(o *proto.Oneof) {
}
func (p *ProtoVistor) VisitOneofField(o *proto.OneOfField) {
}
func (p *ProtoVistor) VisitReserved(rs *proto.Reserved) {
}
func (p *ProtoVistor) VisitRPC(rpc *proto.RPC) {
}
func (p *ProtoVistor) VisitMapField(f *proto.MapField) {
}
func (p *ProtoVistor) VisitGroup(g *proto.Group) {
}
func (p *ProtoVistor) VisitExtensions(e *proto.Extensions) {
}
func (p *ProtoVistor) VisitProto(*proto.Proto) {}
