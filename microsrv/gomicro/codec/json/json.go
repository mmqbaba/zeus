// Package json provides a json codec
package json

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/micro/go-micro/codec"
)

var jsonPBMarshaler = &jsonpb.Marshaler{
	OrigName: true,
}

type Codec struct {
	Conn    io.ReadWriteCloser
	Encoder *json.Encoder
	Decoder *json.Decoder
}

func (c *Codec) ReadHeader(m *codec.Message, t codec.MessageType) error {
	return nil
}

func (c *Codec) ReadBody(b interface{}) error {
	if b == nil {
		return nil
	}
	if raw, ok := b.(*[]byte); ok {
		d, err := ioutil.ReadAll(c.Conn)
		if err != nil {
			return err
		}
		*raw = d
		return nil
	}
	if pb, ok := b.(proto.Message); ok {
		return jsonpb.UnmarshalNext(c.Decoder, pb)
	}
	return c.Decoder.Decode(b)
}

func (c *Codec) Write(m *codec.Message, b interface{}) error {
	if b == nil {
		return nil
	}
	if pb, ok := b.(proto.Message); ok {
		return jsonPBMarshaler.Marshal(c.Conn, pb)
	}
	return c.Encoder.Encode(b)
}

func (c *Codec) Close() error {
	return c.Conn.Close()
}

func (c *Codec) String() string {
	return "json"
}

func NewCodec(c io.ReadWriteCloser) codec.Codec {
	return &Codec{
		Conn:    c,
		Decoder: json.NewDecoder(c),
		Encoder: json.NewEncoder(c),
	}
}
