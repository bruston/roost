package runtime

import (
	"errors"
	"io"
	"os"

	"github.com/bruston/roost/types"
)

type Value interface {
	Type() types.ValueType
	Value() interface{}
}

type Stack struct {
	data []Value
	top  int
}

func (s *Stack) Push(v Value) {
	s.top++
	s.data[s.top] = v
}

func (s *Stack) Pop() Value {
	v := s.data[s.top]
	s.data[s.top] = nil
	s.top--
	return v
}

func (s *Stack) Peek() Value { return s.data[s.top] }

func (s *Stack) Dup() { s.Push(s.data[s.top]) }

func (s *Stack) Drop() { s.Pop() }

func (s *Stack) Swap() {
	s.data[s.top], s.data[s.top-1] = s.data[s.top-1], s.data[s.top]
}

func (s *Stack) PushBool(b bool) { s.Push(types.BoolValue{types.ValueBool, b}) }

func (s *Stack) PushNum(n float64) { s.Push(types.NewNum(n)) }

func (s *Stack) PushString(str string) { s.Push(types.NewString(str)) }

func (s *Stack) PushBlob(b []byte) { s.Push(&types.BlobValue{types.ValueBlob, b}) }

func (s *Stack) Len() int { return s.top + 1 }

func NewStack(size int) *Stack { return &Stack{data: make([]Value, size), top: -1} }

var ErrStackError = errors.New("stack under/overflow")

type FuncValue func(*Env)

type Env struct {
	Stack   *Stack
	Return  *Stack
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Builtin map[string]FuncValue
	Vars    map[string]Value
	Words   map[string]FuncValue
}

func New(stackSize int) *Env {
	return &Env{
		Stack:   NewStack(stackSize),
		Return:  NewStack(stackSize),
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		Builtin: Builtin,
		Vars:    make(map[string]Value),
		Words:   make(map[string]FuncValue),
	}
}
