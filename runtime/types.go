package runtime

import (
	"fmt"
	"io"
)

type ValueType int

const (
	ValueNum = iota
	ValueString
	ValueByte
	ValueList
	ValueSlice
	ValueBlob
	ValueBool
	ValueRef
	ValuePipe
)

func (vt ValueType) Type() ValueType { return vt }

type NumValue struct {
	ValueType
	Val float64
}

func (nv NumValue) Value() interface{} { return nv.Val }

func (nv NumValue) String() string { return fmt.Sprintf("%v", nv.Val) }

type Sizer interface {
	Len() Value
}

type StringValue struct {
	ValueType
	Val string
}

func (sv StringValue) Value() interface{} { return sv.Val }

func (sv StringValue) String() string { return sv.Val }

func (sv StringValue) Len() Value { return NumValue{ValueNum, float64(len(sv.Val))} }

type ByteValue struct {
	ValueType
	Val byte
}

func (bv ByteValue) Value() interface{} { return bv.Val }

type FuncValue func(env *Env)

type BoolValue struct {
	ValueType
	Val bool
}

func (bv BoolValue) Value() interface{} { return bv.Val }

func (bv BoolValue) String() string { return fmt.Sprintf("%v", bv.Val) }

type RefValue struct {
	ValueType
	Key string
}

func (rv RefValue) Value() interface{} { return rv.Key }

func (rv RefValue) String() string { return rv.Key }

type Collection interface {
	Value
	Insert(Value)
}

type Indexable interface {
	Index(Value) Value
}

type SliceValue struct {
	ValueType
	Val []Value
}

func (vv *SliceValue) Value() interface{} { return vv.Val }

func (vv *SliceValue) Insert(v Value) { vv.Val = append(vv.Val, v) }

func (vv *SliceValue) Index(v Value) Value {
	switch n := v.(type) {
	case NumValue:
		return vv.Val[int(n.Val)]
	case *SliceValue:
		if len(n.Val) < 2 {
			return nil
		}
		start, _ := n.Val[0].(NumValue)
		end, _ := n.Val[1].(NumValue)
		return &SliceValue{
			ValueSlice,
			vv.Val[int(start.Val):int(end.Val)],
		}
	}
	return nil
}

func (vv *SliceValue) Len() Value { return NumValue{ValueNum, float64(len(vv.Val))} }

type BlobValue struct {
	ValueType
	Val []byte
}

func (bv *BlobValue) Value() interface{} { return bv.Val }

func (bv *BlobValue) Len() Value { return NumValue{ValueNum, float64(len(bv.Val))} }

func (bv *BlobValue) Insert(v Value) {
	if b, ok := v.(ByteValue); ok {
		bv.Val = append(bv.Val, b.Val)
	}
}

func (bv *BlobValue) Index(v Value) Value {
	switch n := v.(type) {
	case NumValue:
		return ByteValue{ValueByte, bv.Val[int(n.Val)]}
	case *SliceValue:
		if len(n.Val) < 2 {
			return nil
		}
		start, _ := n.Val[0].(NumValue)
		end, _ := n.Val[1].(NumValue)
		return &BlobValue{ValueBlob, bv.Val[int(start.Val):int(end.Val)]}
	}
	return nil
}

type Iterable interface {
	Iter(Value)
}

type PipeValue struct {
	ValueType
	rwc io.ReadWriteCloser
}

func (pv *PipeValue) Value() interface{} { return pv.rwc }

func (pv *PipeValue) Type() ValueType { return pv.ValueType }

func (fp *PipeValue) Write(v Value) (float64, error) {
	var b []byte
	switch p := v.(type) {
	case StringValue:
		b = []byte(p.Val)
	case *BlobValue:
		b = p.Val
	}
	n, err := fp.rwc.Write(b)
	return float64(n), err
}

func (pv *PipeValue) Read(b []byte) (float64, error) {
	return 0, nil
}

func (fp *PipeValue) Close() error { return fp.rwc.Close() }
