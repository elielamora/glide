package parser

import (
	"strings"
	"testing"

	"github.com/elielamora/glide/internal/ast"
)

// helper: parse a string and return the single top-level node.
func parseSingle(t *testing.T, src string) ast.Node {
	t.Helper()
	nodes, err := ParseString(src)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 top-level node, got %d", len(nodes))
	}
	return nodes[0]
}

func TestParse_atom(t *testing.T) {
	node := parseSingle(t, "42")
	a, ok := node.(*ast.Atom)
	if !ok {
		t.Fatalf("expected Atom, got %T", node)
	}
	if a.Value != "42" {
		t.Errorf("want 42, got %q", a.Value)
	}
}

func TestParse_flatList(t *testing.T) {
	node := parseSingle(t, "+ 1 2")
	lst, ok := node.(*ast.List)
	if !ok {
		t.Fatalf("expected List, got %T", node)
	}
	if lst.String() != "(+ 1 2)" {
		t.Errorf("want (+ 1 2), got %s", lst.String())
	}
}

func TestParse_inlineParen(t *testing.T) {
	node := parseSingle(t, "let result (+ 1 2)")
	lst, ok := node.(*ast.List)
	if !ok {
		t.Fatalf("expected List")
	}
	// (let result (+ 1 2))
	if len(lst.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d: %s", len(lst.Nodes), lst)
	}
	inner, ok := lst.Nodes[2].(*ast.List)
	if !ok {
		t.Fatalf("third node should be a List, got %T", lst.Nodes[2])
	}
	if inner.String() != "(+ 1 2)" {
		t.Errorf("inner: want (+ 1 2), got %s", inner)
	}
}

func TestParse_indentedChild(t *testing.T) {
	src := "+ 1\n    * 2 3"
	node := parseSingle(t, src)
	lst, ok := node.(*ast.List)
	if !ok {
		t.Fatalf("expected List, got %T", node)
	}
	// (+ 1 (* 2 3))
	if len(lst.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d: %s", len(lst.Nodes), lst)
	}
	inner, ok := lst.Nodes[2].(*ast.List)
	if !ok {
		t.Fatalf("third node should be List, got %T", lst.Nodes[2])
	}
	if inner.String() != "(* 2 3)" {
		t.Errorf("want (* 2 3), got %s", inner)
	}
}

func TestParse_multipleChildren(t *testing.T) {
	src := "fn body\n    let x 5\n    + x 3"
	nodes, err := ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 top-level node, got %d", len(nodes))
	}
	lst := nodes[0].(*ast.List)
	// (fn body (let x 5) (+ x 3))
	if len(lst.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d: %s", len(lst.Nodes), lst)
	}
}

func TestParse_nestedIndent(t *testing.T) {
	src := `let status
    if (> age 18)
        "Adult"
        "Minor"`
	node := parseSingle(t, src)
	lst := node.(*ast.List)
	// (let status (if (> age 18) "Adult" "Minor"))
	if len(lst.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d: %s", len(lst.Nodes), lst)
	}
	ifExpr := lst.Nodes[2].(*ast.List)
	if ifExpr.HeadValue() != "if" {
		t.Errorf("expected 'if' head, got %q", ifExpr.HeadValue())
	}
	if len(ifExpr.Nodes) != 4 {
		t.Fatalf("if node: expected 4 nodes, got %d: %s", len(ifExpr.Nodes), ifExpr)
	}
}

func TestParse_bracketList(t *testing.T) {
	node := parseSingle(t, `let names ["Alice" "Bob"]`)
	lst := node.(*ast.List)
	bracket := lst.Nodes[2].(*ast.List)
	if bracket.Kind != ast.ListBracket {
		t.Errorf("expected ListBracket, got %v", bracket.Kind)
	}
	if len(bracket.Nodes) != 2 {
		t.Errorf("expected 2 items, got %d", len(bracket.Nodes))
	}
}

func TestParse_multipleTopLevel(t *testing.T) {
	src := "let x 1\nlet y 2"
	nodes, err := ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestParse_trySuffix(t *testing.T) {
	src := "let content (.ReadFile os path)?"
	node := parseSingle(t, src)
	lst := node.(*ast.List)
	// (let content (try (.ReadFile os path)))
	if len(lst.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d: %s", len(lst.Nodes), lst)
	}
	tryExpr, ok := lst.Nodes[2].(*ast.List)
	if !ok {
		t.Fatalf("expected list at index 2, got %T: %s", lst.Nodes[2], lst.Nodes[2])
	}
	if tryExpr.HeadValue() != "try" {
		t.Errorf("expected 'try' head, got %q", tryExpr.HeadValue())
	}
}

func TestParse_functionDef(t *testing.T) {
	src := "fn multiply (a int b int) int\n    * a b"
	node := parseSingle(t, src)
	lst := node.(*ast.List)
	if lst.HeadValue() != "fn" {
		t.Errorf("expected fn head, got %q", lst.HeadValue())
	}
	s := lst.String()
	if !strings.Contains(s, "multiply") {
		t.Errorf("expected function name in %s", s)
	}
	if !strings.Contains(s, "(* a b)") {
		t.Errorf("expected body (* a b) in %s", s)
	}
}
