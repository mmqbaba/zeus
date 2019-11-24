package errors

import (
	"errors"
	"strconv"
	"testing"
)

type custError1 struct{}

func (ce custError1) Error() string {
	return "custError1"
}

type stringError string

func (serr stringError) Error() string {
	return "stringError:" + string(serr)
}

type intError int

func (ierr intError) Error() string {
	return "intError:" + strconv.Itoa(int(ierr))
}

type boolError bool

func (berr boolError) Error() string {
	return "boolError:" + strconv.FormatBool(bool(berr))
}

func TestNilAssertError(t *testing.T) {
	tests := []struct {
		input error
		want  *Error
	}{
		{nil, nil},
		{(*Error)(nil), nil},
		{(*custError1)(nil), nil},
		{stringError(""), nil},
		{intError(0), nil},
		{boolError(false), nil},
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
		want  *Error
	}{
		{errors.New("not nil"), &Error{ErrCode: ECodeSystem}},
		{&Error{ErrCode: ECodeSuccessed}, &Error{ErrCode: ECodeSuccessed}},
		{&custError1{}, &Error{ErrCode: ECodeSystem}},
		{stringError("serror"), &Error{ErrCode: ECodeSystem}},
		{intError(1), &Error{ErrCode: ECodeSystem}},
		{boolError(true), &Error{ErrCode: ECodeSystem}},
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
