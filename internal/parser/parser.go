// Package parser converts a stream of lexer Lines (produced by the Glide
// lexer) into an S-expression AST.
//
// Parsing rules
// -------------
//  1. Each source line whose tokens are fully on that line forms a "head"
//     list starting with those tokens.
//  2. Any subsequent lines whose indentation is *strictly greater* than the
//     head line's indentation are "children" of that head.
//  3. Each child is recursively parsed and *appended* to the head list.
//
// Inline bracket groups are expanded eagerly: ( … ), [ … ], { … } are
// converted to ast.List nodes before the indentation processing runs.
//
// The ? try-suffix and the -> threading macro are recognised here as
// syntactic sugar and desugared into standard S-expressions.
package parser

import (
	"fmt"

	"github.com/elielamora/glide/internal/ast"
	"github.com/elielamora/glide/internal/lexer"
)

// Parse converts a slice of lexer Lines into a top-level list of AST nodes.
func Parse(lines []lexer.Line) ([]ast.Node, error) {
	nodes, _, err := parseBlock(lines, 0, -1)
	return nodes, err
}

// ParseString is a convenience helper that lexes and parses a Glide source string.
func ParseString(src string) ([]ast.Node, error) {
	lines, err := lexer.Lex(src)
	if err != nil {
		return nil, err
	}
	return Parse(lines)
}

// parseBlock parses a contiguous block of lines where every line's indent is
// strictly greater than parentIndent.  It returns the parsed nodes and the
// remaining (unconsumed) lines.
func parseBlock(lines []lexer.Line, pos int, parentIndent int) ([]ast.Node, int, error) {
	var nodes []ast.Node

	for pos < len(lines) {
		line := lines[pos]
		if line.Indent <= parentIndent {
			break // this line belongs to a higher-level block
		}

		baseIndent := line.Indent

		// Parse the tokens on this line into a flat node list (handling
		// inline ( ) [ ] { } groups).
		tokPos := 0
		lineNodes, err := parseTokens(line.Tokens, &tokPos, -1)
		if err != nil {
			return nil, pos, fmt.Errorf("line %d: %w", line.LineNum, err)
		}
		pos++

		// Recursively collect any children (lines indented deeper than this one).
		var children []ast.Node
		if pos < len(lines) && lines[pos].Indent > baseIndent {
			children, pos, err = parseBlock(lines, pos, baseIndent)
			if err != nil {
				return nil, pos, err
			}
		}

		// Combine head tokens + child expressions into one node.
		allNodes := append(lineNodes, children...)

		var node ast.Node
		switch len(allNodes) {
		case 0:
			continue
		case 1:
			node = allNodes[0]
		default:
			node = ast.NewList(allNodes...)
		}

		nodes = append(nodes, node)
	}

	return nodes, pos, nil
}

// parseTokens parses a flat token slice (one logical line's worth) into AST
// nodes, handling nested bracket groups recursively.
//
// endKind == -1 means "parse until end of token slice".
// endKind == TokRParen / TokRBracket / TokRBrace means "parse until that
// closing bracket is consumed".
func parseTokens(tokens []lexer.Token, pos *int, endKind int) ([]ast.Node, error) {
	var nodes []ast.Node

	for *pos < len(tokens) {
		tok := tokens[*pos]

		// Stop if we hit the expected closing bracket.
		if endKind >= 0 && int(tok.Kind) == endKind {
			*pos++ // consume the closing bracket
			return nodes, nil
		}

		switch tok.Kind {
		case lexer.TokAtom, lexer.TokString:
			// The ? suffix is syntactic sugar for the (try …) macro.
			// It wraps the immediately preceding expression.
			if tok.Kind == lexer.TokAtom && tok.Value == "?" {
				if len(nodes) == 0 {
					return nil, fmt.Errorf("? suffix with no preceding expression")
				}
				prev := nodes[len(nodes)-1]
				nodes[len(nodes)-1] = ast.NewList(ast.NewAtom("try"), prev)
				*pos++
				continue
			}
			nodes = append(nodes, ast.NewAtom(tok.Value))
			*pos++

		case lexer.TokLParen:
			*pos++
			children, err := parseTokens(tokens, pos, int(lexer.TokRParen))
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, ast.NewList(children...))

		case lexer.TokLBracket:
			*pos++
			children, err := parseTokens(tokens, pos, int(lexer.TokRBracket))
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, ast.NewBracketList(children...))

		case lexer.TokLBrace:
			*pos++
			children, err := parseTokens(tokens, pos, int(lexer.TokRBrace))
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, ast.NewBraceList(children...))

		case lexer.TokRParen:
			if endKind != int(lexer.TokRParen) {
				return nil, fmt.Errorf("unexpected ')'")
			}
			*pos++
			return nodes, nil

		case lexer.TokRBracket:
			if endKind != int(lexer.TokRBracket) {
				return nil, fmt.Errorf("unexpected ']'")
			}
			*pos++
			return nodes, nil

		case lexer.TokRBrace:
			if endKind != int(lexer.TokRBrace) {
				return nil, fmt.Errorf("unexpected '}'")
			}
			*pos++
			return nodes, nil

		default:
			return nil, fmt.Errorf("unknown token kind %d value %q", tok.Kind, tok.Value)
		}
	}

	// If we reach here with an expected end kind that was never seen, that is
	// an error only when parsing inline groups (the outer block parser never
	// sets an endKind).
	if endKind >= 0 {
		closing := map[int]string{
			int(lexer.TokRParen):   ")",
			int(lexer.TokRBracket): "]",
			int(lexer.TokRBrace):   "}",
		}
		return nil, fmt.Errorf("unclosed bracket, expected %q", closing[endKind])
	}

	return nodes, nil
}
