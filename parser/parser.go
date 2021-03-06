package parser

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/bruston/roost/lexer"
	"github.com/bruston/roost/runtime"
	"github.com/bruston/roost/types"
)

type Parser struct {
	scn           *lexer.Scanner
	tree          []Node
	currentParent Appendable
}

func New(r io.Reader) *Parser {
	p := &Parser{scn: lexer.NewScanner(r)}
	return p
}

type Node interface {
}

type NodeWord struct{ Identifier string }

func (nw NodeWord) String() string { return nw.Identifier }

type Appendable interface {
	Append(Node)
	Parent() Appendable
}

type NodeWordDef struct {
	Identifier string
	Body       []Node
	parent     Appendable
}

func (nw *NodeWordDef) Append(node Node) { nw.Body = append(nw.Body, node) }

func (nw *NodeWordDef) Parent() Appendable { return nw.parent }

type NodeVarDef struct{ Identifier string }

type NodeNumLit struct{ Value float64 }

func (nl NodeNumLit) String() string { return fmt.Sprintf("%f", nl.Value) }

type NodeStringLit struct{ Value string }

func (ns NodeStringLit) String() string { return ns.Value }

type CollectionType int

const (
	ListCollection CollectionType = iota
	SliceCollection
)

type BlobNode struct {
	parent Appendable
	body   []Node
}

func (b *BlobNode) Append(node Node) {
	b.body = append(b.body, node)
}

func (b *BlobNode) Parent() Appendable { return b.parent }

type NodeCollection struct {
	Type   CollectionType
	Body   []Node
	parent Appendable
}

func (nc *NodeCollection) Append(node Node) { nc.Body = append(nc.Body, node) }

func (nc *NodeCollection) Parent() Appendable { return nc.parent }

type NodeIf struct {
	Body   []Node
	Else   *NodeElse
	parent Appendable
}

type NodeElse struct {
	Body   []Node
	parent Appendable
}

func (ne *NodeElse) Append(node Node) { ne.Body = append(ne.Body, node) }

func (ne *NodeElse) Parent() Appendable { return ne.parent }

func (ne *NodeIf) Append(node Node) { ne.Body = append(ne.Body, node) }

func (ne *NodeIf) Parent() Appendable { return ne.parent }

type NodeRef string

func (nr NodeRef) String() string { return string(nr) }

func (nr NodeRef) Value() interface{} { return string(nr) }

func (p *Parser) insertNode(node Node) {
	if p.currentParent != nil {
		p.currentParent.Append(node)
		return
	}
	p.tree = append(p.tree, node)
}

type NodeFor struct {
	Body   []Node
	parent Appendable
}

func (nf *NodeFor) Append(node Node) { nf.Body = append(nf.Body, node) }

func (nf *NodeFor) Parent() Appendable { return nf.parent }

func (p *Parser) Parse() ([]Node, error) {
	for p.scn.Scan() {
		token := p.scn.Token()
		switch token.Type {
		case lexer.EOF:
			break
		case lexer.String:
			p.insertNode(NodeStringLit{Value: token.Value})
		case lexer.Number:
			n, _ := strconv.ParseFloat(token.Value, 64)
			p.insertNode(NodeNumLit{Value: n})
		case lexer.Word:
			if token.Value[0] == '&' && len(token.Value) > 1 {
				p.insertNode(NodeRef(token.Value[1:]))
				continue
			}
			p.insertNode(NodeWord{Identifier: token.Value})
		case lexer.Colon:
			name := p.scn.Token()
			if name.Type != lexer.Word {
				return nil, fmt.Errorf("expecting word after semicolon, got: %v", name)
			}
			node := &NodeWordDef{Identifier: name.Value}
			node.parent = p.currentParent
			p.insertNode(node)
			p.currentParent = node
		case lexer.Semicolon:
			if p.currentParent == nil {
				return nil, errors.New("unexpected semicolon outside of word definition")
			}
			p.tree = append(p.tree, p.currentParent)
			p.currentParent = p.currentParent.Parent()
		case lexer.Var:
			name := p.scn.Token()
			if name.Type != lexer.Word {
				return nil, fmt.Errorf("expecting word after var, got: %v", name)
			}
			node := NodeVarDef{Identifier: name.Value}
			p.insertNode(node)
		case lexer.If:
			node := &NodeIf{
				parent: p.currentParent,
				Else:   &NodeElse{parent: p.currentParent},
			}
			p.insertNode(node)
			p.currentParent = node
		case lexer.Else:
			node, ok := p.currentParent.(*NodeIf)
			if !ok {
				return nil, errors.New("expecting else to be inside if")
			}
			p.currentParent = node.Else
		case lexer.Then:
			if p.currentParent == nil {
				return nil, fmt.Errorf("unexpected then")
			}
			p.currentParent = p.currentParent.Parent()
		case lexer.For:
			node := &NodeFor{parent: p.currentParent}
			p.insertNode(node)
			p.currentParent = node
		case lexer.End:
			if p.currentParent == nil {
				return nil, errors.New("unexpected end")
			}
			p.currentParent = p.currentParent.Parent()
		case lexer.BracketOpen:
			node := &NodeCollection{
				Type:   ListCollection,
				parent: p.currentParent,
			}
			p.insertNode(node)
			p.currentParent = node
		case lexer.BraceOpen:
			node := &NodeCollection{
				Type:   SliceCollection,
				parent: p.currentParent,
			}
			p.insertNode(node)
			p.currentParent = node
		case lexer.BracketClose, lexer.BraceClose:
			p.currentParent = p.currentParent.Parent()
		case lexer.ParenOpen:
		case lexer.ParenClose:
		}
	}
	return p.tree, nil
}

type Visitor interface {
	Visit(Node) Visitor
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}
}

func funcFromDef(ev *Evaluator, n *NodeWordDef) runtime.FuncValue {
	return runtime.FuncValue(func(e *runtime.Env) {
		for _, c := range n.Body {
			Walk(ev, c)
		}
	})
}

func (ev *Evaluator) Visit(node Node) Visitor {
	defer func() {
		if r := recover(); r != nil {
			ev.err = runtime.ErrStackError
		}
	}()
	switch n := node.(type) {
	case *NodeWordDef:
		ev.env.Words[n.Identifier] = funcFromDef(ev, n)
	case NodeWord:
		if word, ok := ev.env.Words[n.Identifier]; ok {
			word(ev.env)
			return ev
		}
		if word, ok := ev.env.Builtin[n.Identifier]; ok {
			word(ev.env)
		}
	case NodeStringLit:
		ev.env.Stack.PushString(n.Value)
	case NodeNumLit:
		ev.env.Stack.PushNum(n.Value)
	case NodeVarDef:
		ref := NodeRef(n.Identifier)
		ev.env.Words[n.Identifier] = funcFromDef(ev, &NodeWordDef{
			Identifier: n.Identifier,
			Body:       []Node{ref},
		})
		ev.env.Stack.Push(types.NewRef(n.Identifier))
	case NodeRef:
		ev.env.Stack.Push(types.NewRef(string(n)))
	case *NodeIf:
		cond := ev.env.Stack.Pop()
		if cond.Value() == true || cond.Value() == 1 {
			for _, c := range n.Body {
				Walk(ev, c)
			}
			return ev
		}
		for _, c := range n.Else.Body {
			Walk(ev, c)
		}
	case *NodeFor:
		ev.env.Stack.Swap()
		ev.env.Return.Push(ev.env.Stack.Pop())
		ev.env.Return.Push(ev.env.Stack.Pop())
		for {
			index := ev.env.Return.Pop()
			limit := ev.env.Return.Peek()
			if index.Type() != types.ValueNum || limit.Type() != types.ValueNum {
				return ev
			}
			if index.Value().(float64) < limit.Value().(float64) || limit.Value().(float64) == 0 {
				ev.env.Return.Push(index)
				for _, c := range n.Body {
					Walk(ev, c)
				}
				ev.env.Return.Drop()
				ev.env.Return.PushNum(index.Value().(float64) + 1)
				continue
			}
			break
		}
	case *NodeCollection:
		collection := newCollection(n.Type)
		for _, node := range n.Body {
			collection.Insert(ev.evalNode(node))
		}
		ev.env.Stack.Push(collection)
	}
	return ev
}

func newCollection(n CollectionType) types.Collection {
	if n == SliceCollection {
		return &types.SliceValue{ValueType: types.ValueSlice}
	}
	return nil
}

func (ev *Evaluator) evalNode(node Node) types.Value {
	switch n := node.(type) {
	case NodeNumLit:
		return types.NewNum(n.Value)
	case NodeStringLit:
		return types.NewString(n.Value)
	case *NodeCollection:
		collection := newCollection(n.Type)
		for _, c := range n.Body {
			collection.Insert(ev.evalNode(c))
		}
		return collection
	case NodeWord:
		if n.Identifier == "true" {
			return types.NewBool(true)
		}
		if n.Identifier == "false" {
			return types.NewBool(false)
		}
	case NodeRef:
		// TODO
	}
	return nil
}

func Eval(env *runtime.Env, ast []Node) error {
	eval := &Evaluator{
		env: env,
	}
	for _, node := range ast {
		if eval.err != nil {
			break
		}
		Walk(eval, node)
	}
	return eval.err
}

type Evaluator struct {
	env *runtime.Env
	err error
}
