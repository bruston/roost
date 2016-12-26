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
	}{
		{`"Hello, World!" .`, "Hello, World!"},
		{"5 5 + .", "10"},
		{"10 3 - .", "7"},
		{"5 10 * . ", "50"},
		{"6 2 / .", "3"},
		{"6 2 % .", "0"},
		{"5 2 % .", "1"},
		{"10 0 for I . end", "0123456789"},
		{`: hello "Hello, World!" . ; hello`, "Hello, World!"},
		{`{ "foo" "bar" "baz" } 1 # .`, "bar"},
		{`1 1 = if "foo" else "bar" then .`, "foo"},
		{`{ "foo" "bar" "baz" } len 0 for I # . end`, "foobarbaz"},
		{`{ "foo" } "bar" insert 0 # . 1 # .`, "foobar"},
		{`1 0 = if "foo" else "bar" then .`, "bar"},
		{`var foo "bar" ! "foo" foo @ .`, "bar"},
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
		runtime.Eval(env, ast)
		if buf.String() != tt.expected {
			t.Errorf("%d. code: %s\nshould produce output: %s\nbut received: %s", i, tt.code, tt.expected, buf.String())
		}
	}
}
