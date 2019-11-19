package errors

import (
	"testing"
	"errors"
)

type custError1 struct {}
func (ce custError1) Error() string {
	return "custError1"
}

func TestNilAssertError(t *testing.T) {
	tests := []struct {
		input error
		want *Error
	}{
		{ nil, nil },
		{ (*Error)(nil), nil },
		{ (*custError1)(nil), nil },
	}
	for _, r := range tests {
		if res := AssertError(r.input); res != r.want {
			t.Errorf(`AssertError(%q) = %v, want(%v)`, r.input, res, r.want)
		}
	}
}

func TestNotNilAssertError(t *testing.T) {
	tests := []struct {
		input error
		want *Error
	}{
		{ errors.New("not nil"), &Error{ErrCode: ECodeSystem} },
		{ &Error{ErrCode: ECodeSuccessed}, &Error{ErrCode: ECodeSuccessed}},
		{ &custError1{}, &Error{ErrCode: ECodeSystem} },
	}
	for _, r := range tests {
		if res := AssertError(r.input); res.ErrCode != r.want.ErrCode {
			t.Errorf(`AssertError(%q) = %s, want(%s)`, r.input, res.ErrCode, r.want.ErrCode)
		}
	}
}

func BenchmarkNilAssertError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		AssertError((*custError1)(nil))
	}
}

func BenchmarkZeusAssertError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		AssertError(&Error{ErrCode: ECodeSuccessed})
	}
}