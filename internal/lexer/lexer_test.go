package lexer

import (
	"testing"
)

func TestLex_empty(t *testing.T) {
	lines, err := Lex("")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 0 {
		t.Fatalf("expected 0 lines, got %d", len(lines))
	}
}

func TestLex_comment(t *testing.T) {
	lines, err := Lex("; this is a comment\n; another comment")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 0 {
		t.Fatalf("expected 0 lines after stripping comments, got %d", len(lines))
	}
}

func TestLex_simpleAtoms(t *testing.T) {
	lines, err := Lex("+ 1 2")
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	line := lines[0]
	if line.Indent != 0 {
		t.Errorf("expected indent 0, got %d", line.Indent)
	}
	want := []string{"+", "1", "2"}
	if len(line.Tokens) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(line.Tokens), line.Tokens)
	}
	for i, w := range want {
		if line.Tokens[i].Value != w {
			t.Errorf("token[%d]: want %q got %q", i, w, line.Tokens[i].Value)
		}
	}
}

func TestLex_indentation(t *testing.T) {
	src := "+ 1\n    * 2 3"
	lines, err := Lex(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0].Indent != 0 {
		t.Errorf("line 0: expected indent 0, got %d", lines[0].Indent)
	}
	if lines[1].Indent != 4 {
		t.Errorf("line 1: expected indent 4, got %d", lines[1].Indent)
	}
}

func TestLex_stringLiteral(t *testing.T) {
	lines, err := Lex(`let name "Alice"`)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 line")
	}
	toks := lines[0].Tokens
	if len(toks) != 3 {
		t.Fatalf("expected 3 tokens, got %d: %v", len(toks), toks)
	}
	if toks[2].Kind != TokString {
		t.Errorf("token[2] should be TokString, got kind %d", toks[2].Kind)
	}
	if toks[2].Value != `"Alice"` {
		t.Errorf("token[2] value: want %q got %q", `"Alice"`, toks[2].Value)
	}
}

func TestLex_brackets(t *testing.T) {
	lines, err := Lex(`let names ["Alice" "Bob"]`)
	if err != nil {
		t.Fatal(err)
	}
	toks := lines[0].Tokens
	// let names [ "Alice" "Bob" ]  — 6 tokens
	if len(toks) != 6 {
		t.Fatalf("expected 6 tokens, got %d: %v", len(toks), toks)
	}
	if toks[2].Kind != TokLBracket {
		t.Errorf("expected TokLBracket at index 2")
	}
	if toks[5].Kind != TokRBracket {
		t.Errorf("expected TokRBracket at index 5")
	}
}

func TestLex_inlineParen(t *testing.T) {
	lines, err := Lex("let result (+ 1 2)")
	if err != nil {
		t.Fatal(err)
	}
	toks := lines[0].Tokens
	// let result ( + 1 2 )
	if len(toks) != 7 {
		t.Fatalf("expected 7 tokens, got %d: %v", len(toks), toks)
	}
	if toks[2].Kind != TokLParen {
		t.Errorf("expected TokLParen at index 2")
	}
}

func TestLex_tabIndent(t *testing.T) {
	src := "fn foo\n\t+ 1 2"
	lines, err := Lex(src)
	if err != nil {
		t.Fatal(err)
	}
	if lines[1].Indent != 4 {
		t.Errorf("tab should expand to 4 spaces, got indent %d", lines[1].Indent)
	}
}

func TestLex_commaIgnored(t *testing.T) {
	lines, err := Lex("fn add (a int, b int) int")
	if err != nil {
		t.Fatal(err)
	}
	// tokens: fn add ( a int b int ) int   (comma stripped)
	toks := lines[0].Tokens
	if len(toks) != 9 {
		t.Fatalf("expected 9 tokens (comma stripped), got %d: %v", len(toks), toks)
	}
}

func TestLex_stringEscape(t *testing.T) {
	lines, err := Lex(`print "hello \"world\""`)
	if err != nil {
		t.Fatal(err)
	}
	toks := lines[0].Tokens
	if len(toks) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(toks))
	}
	if toks[1].Kind != TokString {
		t.Errorf("expected TokString")
	}
}

func TestLex_operators(t *testing.T) {
	lines, err := Lex(">= <= == != && ||")
	if err != nil {
		t.Fatal(err)
	}
	toks := lines[0].Tokens
	want := []string{">=", "<=", "==", "!=", "&&", "||"}
	if len(toks) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(toks), toks)
	}
	for i, w := range want {
		if toks[i].Value != w {
			t.Errorf("token[%d]: want %q got %q", i, w, toks[i].Value)
		}
	}
}
