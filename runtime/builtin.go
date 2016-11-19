package runtime

import (
	"fmt"
	"net"
	"os"
)

var Builtin = map[string]FuncValue{
	"+": func(e *Env) {
		n2, n1 := e.Stack.Pop(), e.Stack.Pop()
		if n1.Type() == ValueNum && n2.Type() == ValueNum {
			e.Stack.PushNum(n1.Value().(float64) + n2.Value().(float64))
			return
		}
		if n1.Type() == ValueString && n2.Type() == ValueString {
			e.Stack.PushString(n1.Value().(string) + n2.Value().(string))
		}
	},
	"*": func(e *Env) {
		n2, n1 := e.Stack.Pop(), e.Stack.Pop()
		if n1.Type() == ValueNum && n2.Type() == ValueNum {
			e.Stack.PushNum(n1.Value().(float64) * n2.Value().(float64))
		}
	},
	"-": func(e *Env) {
		n2, n1 := e.Stack.Pop(), e.Stack.Pop()
		if n1.Type() == ValueNum && n2.Type() == ValueNum {
			e.Stack.PushNum(n1.Value().(float64) - n2.Value().(float64))
		}
	},
	"/": func(e *Env) {
		n2, n1 := e.Stack.Pop(), e.Stack.Pop()
		if n1.Type() == ValueNum && n2.Type() == ValueNum {
			e.Stack.PushNum(n1.Value().(float64) / n2.Value().(float64))
		}
	},
	"%": func(e *Env) {
		n2, n1 := e.Stack.Pop(), e.Stack.Pop()
		if n1.Type() == ValueNum && n2.Type() == ValueNum {
			e.Stack.PushNum(float64(int64(n1.Value().(float64)) % int64(n2.Value().(float64))))
		}
	},
	"dup":  func(e *Env) { e.Stack.Dup() },
	"drop": func(e *Env) { e.Stack.Drop() },
	".": func(e *Env) {
		n := e.Stack.Pop()
		if n == nil {
			fmt.Fprintf(e.Stdout, "<nil>")
			return
		}
		fmt.Fprintf(e.Stdout, "%s", n)
	},
	"LF":    func(e *Env) { e.Stack.PushString("\n") },
	"CR":    func(e *Env) { e.Stack.PushString("\r") },
	"true":  func(e *Env) { e.Stack.PushBool(true) },
	"false": func(e *Env) { e.Stack.PushBool(false) },
	"<": func(e *Env) {
		n2, n1 := e.Stack.Pop(), e.Stack.Pop()
		if n1.Type() != ValueNum || n2.Type() != ValueNum {
			return
		}
		e.Stack.PushBool(n1.Value().(float64) < n2.Value().(float64))
	},
	">": func(e *Env) {
		n2, n1 := e.Stack.Pop(), e.Stack.Pop()
		if n1.Type() != ValueNum || n2.Type() != ValueNum {
			return
		}
		e.Stack.PushBool(n1.Value().(float64) > n2.Value().(float64))
	},
	"=": func(e *Env) {
		e.Stack.PushBool(e.Stack.Pop().Value() == e.Stack.Pop().Value())
	},
	"!": func(e *Env) {
		val, name := e.Stack.Pop(), e.Stack.Pop()
		if ref, ok := name.(RefValue); ok {
			e.Vars[ref.Key] = val
		}
	},
	"@": func(e *Env) {
		if ref, ok := e.Stack.Pop().(RefValue); ok {
			e.Stack.Push(e.Vars[ref.Key])
		}
	},
	"swap": func(e *Env) {
		if e.Stack.top < 1 {
			return
		}
		e.Stack.Swap()
	},
	"I": func(e *Env) { e.Stack.Push(e.Return.Peek()) },
	"map": func(e *Env) {
		ref, ok := e.Stack.Pop().(RefValue)
		if !ok {
			return
		}
		collection := e.Stack.Peek()
		var result Collection
		if collection.Type() == ValueList {
			result = &ListValue{ValueType: ValueList}
		}
		if collection.Type() == ValueVector {
			result = &VectorValue{ValueType: ValueVector}
		}
		iterator, ok := collection.(Iterable)
		if !ok {
			return
		}
		fn := func(v Value) {
			e.Stack.Push(v)
			Eval(e, e.Words[ref.Key].Body)
			result.Insert(e.Stack.Pop())
		}
		iterator.Iter(fn)
		e.Stack.Push(result)
	},
	"insert": func(e *Env) {
		val := e.Stack.Pop()
		if collection, ok := e.Stack.Peek().(Collection); ok {
			collection.Insert(val)
		}
	},
	"open": func(e *Env) {
		name, ok := e.Stack.Pop().(StringValue)
		if !ok {
			return
		}
		file, err := os.Open(name.Val)
		if err != nil {
			e.Stack.PushNum(1)
			return
		}
		pipe := &PipeValue{ValuePipe, file}
		e.Stack.Push(pipe)
		e.Stack.PushNum(0)
	},
	"dial": func(e *Env) {
		addr, ok := e.Stack.Pop().(StringValue)
		if !ok {
			return
		}
		prot, ok := e.Stack.Pop().(StringValue)
		if !ok {
			return
		}
		conn, err := net.Dial(prot.Val, addr.Val)
		if err != nil {
			e.Stack.PushString(err.Error())
			e.Stack.PushNum(1)
			return
		}
		pipe := &PipeValue{ValuePipe, conn}
		e.Stack.Push(pipe)
		e.Stack.PushNum(0)
	},
	"send": func(e *Env) {
		payload := e.Stack.Pop()
		p, ok := e.Stack.Peek().(*PipeValue)
		if !ok {
			return
		}
		n, err := p.Write(payload)
		if err != nil {
			e.Stack.PushString(err.Error())
			e.Stack.PushNum(1)
			return
		}
		e.Stack.PushNum(n)
		e.Stack.PushNum(0)
	},
	"close": func(e *Env) {
		pipe, ok := e.Stack.Peek().(*PipeValue)
		if !ok {
			return
		}
		if err := pipe.Close(); err != nil {
			e.Stack.PushString(err.Error())
			e.Stack.PushNum(1)
			return
		}
		e.Stack.PushNum(0)
		e.Stack.Drop()
	},
	"recv": func(e *Env) {
		arg, ok := e.Stack.Pop().(NumValue)
		if !ok {
			return
		}
		pipe, ok := e.Stack.Peek().(*PipeValue)
		if !ok {
			return
		}
		b := make([]byte, 0, int(arg.Val))
		n, err := pipe.Read(b)
		if err != nil {
			e.Stack.PushString(err.Error())
			e.Stack.PushNum(1)
			return
		}
		e.Stack.PushBlob(b)
		e.Stack.PushNum(n)
		e.Stack.PushNum(0)
	},
	"#": func(e *Env) {
		v := e.Stack.Pop()
		if indexable, ok := e.Stack.Peek().(Indexable); ok {
			e.Stack.Push(indexable.Index(v))
		}
	},
	"len": func(e *Env) {
		sizer, ok := e.Stack.Peek().(Sizer)
		if !ok {
			return
		}
		e.Stack.Push(sizer.Len())
	},
	"make": func(e *Env) {},
}
