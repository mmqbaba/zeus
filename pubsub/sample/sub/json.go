package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/micro/go-micro/codec"

	zjson "gitlab.dg.com/BackEnd/deliver/tif/zeus/microsrv/gomicro/codec/json"
)

var jsonPBMarshaler = &jsonpb.Marshaler{
	OrigName: true,
}

type Codec struct {
	*zjson.Codec
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
		err := jsonpb.UnmarshalNext(c.Decoder, pb)
		if err != nil {
			log.Println(err)
		}
		return err
	}
	err := c.Decoder.Decode(b)
	if err != nil {
		log.Println(err)
	}
	return err
}

func NewCodec(c io.ReadWriteCloser) codec.Codec {
	return &Codec{
		Codec: &zjson.Codec{
			Conn:    c,
			Decoder: json.NewDecoder(c),
			Encoder: json.NewEncoder(c),
		},
	}
}
