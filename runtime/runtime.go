package runtime

import (
	"io"
	"os"

	"github.com/bruston/roost/parser"
)

type Value interface {
	Type() ValueType
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

func (s *Stack) PushBool(b bool) { s.Push(BoolValue{ValueBool, b}) }

func (s *Stack) PushNum(n float64) { s.Push(NumValue{ValueNum, n}) }

func (s *Stack) PushString(str string) { s.Push(StringValue{ValueString, str}) }

func (s *Stack) PushBlob(b []byte) { s.Push(&BlobValue{ValueBlob, b}) }

func NewStack(size int) *Stack { return &Stack{data: make([]Value, size), top: -1} }

type Env struct {
	Stack   *Stack
	Return  *Stack
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Builtin map[string]FuncValue
	Vars    map[string]Value
	Words   map[string]*parser.NodeWordDef
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
		Words:   make(map[string]*parser.NodeWordDef),
	}
}

func (ev *Evaluator) Visit(node parser.Node) parser.Visitor {
	switch n := node.(type) {
	case *parser.NodeWordDef:
		ev.env.Words[n.Identifier] = n
	case parser.NodeWord:
		if word, ok := ev.env.Words[n.Identifier]; ok {
			for _, c := range word.Body {
				parser.Walk(ev, c)
			}
			return ev
		}
		if word, ok := ev.env.Builtin[n.Identifier]; ok {
			word(ev.env)
		}
	case parser.NodeStringLit:
		ev.env.Stack.PushString(n.Value)
	case parser.NodeNumLit:
		ev.env.Stack.PushNum(n.Value)
	case parser.NodeVarDef:
		ref := parser.NodeRef(n.Identifier)
		ev.env.Words[n.Identifier] = &parser.NodeWordDef{
			Identifier: n.Identifier,
			Body:       []parser.Node{ref},
		}
		ev.env.Stack.Push(RefValue{ValueRef, n.Identifier})
	case parser.NodeRef:
		ev.env.Stack.Push(RefValue{ValueRef, string(n)})
	case *parser.NodeIf:
		cond := ev.env.Stack.Pop()
		if cond.Value() == true || cond.Value() == 1 {
			for _, c := range n.Body {
				parser.Walk(ev, c)
			}
			return ev
		}
		for _, c := range n.Else.Body {
			parser.Walk(ev, c)
		}
	case *parser.NodeFor:
		ev.env.Stack.Swap()
		ev.env.Return.Push(ev.env.Stack.Pop())
		ev.env.Return.Push(ev.env.Stack.Pop())
		for {
			index := ev.env.Return.Pop()
			limit := ev.env.Return.Peek()
			if index.Type() != ValueNum || limit.Type() != ValueNum {
				return ev
			}
			if index.Value().(float64) < limit.Value().(float64) || limit.Value().(float64) == 0 {
				ev.env.Return.Push(index)
				for _, c := range n.Body {
					parser.Walk(ev, c)
				}
				ev.env.Return.Drop()
				ev.env.Return.PushNum(index.Value().(float64) + 1)
				continue
			}
			break
		}
	case *parser.NodeCollection:
		collection := newCollection(n.Type)
		for _, node := range n.Body {
			collection.Insert(ev.evalNode(node))
		}
		ev.env.Stack.Push(collection)
	}
	return ev
}

func newCollection(n parser.CollectionType) Collection {
	if n == parser.VectorCollection {
		return &VectorValue{ValueType: ValueVector}
	}
	if n == parser.ListCollection {
		return &ListValue{ValueType: ValueList}
	}
	return nil
}

func (ev *Evaluator) evalNode(node parser.Node) Value {
	switch n := node.(type) {
	case parser.NodeNumLit:
		return NumValue{ValueNum, n.Value}
	case parser.NodeStringLit:
		return StringValue{ValueString, n.Value}
	case *parser.NodeCollection:
		collection := newCollection(n.Type)
		for _, c := range n.Body {
			collection.Insert(ev.evalNode(c))
		}
		return collection
	case parser.NodeWord:
		if n.Identifier == "true" {
			return BoolValue{ValueBool, true}
		}
		if n.Identifier == "false" {
			return BoolValue{ValueBool, false}
		}
	case parser.NodeRef:
		// TODO
	}
	return nil
}

func Eval(env *Env, ast []parser.Node) {
	eval := &Evaluator{
		env: env,
	}
	for _, node := range ast {
		parser.Walk(eval, node)
	}
}

type Evaluator struct {
	env *Env
}
