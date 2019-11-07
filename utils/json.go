// Package json 使用开源第三方库json-iterator封装的json api，与标准库api完全一模一样，只需将import路径由encoding/json改成going/json即可。标准包使用了反射来实现，性能极低，使用json-iterator解码能提升5倍性能，编码也比标准包性能好，不过较不明显
package utils

import (
	"io"

	stdjson "encoding/json"

	"github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

var j = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	extra.RegisterFuzzyDecoders()
}

// Marshal 利用json-iterator进行json编码
func Marshal(v interface{}) ([]byte, error) {
	return j.Marshal(v)
}

//MarshalIndent MarshalIndent
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return j.MarshalIndent(v, prefix, indent)
}

// Unmarshal 利用json-iterator进行json解码
func Unmarshal(data []byte, v interface{}) error {
	return j.Unmarshal(data, v)
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *jsoniter.Encoder {
	return j.NewEncoder(w)
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *jsoniter.Decoder {
	return j.NewDecoder(r)
}

type RawMessage = stdjson.RawMessage