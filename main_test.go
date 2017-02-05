package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/bruston/roost/parser"
	"github.com/bruston/roost/runtime"
)

func TestEndToEnd(t *testing.T) {
	for i, tt := range []struct {
		code     string
		expected string
		err      error
	}{
		{`"Hello, World!" .`, "Hello, World!", nil},
		{"5 5 + .", "10", nil},
		{"10 3 - .", "7", nil},
		{"5 10 * . ", "50", nil},
		{"6 2 / .", "3", nil},
		{"6 2 % .", "0", nil},
		{"5 2 % .", "1", nil},
		{"10 0 for I . end", "0123456789", nil},
		{`: hello "Hello, World!" . ; hello`, "Hello, World!", nil},
		{`{ "foo" "bar" "baz" } 1 # .`, "bar", nil},
		{`1 1 = if "foo" else "bar" then .`, "foo", nil},
		{`{ "foo" "bar" "baz" } len 0 for I # . end`, "foobarbaz", nil},
		{`{ "foo" } "bar" insert 0 # . 1 # .`, "foobar", nil},
		{`1 0 = if "foo" else "bar" then .`, "bar", nil},
		{`var foo "bar" ! "foo" foo @ .`, "bar", nil},
		{`.`, "", runtime.ErrStackError},
	} {
		p := parser.New(strings.NewReader(tt.code))
		ast, err := p.Parse()
		if err != nil {
			t.Errorf("%d. error parsing: %s\n%s", i, tt.code, err)
			continue
		}
		env := runtime.New(1024)
		buf := &bytes.Buffer{}
		env.Stdout = buf
		if err := runtime.Eval(env, ast); err != tt.err {
			t.Errorf("%d. code %s\nshould produce error: %v but received: %v", i, tt.code, tt.err, err)
		}
		if buf.String() != tt.expected {
			t.Errorf("%d. code: %s\nshould produce output: %s\nbut received: %s", i, tt.code, tt.expected, buf.String())
		}
	}
}
