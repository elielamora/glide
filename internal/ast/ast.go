// Package ast defines the abstract syntax tree for the Glide language.
// Glide uses a Lisp-style S-expression AST where every node is either
// an Atom (a single token) or a List (a sequence of nodes).
package ast

import "strings"

// Node is the interface implemented by every AST node.
type Node interface {
	String() string
	IsAtom() bool
}

// Atom represents a single indivisible token: identifier, keyword,
// numeric literal, string literal, or operator.
type Atom struct {
	Value string
}

// NewAtom creates a new Atom node.
func NewAtom(v string) *Atom { return &Atom{Value: v} }

func (a *Atom) String() string { return a.Value }
func (a *Atom) IsAtom() bool   { return true }

// ListKind distinguishes the bracket style used to create a list,
// which affects transpiler behaviour.
type ListKind int

const (
	// ListParen is a parenthesised list: (...)
	ListParen ListKind = iota
	// ListBracket is a square-bracket list: [...] — slice/array literal or type
	ListBracket
	// ListBrace is a curly-brace list: {...} — map literal
	ListBrace
)

// List represents an S-expression — an ordered sequence of nodes.
type List struct {
	Kind  ListKind
	Nodes []Node
}

// NewList creates a new paren-style List.
func NewList(nodes ...Node) *List {
	return &List{Kind: ListParen, Nodes: nodes}
}

// NewBracketList creates a square-bracket list.
func NewBracketList(nodes ...Node) *List {
	return &List{Kind: ListBracket, Nodes: nodes}
}

// NewBraceList creates a curly-brace list.
func NewBraceList(nodes ...Node) *List {
	return &List{Kind: ListBrace, Nodes: nodes}
}

func (l *List) IsAtom() bool { return false }

func (l *List) String() string {
	open, close := "(", ")"
	switch l.Kind {
	case ListBracket:
		open, close = "[", "]"
	case ListBrace:
		open, close = "{", "}"
	}
	if len(l.Nodes) == 0 {
		return open + close
	}
	var sb strings.Builder
	sb.WriteString(open)
	for i, n := range l.Nodes {
		if i > 0 {
			sb.WriteRune(' ')
		}
		sb.WriteString(n.String())
	}
	sb.WriteString(close)
	return sb.String()
}

// Head returns the first element of the list, or nil for an empty list.
func (l *List) Head() Node {
	if len(l.Nodes) == 0 {
		return nil
	}
	return l.Nodes[0]
}

// Args returns all nodes after the first element (the arguments).
func (l *List) Args() []Node {
	if len(l.Nodes) <= 1 {
		return nil
	}
	return l.Nodes[1:]
}

// HeadValue returns the string value of the head atom, or "" if the head
// is not an atom or the list is empty.
func (l *List) HeadValue() string {
	if h := l.Head(); h != nil {
		if a, ok := h.(*Atom); ok {
			return a.Value
		}
	}
	return ""
}
