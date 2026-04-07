// Package lexer tokenises Glide source code into logical lines.
//
// Glide uses indentation to define S-expression nesting (the off-side rule).
// Each non-empty, non-comment line is returned as a Line containing its
// indentation depth (in spaces) and a flat list of Tokens.
//
// Inline bracket groups ( … ), [ … ], { … } are NOT folded here;
// that is the parser's job.  Commas are treated as whitespace.
package lexer

import (
	"fmt"
	"strings"
)

// TokenKind classifies a single lexer token.
type TokenKind int

const (
	TokAtom    TokenKind = iota // identifier, keyword, number, or operator
	TokString                   // "…" string literal (includes the quotes)
	TokLParen                   // (
	TokRParen                   // )
	TokLBracket                 // [
	TokRBracket                 // ]
	TokLBrace                   // {
	TokRBrace                   // }
)

// Token is a single lexical unit.
type Token struct {
	Kind  TokenKind
	Value string
	// Line and Col are 1-based source positions (for error messages).
	Line int
	Col  int
}

func (t Token) String() string { return t.Value }

// Line groups all tokens on one logical source line together with the
// number of leading spaces (the indentation level).
type Line struct {
	Indent  int
	Tokens  []Token
	LineNum int // 1-based line number of the first token
}

// Lex converts Glide source text into a slice of logical Lines.
// Empty lines and pure-comment lines are omitted.
// Tabs are expanded to 4 spaces.
func Lex(src string) ([]Line, error) {
	var lines []Line
	rawLines := strings.Split(src, "\n")

	for idx, raw := range rawLines {
		lineNum := idx + 1

		// Expand tabs.
		raw = strings.ReplaceAll(raw, "\t", "    ")

		// Measure indentation.
		indent := 0
		for indent < len(raw) && raw[indent] == ' ' {
			indent++
		}

		rest := raw[indent:]

		// Skip blank lines and full-line comments.
		if rest == "" || strings.HasPrefix(rest, ";") {
			continue
		}

		tokens, err := tokeniseLine(rest, lineNum)
		if err != nil {
			return nil, err
		}
		if len(tokens) > 0 {
			lines = append(lines, Line{
				Indent:  indent,
				Tokens:  tokens,
				LineNum: lineNum,
			})
		}
	}

	return lines, nil
}

// tokeniseLine scans a single line (already stripped of leading whitespace)
// and returns its tokens.  The lineNum argument is used for error messages.
func tokeniseLine(line string, lineNum int) ([]Token, error) {
	var tokens []Token
	i := 0

	for i < len(line) {
		ch := line[i]

		// Skip inline whitespace and commas (commas are separators for
		// readability only and carry no semantic meaning).
		if ch == ' ' || ch == '\t' || ch == ',' {
			i++
			continue
		}

		// Inline comment — skip the rest of the line.
		if ch == ';' {
			break
		}

		col := i + 1 // 1-based column

		// String literal.
		if ch == '"' {
			j := i + 1
			for j < len(line) {
				if line[j] == '\\' {
					j += 2 // skip escaped character
				} else if line[j] == '"' {
					j++
					break
				} else {
					j++
				}
			}
			tokens = append(tokens, Token{Kind: TokString, Value: line[i:j], Line: lineNum, Col: col})
			i = j
			continue
		}

		// Single-character bracket tokens.
		switch ch {
		case '(':
			tokens = append(tokens, Token{Kind: TokLParen, Value: "(", Line: lineNum, Col: col})
			i++
			continue
		case ')':
			tokens = append(tokens, Token{Kind: TokRParen, Value: ")", Line: lineNum, Col: col})
			i++
			continue
		case '[':
			tokens = append(tokens, Token{Kind: TokLBracket, Value: "[", Line: lineNum, Col: col})
			i++
			continue
		case ']':
			tokens = append(tokens, Token{Kind: TokRBracket, Value: "]", Line: lineNum, Col: col})
			i++
			continue
		case '{':
			tokens = append(tokens, Token{Kind: TokLBrace, Value: "{", Line: lineNum, Col: col})
			i++
			continue
		case '}':
			tokens = append(tokens, Token{Kind: TokRBrace, Value: "}", Line: lineNum, Col: col})
			i++
			continue
		}

		// Atom: any non-whitespace run that does not start a bracket or string.
		// We stop at whitespace, comma, comment, brackets, or quote.
		j := i
		for j < len(line) {
			c := line[j]
			if c == ' ' || c == '\t' || c == ',' || c == ';' ||
				c == '(' || c == ')' || c == '[' || c == ']' ||
				c == '{' || c == '}' || c == '"' {
				break
			}
			j++
		}

		if j > i {
			tokens = append(tokens, Token{Kind: TokAtom, Value: line[i:j], Line: lineNum, Col: col})
			i = j
			continue
		}

		return nil, fmt.Errorf("line %d col %d: unexpected character %q", lineNum, col, ch)
	}

	return tokens, nil
}
