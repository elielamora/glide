// Package transpiler converts a Glide S-expression AST into an executable
// Go source file.
//
// Supported forms
// ---------------
//  Binding    : (let name val), (let name type val), (const name val), (set name val)
//  Functions  : (fn name (params…) rettype body…), (fn (recv) .Method (params…) rettype body…)
//  Calls      : (funcName arg…), (.MethodName receiver arg…)
//  Operators  : +  -  *  /  %  >  <  >=  <=  ==  !=  &&  ||  !  &  <-
//  Control    : (if cond then [else]), (return expr)
//  Loops      : (for x in range body…), (each [idx val] in slice body…),
//               (loop (bindings…) body), (recur arg…)
//  Types      : (type Name struct field…)
//  Imports    : (use pkg.path [alias])
//  Testing    : (test "name" body…), (assert expr), (bench "name" body…)
//  Errors     : (try expr), (guard err body…)
//  Threading  : (-> val f1 f2 …), (->> val f1 f2 …)
//  Concurrency: (go expr), (chan type cap)
//  Misc       : (nil), (return expr), literals
package transpiler

import (
	"fmt"
	"go/format"
	"path"
	"strings"
	"unicode"

	"github.com/elielamora/glide/internal/ast"
	"github.com/elielamora/glide/internal/parser"
)

// wellKnownPackages lists Go standard-library package names that are safe to
// auto-import when a qualified identifier like fmt.Println is encountered.
// Single-letter variables (p, x, n, etc.) are excluded from auto-import.
var wellKnownPackages = map[string]bool{
	"fmt": true, "os": true, "io": true, "strings": true, "bytes": true,
	"errors": true, "strconv": true, "bufio": true, "sort": true, "sync": true,
	"math": true, "time": true, "rand": true, "regexp": true,
	"net": true, "http": true, "url": true, "context": true, "log": true,
	"filepath": true, "path": true, "unicode": true, "utf8": true,
	"runtime": true, "atomic": true, "reflect": true, "json": true,
	"encoding": true, "binary": true, "crypto": true, "hash": true,
	"big": true, "bits": true, "exec": true, "signal": true, "user": true,
	"flag": true, "testing": true, "plugin": true, "unsafe": true,
}
var binaryOps = map[string]string{
	"+":  "+",
	"-":  "-",
	"*":  "*",
	"/":  "/",
	"%":  "%",
	">":  ">",
	"<":  "<",
	">=": ">=",
	"<=": "<=",
	"==": "==",
	"!=": "!=",
	"&&": "&&",
	"||": "||",
}

// Transpiler holds per-file compilation state.
type Transpiler struct {
	pkgName string

	// imports: alias -> import-path (e.g. "fmt" -> "fmt", "web" -> "net/http")
	imports map[string]string
	// autoImports: packages discovered from qualified calls (fmt.Println → "fmt")
	autoImports map[string]bool

	// structTypes tracks names defined as struct types so that
	// (TypeName args…) can be rendered as TypeName{args…} rather than TypeName(args…).
	structTypes map[string]bool

	// top-level declarations (type defs, func defs, var/const at file scope)
	topLevel []string

	// main-function body lines
	mainBody []string

	// loop counter for generating unique variable names
	tmpCounter int

	// test functions accumulated for _test.go output
	testFuncs []string
}

// New creates a fresh Transpiler for a single source file.
func New() *Transpiler {
	return &Transpiler{
		pkgName:     "main",
		imports:     make(map[string]string),
		autoImports: make(map[string]bool),
		structTypes: make(map[string]bool),
	}
}

// tmpVar returns a unique temporary variable name.
func (tr *Transpiler) tmpVar() string {
	tr.tmpCounter++
	return fmt.Sprintf("_glide%d", tr.tmpCounter)
}

// indent returns a string of tabs appropriate for a given nesting depth.
func indent(depth int) string {
	return strings.Repeat("\t", depth)
}

// --------------------------------------------------------------------------
// Public API
// --------------------------------------------------------------------------

// TranspileFile converts a Glide source file into a Go source file.
// The returned string is valid, gofmt-formatted Go source.
func TranspileFile(src string) (string, error) {
	nodes, err := parser.ParseString(src)
	if err != nil {
		return "", fmt.Errorf("parse: %w", err)
	}
	tr := New()
	return tr.transpileNodes(nodes)
}

// transpileNodes processes the top-level AST nodes and assembles the output.
func (tr *Transpiler) transpileNodes(nodes []ast.Node) (string, error) {
	for _, node := range nodes {
		if err := tr.evalTopLevel(node); err != nil {
			return "", err
		}
	}
	return tr.render()
}

// evalTopLevel processes a single top-level AST node.
func (tr *Transpiler) evalTopLevel(node ast.Node) error {
	lst, ok := node.(*ast.List)
	if !ok {
		// bare atom at top level — treat as a main-body statement
		tr.mainBody = append(tr.mainBody, node.String())
		return nil
	}

	switch lst.HeadValue() {
	case "fn":
		out, err := tr.transpileFn(lst)
		if err != nil {
			return err
		}
		tr.topLevel = append(tr.topLevel, out)

	case "type":
		out, err := tr.transpileTypeDef(lst)
		if err != nil {
			return err
		}
		tr.topLevel = append(tr.topLevel, out)

	case "use":
		return tr.transpileUse(lst)

	case "const":
		out, err := tr.transpileConst(lst, 0)
		if err != nil {
			return err
		}
		tr.topLevel = append(tr.topLevel, out)

	case "test":
		out, err := tr.transpileTest(lst)
		if err != nil {
			return err
		}
		tr.testFuncs = append(tr.testFuncs, out)

	case "bench":
		out, err := tr.transpileBench(lst)
		if err != nil {
			return err
		}
		tr.testFuncs = append(tr.testFuncs, out)

	default:
		// Anything else becomes a statement inside main().
		lines, err := tr.transpileStmt(node, 1)
		if err != nil {
			return err
		}
		tr.mainBody = append(tr.mainBody, lines...)
	}
	return nil
}

// render assembles the final Go source file.
func (tr *Transpiler) render() (string, error) {
	var sb strings.Builder

	sb.WriteString("package ")
	sb.WriteString(tr.pkgName)
	sb.WriteString("\n\n")

	// Collect all imports.
	allImports := make(map[string]string) // alias -> path
	for alias, p := range tr.imports {
		allImports[alias] = p
	}
	for pkg := range tr.autoImports {
		if _, exists := allImports[pkg]; !exists {
			allImports[pkg] = pkg
		}
	}

	if len(allImports) > 0 {
		sb.WriteString("import (\n")
		for alias, p := range allImports {
			if alias == path.Base(p) {
				// standard import — no alias needed
				sb.WriteString(fmt.Sprintf("\t%q\n", p))
			} else {
				sb.WriteString(fmt.Sprintf("\t%s %q\n", alias, p))
			}
		}
		sb.WriteString(")\n\n")
	}

	for _, decl := range tr.topLevel {
		sb.WriteString(decl)
		sb.WriteString("\n\n")
	}

	// Only emit a main() if there are statements for it.
	if len(tr.mainBody) > 0 {
		sb.WriteString("func main() {\n")
		for _, line := range tr.mainBody {
			sb.WriteString(indent(1))
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		sb.WriteString("}\n")
	}

	src := sb.String()
	formatted, err := format.Source([]byte(src))
	if err != nil {
		// Return unformatted source so callers can still inspect it.
		return src, fmt.Errorf("go/format: %w\nsource:\n%s", err, src)
	}
	return string(formatted), nil
}

// --------------------------------------------------------------------------
// Statement transpilation
// --------------------------------------------------------------------------

// transpileStmt transpiles a single AST node as a Go statement.
// It returns one or more Go statement lines (without trailing newline).
// depth is the indentation depth.
func (tr *Transpiler) transpileStmt(node ast.Node, depth int) ([]string, error) {
	lst, ok := node.(*ast.List)
	if !ok {
		// Bare atom — treat as expression statement.
		return []string{node.String()}, nil
	}

	head := lst.HeadValue()

	switch head {
	case "let":
		return tr.transpileLet(lst, depth)
	case "const":
		line, err := tr.transpileConst(lst, depth)
		if err != nil {
			return nil, err
		}
		return []string{line}, nil
	case "set":
		return tr.transpileSet(lst, depth)
	case "fn":
		decl, err := tr.transpileFn(lst)
		if err != nil {
			return nil, err
		}
		// fn inside a body — assign as variable (closure).
		return []string{decl}, nil
	case "if":
		return tr.transpileIfStmt(lst, depth, false)
	case "for":
		return tr.transpileFor(lst, depth)
	case "each":
		return tr.transpileEach(lst, depth)
	case "loop":
		return tr.transpileLoop(lst, depth)
	case "return":
		if len(lst.Args()) == 0 {
			return []string{"return"}, nil
		}
		val, err := tr.transpileExpr(lst.Args()[0])
		if err != nil {
			return nil, err
		}
		return []string{"return " + val}, nil
	case "go":
		if len(lst.Args()) == 0 {
			return nil, fmt.Errorf("go: requires an expression argument")
		}
		expr, err := tr.transpileExpr(lst.Args()[0])
		if err != nil {
			return nil, err
		}
		return []string{"go " + expr}, nil
	case "try":
		return tr.transpileTryStmt(lst, depth)
	case "guard":
		return tr.transpileGuard(lst, depth)
	case "assert":
		return tr.transpileAssert(lst, depth)
	case "->", "->>":
		expr, err := tr.transpileThread(lst, head == "->>")
		if err != nil {
			return nil, err
		}
		return []string{expr}, nil
	case "close":
		if len(lst.Args()) == 0 {
			return nil, fmt.Errorf("close: requires channel argument")
		}
		ch, err := tr.transpileExpr(lst.Args()[0])
		if err != nil {
			return nil, err
		}
		return []string{fmt.Sprintf("close(%s)", ch)}, nil
	case "<-":
		// Channel send: (<- ch val)
		return tr.transpileChanSend(lst)
	case "recur":
		// recur is handled by the loop transpiler; here it's a no-op sentinel.
		return nil, fmt.Errorf("recur used outside loop/fn context")
	default:
		// Generic function/method call.
		expr, err := tr.transpileExpr(node)
		if err != nil {
			return nil, err
		}
		return []string{expr}, nil
	}
}

// transpileBlock transpiles a slice of AST nodes as a Go block body,
// optionally wrapping the last expression in a return statement.
func (tr *Transpiler) transpileBlock(nodes []ast.Node, depth int, autoReturn bool, retType string) ([]string, error) {
	var lines []string

	for i, node := range nodes {
		isLast := i == len(nodes)-1

		if isLast && autoReturn {
			// Check whether this node should be returned.
			if shouldReturn(node, retType) {
				stmts, err := tr.transpileReturnExpr(node, depth)
				if err != nil {
					return nil, err
				}
				lines = append(lines, stmts...)
				continue
			}
		}

		stmts, err := tr.transpileStmt(node, depth)
		if err != nil {
			return nil, err
		}
		lines = append(lines, stmts...)
	}

	return lines, nil
}

// shouldReturn reports whether the final node in a function body should be
// automatically wrapped with return.  retType is the declared return type.
func shouldReturn(node ast.Node, retType string) bool {
	if retType == "" || retType == "()" {
		return false
	}
	lst, ok := node.(*ast.List)
	if !ok {
		return true // bare atom is always an expression
	}
	switch lst.HeadValue() {
	case "let", "const", "set", "return", "go", "each", "for", "loop",
		"close", "guard", "assert", "type", "use":
		return false
	}
	return true
}

// transpileReturnExpr transpiles a node that should be used as a return value.
func (tr *Transpiler) transpileReturnExpr(node ast.Node, depth int) ([]string, error) {
	lst, ok := node.(*ast.List)
	if !ok {
		return []string{"return " + node.String()}, nil
	}
	switch lst.HeadValue() {
	case "if":
		// if in return position: expand into if/else with return on each branch
		return tr.transpileIfStmt(lst, depth, true)
	case "for":
		// for comprehension in return position
		return tr.transpileForReturn(lst, depth)
	case "loop":
		return tr.transpileLoop(lst, depth)
	default:
		expr, err := tr.transpileExpr(node)
		if err != nil {
			return nil, err
		}
		return []string{"return " + expr}, nil
	}
}

// --------------------------------------------------------------------------
// let / const / set
// --------------------------------------------------------------------------

func (tr *Transpiler) transpileLet(lst *ast.List, _ int) ([]string, error) {
	args := lst.Args()
	if len(args) < 2 {
		return nil, fmt.Errorf("let: need at least name and value, got %d args", len(args))
	}
	name := args[0].String()

	if len(args) == 2 {
		// (let name val)
		val, err := tr.transpileExpr(args[1])
		if err != nil {
			return nil, err
		}
		return []string{name + " := " + val}, nil
	}

	// (let name type val)
	typStr := tr.transpileType(args[1])
	val, err := tr.transpileExpr(args[2])
	if err != nil {
		return nil, err
	}
	return []string{fmt.Sprintf("var %s %s = %s", name, typStr, val)}, nil
}

func (tr *Transpiler) transpileConst(lst *ast.List, _ int) (string, error) {
	args := lst.Args()
	if len(args) < 2 {
		return "", fmt.Errorf("const: need name and value")
	}
	name := args[0].String()
	val, err := tr.transpileExpr(args[1])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("const %s = %s", name, val), nil
}

func (tr *Transpiler) transpileSet(lst *ast.List, _ int) ([]string, error) {
	args := lst.Args()
	if len(args) < 2 {
		return nil, fmt.Errorf("set: need name and value")
	}
	name := args[0].String()
	val, err := tr.transpileExpr(args[1])
	if err != nil {
		return nil, err
	}
	return []string{name + " = " + val}, nil
}

// --------------------------------------------------------------------------
// fn — function definition
// --------------------------------------------------------------------------

func (tr *Transpiler) transpileFn(lst *ast.List) (string, error) {
	args := lst.Args()
	if len(args) < 2 {
		return "", fmt.Errorf("fn: insufficient arguments")
	}

	// Detect method syntax: (fn (u *User) .MethodName (params…) rettype body…)
	// vs regular:          (fn name (params…) rettype body…)
	var receiver, funcName, retType string
	var paramList *ast.List
	var bodyNodes []ast.Node

	idx := 0

	// Optional receiver: first arg is a List → (recv *Type)
	if recvList, ok := args[idx].(*ast.List); ok {
		receiver = tr.transpileParams(recvList.Nodes)
		idx++
	}

	// Function name
	if idx >= len(args) {
		return "", fmt.Errorf("fn: missing name")
	}
	funcName = args[idx].String()
	// Strip leading dot for method names.
	funcName = strings.TrimPrefix(funcName, ".")
	// Only capitalise when the function name starts with a dot in the source
	// (explicit export marker).  The special Go entrypoints main and init
	// must stay lower-case.  All other names are kept exactly as written.
	idx++

	// Parameter list
	if idx >= len(args) {
		return "", fmt.Errorf("fn %s: missing parameter list", funcName)
	}
	if pl, ok := args[idx].(*ast.List); ok {
		paramList = pl
		idx++
	} else {
		paramList = ast.NewList()
	}

	// Return type (optional — anything that looks like a type before body nodes)
	if idx < len(args) {
		candidate := args[idx]
		if isTypeNode(candidate) {
			retType = tr.transpileType(candidate)
			idx++
		}
	}

	bodyNodes = args[idx:]

	// Build parameter string.
	params := tr.transpileParamPairs(paramList.Nodes)

	// Build function signature.
	var sig string
	if receiver != "" {
		sig = fmt.Sprintf("func (%s) %s(%s) %s", receiver, funcName, params, retType)
	} else {
		sig = fmt.Sprintf("func %s(%s) %s", funcName, params, retType)
	}
	sig = strings.TrimRight(sig, " ")

	// Transpile body.
	autoReturn := retType != "" && retType != "()"
	bodyLines, err := tr.transpileBlock(bodyNodes, 1, autoReturn, retType)
	if err != nil {
		return "", fmt.Errorf("fn %s body: %w", funcName, err)
	}

	var sb strings.Builder
	sb.WriteString(sig)
	sb.WriteString(" {\n")
	for _, l := range bodyLines {
		sb.WriteString(indent(1))
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	sb.WriteString("}")
	return sb.String(), nil
}

// transpileParamPairs converts a flat list of (name type name type …) tokens
// into a Go parameter list string: "a int, b int".
func (tr *Transpiler) transpileParamPairs(nodes []ast.Node) string {
	var pairs []string
	i := 0
	for i < len(nodes) {
		name := nodes[i].String()
		i++
		if i < len(nodes) {
			typ := tr.transpileType(nodes[i])
			i++
			pairs = append(pairs, name+" "+typ)
		}
	}
	return strings.Join(pairs, ", ")
}

// transpileParams converts a receiver or parameter list into a string.
func (tr *Transpiler) transpileParams(nodes []ast.Node) string {
	return tr.transpileParamPairs(nodes)
}

// --------------------------------------------------------------------------
// type — struct definition
// --------------------------------------------------------------------------

func (tr *Transpiler) transpileTypeDef(lst *ast.List) (string, error) {
	args := lst.Args()
	// (type Name struct field1 type1 field2 type2 …) — flat form
	// or (type Name struct (field1 type1) (field2 type2) …) — sub-list form
	if len(args) < 2 {
		return "", fmt.Errorf("type: need name and kind")
	}
	typeName := args[0].String()
	kind := args[1].String()
	if kind != "struct" {
		return "", fmt.Errorf("type: only 'struct' is supported, got %q", kind)
	}

	// Track this as a known struct type for constructor detection.
	tr.structTypes[typeName] = true

	fields := args[2:]
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("type %s struct {\n", typeName))

	// Detect form: each field may be a 2-element list (name type) — the
	// indented sub-expression form — or the fields may be flat atoms.
	if len(fields) > 0 {
		if _, ok := fields[0].(*ast.List); ok {
			// Sub-list form: each field is (name type)
			for _, f := range fields {
				if fl, ok := f.(*ast.List); ok && len(fl.Nodes) == 2 {
					fieldName := fl.Nodes[0].String()
					fieldType := tr.transpileType(fl.Nodes[1])
					sb.WriteString(fmt.Sprintf("\t%s %s\n", fieldName, fieldType))
				}
			}
		} else {
			// Flat form: fieldName type fieldName type …
			for i := 0; i+1 < len(fields); i += 2 {
				fieldName := fields[i].String()
				fieldType := tr.transpileType(fields[i+1])
				sb.WriteString(fmt.Sprintf("\t%s %s\n", fieldName, fieldType))
			}
		}
	}

	sb.WriteString("}")
	return sb.String(), nil
}

// --------------------------------------------------------------------------
// use — import
// --------------------------------------------------------------------------

func (tr *Transpiler) transpileUse(lst *ast.List) error {
	args := lst.Args()
	if len(args) == 0 {
		return fmt.Errorf("use: requires a package path")
	}
	pkgPath := strings.ReplaceAll(args[0].String(), ".", "/")
	alias := path.Base(pkgPath)
	if len(args) >= 2 {
		alias = args[1].String()
	}
	tr.imports[alias] = pkgPath
	return nil
}

// --------------------------------------------------------------------------
// if — conditional
// --------------------------------------------------------------------------

func (tr *Transpiler) transpileIfStmt(lst *ast.List, depth int, returnBranches bool) ([]string, error) {
	args := lst.Args()
	if len(args) < 2 {
		return nil, fmt.Errorf("if: need condition and then-branch")
	}

	cond, err := tr.transpileExpr(args[0])
	if err != nil {
		return nil, err
	}

	thenStmts, err := tr.transpileBranch(args[1], depth+1, returnBranches)
	if err != nil {
		return nil, err
	}

	var lines []string
	lines = append(lines, "if "+cond+" {")
	for _, s := range thenStmts {
		lines = append(lines, indent(1)+s)
	}

	if len(args) >= 3 {
		elseStmts, err := tr.transpileBranch(args[2], depth+1, returnBranches)
		if err != nil {
			return nil, err
		}
		lines = append(lines, "} else {")
		for _, s := range elseStmts {
			lines = append(lines, indent(1)+s)
		}
	}
	lines = append(lines, "}")
	return lines, nil
}

// transpileBranch transpiles a single branch of an if expression.
func (tr *Transpiler) transpileBranch(node ast.Node, depth int, withReturn bool) ([]string, error) {
	if withReturn {
		return tr.transpileReturnExpr(node, depth)
	}
	return tr.transpileStmt(node, depth)
}

// --------------------------------------------------------------------------
// for — comprehension loop
// --------------------------------------------------------------------------

// transpileFor handles (for x in start..end body…) and (for x in slice body…).
func (tr *Transpiler) transpileFor(lst *ast.List, depth int) ([]string, error) {
	args := lst.Args()
	// Expected: varName "in" rangeExpr body…
	if len(args) < 3 {
		return nil, fmt.Errorf("for: need var 'in' range body")
	}
	varName := args[0].String()
	if args[1].String() != "in" {
		return nil, fmt.Errorf("for: expected 'in', got %q", args[1].String())
	}
	rangeExpr := args[2]
	bodyNodes := args[3:]

	rangeStr := rangeExpr.String()

	var header string
	if strings.Contains(rangeStr, "..=") {
		// Inclusive range: 1..=5  →  x := 1; x <= 5; x++
		parts := strings.SplitN(rangeStr, "..=", 2)
		header = fmt.Sprintf("for %s := %s; %s <= %s; %s++", varName, parts[0], varName, parts[1], varName)
	} else if strings.Contains(rangeStr, "..") {
		// Exclusive range: 1..5  →  x := 1; x < 5; x++
		parts := strings.SplitN(rangeStr, "..", 2)
		header = fmt.Sprintf("for %s := %s; %s < %s; %s++", varName, parts[0], varName, parts[1], varName)
	} else {
		// Slice range: for x in slice
		header = fmt.Sprintf("for _, %s := range %s", varName, rangeStr)
	}

	return tr.buildForBlock(header, bodyNodes, depth)
}

// transpileForReturn wraps a for comprehension in a block that accumulates
// results and returns the slice.
func (tr *Transpiler) transpileForReturn(lst *ast.List, depth int) ([]string, error) {
	args := lst.Args()
	if len(args) < 3 {
		return nil, fmt.Errorf("for (return): insufficient args")
	}
	varName := args[0].String()
	if args[1].String() != "in" {
		return nil, fmt.Errorf("for: expected 'in'")
	}
	rangeExpr := args[2]
	bodyNodes := args[3:]

	rangeStr := rangeExpr.String()
	resultVar := tr.tmpVar()

	var header string
	if strings.Contains(rangeStr, "..=") {
		parts := strings.SplitN(rangeStr, "..=", 2)
		header = fmt.Sprintf("for %s := %s; %s <= %s; %s++", varName, parts[0], varName, parts[1], varName)
	} else if strings.Contains(rangeStr, "..") {
		parts := strings.SplitN(rangeStr, "..", 2)
		header = fmt.Sprintf("for %s := %s; %s < %s; %s++", varName, parts[0], varName, parts[1], varName)
	} else {
		header = fmt.Sprintf("for _, %s := range %s", varName, rangeStr)
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("var %s []interface{}", resultVar))
	lines = append(lines, header+" {")
	for _, bNode := range bodyNodes {
		expr, err := tr.transpileExpr(bNode)
		if err != nil {
			return nil, err
		}
		lines = append(lines, indent(1)+fmt.Sprintf("%s = append(%s, %s)", resultVar, resultVar, expr))
	}
	lines = append(lines, "}")
	lines = append(lines, "return "+resultVar)
	return lines, nil
}

func (tr *Transpiler) buildForBlock(header string, bodyNodes []ast.Node, depth int) ([]string, error) {
	var lines []string
	lines = append(lines, header+" {")
	for _, bNode := range bodyNodes {
		stmts, err := tr.transpileStmt(bNode, depth+1)
		if err != nil {
			return nil, err
		}
		for _, s := range stmts {
			lines = append(lines, indent(1)+s)
		}
	}
	lines = append(lines, "}")
	return lines, nil
}

// --------------------------------------------------------------------------
// each — range loop with index
// --------------------------------------------------------------------------

func (tr *Transpiler) transpileEach(lst *ast.List, depth int) ([]string, error) {
	args := lst.Args()
	// (each [idx val] in slice body…) or (each val in slice body…)
	if len(args) < 3 {
		return nil, fmt.Errorf("each: need bindings 'in' slice body")
	}

	var idxVar, valVar string
	inKeyword := ""
	bodyStart := 3

	if bracket, ok := args[0].(*ast.List); ok && bracket.Kind == ast.ListBracket {
		// [idx val]
		if len(bracket.Nodes) >= 1 {
			idxVar = bracket.Nodes[0].String()
		}
		if len(bracket.Nodes) >= 2 {
			valVar = bracket.Nodes[1].String()
		}
		inKeyword = args[1].String()
	} else {
		valVar = args[0].String()
		inKeyword = args[1].String()
	}
	_ = inKeyword

	sliceExpr, err := tr.transpileExpr(args[2])
	if err != nil {
		return nil, err
	}

	var rangeVars string
	if idxVar == "" {
		rangeVars = "_, " + valVar
	} else if valVar == "" {
		rangeVars = idxVar + ", _"
	} else {
		rangeVars = idxVar + ", " + valVar
	}

	header := fmt.Sprintf("for %s := range %s", rangeVars, sliceExpr)
	bodyNodes := args[bodyStart:]

	// Check if iterating over a channel (no range vars for channel receive).
	// If args[0] is not a bracket and there's no second binding, assume channel.
	if _, ok := args[0].(*ast.List); !ok {
		header = fmt.Sprintf("for %s := range %s", valVar, sliceExpr)
	}

	return tr.buildForBlock(header, bodyNodes, depth)
}

// --------------------------------------------------------------------------
// loop / recur — tail-recursive loop
// --------------------------------------------------------------------------

func (tr *Transpiler) transpileLoop(lst *ast.List, depth int) ([]string, error) {
	args := lst.Args()
	if len(args) < 2 {
		return nil, fmt.Errorf("loop: need bindings and body")
	}

	// First arg: (var1 init1 var2 init2 …) or single binding node
	bindingNode := args[0]
	bodyNodes := args[1:]

	var initLines []string
	var loopVars []string

	if bindList, ok := bindingNode.(*ast.List); ok {
		nodes := bindList.Nodes
		for i := 0; i+1 < len(nodes); i += 2 {
			varName := nodes[i].String()
			initVal, err := tr.transpileExpr(nodes[i+1])
			if err != nil {
				return nil, err
			}
			initLines = append(initLines, fmt.Sprintf("%s := %s", varName, initVal))
			loopVars = append(loopVars, varName)
		}
	}

	// Transpile the body, replacing recur with assignments + continue.
	bodyLines, err := tr.transpileLoopBody(bodyNodes, loopVars, 1)
	if err != nil {
		return nil, err
	}

	var lines []string
	lines = append(lines, initLines...)
	lines = append(lines, "for {")
	for _, l := range bodyLines {
		lines = append(lines, indent(1)+l)
	}
	lines = append(lines, "}")
	return lines, nil
}

// transpileLoopBody transpiles the body of a loop, expanding recur into
// assignment + continue.
func (tr *Transpiler) transpileLoopBody(nodes []ast.Node, loopVars []string, depth int) ([]string, error) {
	var lines []string
	for i, node := range nodes {
		isLast := i == len(nodes)-1

		// Direct recur — expand unconditionally.
		if lst, ok := node.(*ast.List); ok && lst.HeadValue() == "recur" {
			recurLines, err := tr.transpileRecur(lst, loopVars)
			if err != nil {
				return nil, err
			}
			lines = append(lines, recurLines...)
			continue
		}

		// if expression whose branches may contain recur.
		if lst, ok := node.(*ast.List); ok && lst.HeadValue() == "if" {
			if branchContainsRecur(lst) {
				ifLines, err := tr.transpileLoopIf(lst, loopVars, depth)
				if err != nil {
					return nil, err
				}
				lines = append(lines, ifLines...)
				continue
			}
		}

		if isLast {
			retLines, err := tr.transpileReturnExpr(node, depth)
			if err != nil {
				return nil, err
			}
			lines = append(lines, retLines...)
		} else {
			stmts, err := tr.transpileStmt(node, depth)
			if err != nil {
				return nil, err
			}
			lines = append(lines, stmts...)
		}
	}
	return lines, nil
}

// branchContainsRecur reports whether any branch argument of an if list
// is a direct recur form.
func branchContainsRecur(ifLst *ast.List) bool {
	args := ifLst.Args()
	for _, arg := range args[1:] { // skip condition
		if isRecurNode(arg) {
			return true
		}
	}
	return false
}

// isRecurNode reports whether node is a (recur …) list.
func isRecurNode(node ast.Node) bool {
	if lst, ok := node.(*ast.List); ok {
		return lst.HeadValue() == "recur"
	}
	return false
}

// transpileLoopIf handles an (if cond then else) node where one or both
// branches contain a recur form.  The recur branch is expanded into
// variable assignments; the non-recur branch becomes a return statement.
func (tr *Transpiler) transpileLoopIf(lst *ast.List, loopVars []string, depth int) ([]string, error) {
	args := lst.Args()
	if len(args) < 2 {
		return nil, fmt.Errorf("if (loop): too few arguments")
	}

	cond, err := tr.transpileExpr(args[0])
	if err != nil {
		return nil, err
	}

	// Two-branch if (with else).
	if len(args) >= 3 {
		thenNode := args[1]
		elseNode := args[2]

		thenIsRecur := isRecurNode(thenNode)
		elseIsRecur := isRecurNode(elseNode)

		switch {
		case !thenIsRecur && elseIsRecur:
			// if cond { return thenVal }   <then recur assignments>
			thenLines, err := tr.transpileReturnExpr(thenNode, depth)
			if err != nil {
				return nil, err
			}
			recurLines, err := tr.transpileRecur(elseNode.(*ast.List), loopVars)
			if err != nil {
				return nil, err
			}
			var lines []string
			lines = append(lines, "if "+cond+" {")
			for _, l := range thenLines {
				lines = append(lines, indent(1)+l)
			}
			lines = append(lines, "}")
			lines = append(lines, recurLines...)
			return lines, nil

		case thenIsRecur && !elseIsRecur:
			// if !cond { return elseVal }   <then recur assignments>
			elseLines, err := tr.transpileReturnExpr(elseNode, depth)
			if err != nil {
				return nil, err
			}
			recurLines, err := tr.transpileRecur(thenNode.(*ast.List), loopVars)
			if err != nil {
				return nil, err
			}
			var lines []string
			lines = append(lines, fmt.Sprintf("if !(%s) {", cond))
			for _, l := range elseLines {
				lines = append(lines, indent(1)+l)
			}
			lines = append(lines, "}")
			lines = append(lines, recurLines...)
			return lines, nil

		case thenIsRecur && elseIsRecur:
			// Both recur — generate conditional recur expansion.
			thenLines, err := tr.transpileRecur(thenNode.(*ast.List), loopVars)
			if err != nil {
				return nil, err
			}
			elseLines, err := tr.transpileRecur(elseNode.(*ast.List), loopVars)
			if err != nil {
				return nil, err
			}
			var lines []string
			lines = append(lines, "if "+cond+" {")
			for _, l := range thenLines {
				lines = append(lines, indent(1)+l)
			}
			lines = append(lines, "} else {")
			for _, l := range elseLines {
				lines = append(lines, indent(1)+l)
			}
			lines = append(lines, "}")
			return lines, nil
		}
	}

	// One-branch if or no-recur case — fall back to regular if with return.
	return tr.transpileIfStmt(lst, depth, true)
}

// transpileRecur expands (recur arg1 arg2 …) into temp assignments + continue.
func (tr *Transpiler) transpileRecur(lst *ast.List, loopVars []string) ([]string, error) {
	args := lst.Args()
	if len(args) != len(loopVars) {
		return nil, fmt.Errorf("recur: expected %d args, got %d", len(loopVars), len(args))
	}

	// Use temps to avoid clobbering values before all new values are computed.
	var tmpNames []string
	var lines []string
	for i, arg := range args {
		tmp := tr.tmpVar()
		tmpNames = append(tmpNames, tmp)
		val, err := tr.transpileExpr(arg)
		if err != nil {
			return nil, err
		}
		lines = append(lines, fmt.Sprintf("%s := %s", tmp, val))
		_ = i
	}
	for i, orig := range loopVars {
		lines = append(lines, fmt.Sprintf("%s = %s", orig, tmpNames[i]))
	}
	lines = append(lines, "continue")
	return lines, nil
}

// --------------------------------------------------------------------------
// try / guard — error handling
// --------------------------------------------------------------------------

func (tr *Transpiler) transpileTryStmt(lst *ast.List, _ int) ([]string, error) {
	args := lst.Args()
	if len(args) == 0 {
		return nil, fmt.Errorf("try: requires an expression")
	}
	expr, err := tr.transpileExpr(args[0])
	if err != nil {
		return nil, err
	}
	tmp := tr.tmpVar()
	return []string{
		fmt.Sprintf("%s, _err := %s", tmp, expr),
		"if _err != nil { return _err }",
		"_ = " + tmp,
	}, nil
}

func (tr *Transpiler) transpileGuard(lst *ast.List, depth int) ([]string, error) {
	args := lst.Args()
	if len(args) < 2 {
		return nil, fmt.Errorf("guard: need error var and handler body")
	}
	errVar := args[0].String()
	bodyNodes := args[1:]

	bodyLines, err := tr.transpileBlock(bodyNodes, depth+1, false, "")
	if err != nil {
		return nil, err
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("if %s != nil {", errVar))
	for _, l := range bodyLines {
		lines = append(lines, indent(1)+l)
	}
	lines = append(lines, "}")
	return lines, nil
}

// --------------------------------------------------------------------------
// assert
// --------------------------------------------------------------------------

func (tr *Transpiler) transpileAssert(lst *ast.List, _ int) ([]string, error) {
	args := lst.Args()
	if len(args) == 0 {
		return nil, fmt.Errorf("assert: requires an expression")
	}
	expr, err := tr.transpileExpr(args[0])
	if err != nil {
		return nil, err
	}
	return []string{
		fmt.Sprintf(`if !(%s) { t.Fatalf("assertion failed: %%v", %q) }`, expr, "assert "+args[0].String()),
	}, nil
}

// --------------------------------------------------------------------------
// test / bench — test generation
// --------------------------------------------------------------------------

func (tr *Transpiler) transpileTest(lst *ast.List) (string, error) {
	args := lst.Args()
	if len(args) < 2 {
		return "", fmt.Errorf("test: need name and body")
	}
	rawName := strings.Trim(args[0].String(), `"`)
	funcName := "Test" + toGoIdentifier(rawName)
	bodyNodes := args[1:]

	bodyLines, err := tr.transpileBlock(bodyNodes, 1, false, "")
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", funcName))
	for _, l := range bodyLines {
		sb.WriteString(indent(1))
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	sb.WriteString("}")
	return sb.String(), nil
}

func (tr *Transpiler) transpileBench(lst *ast.List) (string, error) {
	args := lst.Args()
	if len(args) < 2 {
		return "", fmt.Errorf("bench: need name and body")
	}
	rawName := strings.Trim(args[0].String(), `"`)
	funcName := "Benchmark" + toGoIdentifier(rawName)
	bodyNodes := args[1:]

	bodyLines, err := tr.transpileBlock(bodyNodes, 1, false, "")
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("func %s(b *testing.B) {\n", funcName))
	sb.WriteString("\tfor i := 0; i < b.N; i++ {\n")
	for _, l := range bodyLines {
		sb.WriteString(indent(2))
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	sb.WriteString("\t}\n")
	sb.WriteString("}")
	return sb.String(), nil
}

// --------------------------------------------------------------------------
// threading macros -> and ->>
// --------------------------------------------------------------------------

// transpileThread handles (-> val f1 f2 …) and (->> val f1 f2 …).
// Thread-first (->)  pipes the value as the FIRST argument of each step.
// Thread-last  (->>) pipes the value as the LAST  argument of each step.
func (tr *Transpiler) transpileThread(lst *ast.List, last bool) (string, error) {
	args := lst.Args()
	if len(args) < 1 {
		return "", fmt.Errorf("->: requires at least a value")
	}

	current, err := tr.transpileExpr(args[0])
	if err != nil {
		return "", err
	}

	for _, step := range args[1:] {
		switch s := step.(type) {
		case *ast.Atom:
			// bare identifier — call as function with value as sole argument
			current = fmt.Sprintf("%s(%s)", s.Value, current)
		case *ast.List:
			// existing call with additional args — splice value in
			if len(s.Nodes) == 0 {
				current = fmt.Sprintf("(%s)(%s)", s.String(), current)
				continue
			}
			fn := s.Nodes[0].String()
			var argExprs []string
			for _, a := range s.Nodes[1:] {
				ae, err := tr.transpileExpr(a)
				if err != nil {
					return "", err
				}
				argExprs = append(argExprs, ae)
			}
			if last {
				argExprs = append(argExprs, current)
			} else {
				argExprs = append([]string{current}, argExprs...)
			}
			current = fmt.Sprintf("%s(%s)", fn, strings.Join(argExprs, ", "))
		default:
			current = fmt.Sprintf("%s(%s)", step.String(), current)
		}
	}
	return current, nil
}

// --------------------------------------------------------------------------
// channel send
// --------------------------------------------------------------------------

func (tr *Transpiler) transpileChanSend(lst *ast.List) ([]string, error) {
	args := lst.Args()
	if len(args) < 2 {
		return nil, fmt.Errorf("<-: need channel and value")
	}
	ch, err := tr.transpileExpr(args[0])
	if err != nil {
		return nil, err
	}
	val, err := tr.transpileExpr(args[1])
	if err != nil {
		return nil, err
	}
	return []string{fmt.Sprintf("%s <- %s", ch, val)}, nil
}

// --------------------------------------------------------------------------
// Expression transpilation
// --------------------------------------------------------------------------

// transpileExpr converts an AST node to a Go expression string.
func (tr *Transpiler) transpileExpr(node ast.Node) (string, error) {
	switch n := node.(type) {
	case *ast.Atom:
		return tr.transpileAtomExpr(n.Value), nil

	case *ast.List:
		switch n.Kind {
		case ast.ListBracket:
			return tr.transpileSliceLit(n)
		case ast.ListBrace:
			return tr.transpileMapLit(n)
		case ast.ListParen:
			return tr.transpileListExpr(n)
		}
	}
	return "", fmt.Errorf("transpileExpr: unknown node type %T", node)
}

func (tr *Transpiler) transpileAtomExpr(v string) string {
	// Auto-detect well-known package prefixes for imports.
	if idx := strings.Index(v, "."); idx > 0 {
		pkg := v[:idx]
		if wellKnownPackages[pkg] {
			tr.autoImports[pkg] = true
		}
	}
	return v
}

func (tr *Transpiler) transpileListExpr(lst *ast.List) (string, error) {
	if len(lst.Nodes) == 0 {
		return "()", nil
	}

	head := lst.HeadValue()

	// Binary operator?
	if goOp, ok := binaryOps[head]; ok {
		return tr.transpileBinaryOp(goOp, lst.Args())
	}

	// Unary operators.
	switch head {
	case "!":
		if len(lst.Args()) != 1 {
			return "", fmt.Errorf("!: expected 1 argument")
		}
		operand, err := tr.transpileExpr(lst.Args()[0])
		if err != nil {
			return "", err
		}
		return "!" + operand, nil

	case "not":
		if len(lst.Args()) != 1 {
			return "", fmt.Errorf("not: expected 1 argument")
		}
		operand, err := tr.transpileExpr(lst.Args()[0])
		if err != nil {
			return "", err
		}
		return "!" + operand, nil

	case "&":
		if len(lst.Args()) != 1 {
			return "", fmt.Errorf("&: expected 1 argument")
		}
		operand, err := tr.transpileExpr(lst.Args()[0])
		if err != nil {
			return "", err
		}
		return "&" + operand, nil

	case "<-":
		// Channel receive: (<- ch)
		if len(lst.Args()) != 1 {
			return "", fmt.Errorf("<-: receive needs 1 argument")
		}
		ch, err := tr.transpileExpr(lst.Args()[0])
		if err != nil {
			return "", err
		}
		return "<-" + ch, nil

	case "if":
		return tr.transpileIfExpr(lst)

	case "set":
		// Inline set — returns the assigned value.
		args := lst.Args()
		if len(args) < 2 {
			return "", fmt.Errorf("set (expr): need name and value")
		}
		name := args[0].String()
		val, err := tr.transpileExpr(args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() interface{} { %s = %s; return %s }()", name, val, name), nil

	case "try":
		args := lst.Args()
		if len(args) == 0 {
			return "", fmt.Errorf("try: requires an expression")
		}
		inner, err := tr.transpileExpr(args[0])
		if err != nil {
			return "", err
		}
		// Return the first result value; caller is responsible for error handling.
		tmp := tr.tmpVar()
		return fmt.Sprintf("func() interface{} { %s, _err := %s; if _err != nil { return _err }; return %s }()", tmp, inner, tmp), nil

	case "->", "->>":
		return tr.transpileThread(lst, head == "->>")

	case "chan":
		// (chan type cap) → make(chan type, cap)
		args := lst.Args()
		if len(args) < 1 {
			return "", fmt.Errorf("chan: need type")
		}
		typ := tr.transpileType(args[0])
		if len(args) >= 2 {
			cap, err := tr.transpileExpr(args[1])
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("make(chan %s, %s)", typ, cap), nil
		}
		return fmt.Sprintf("make(chan %s)", typ), nil

	case "fn":
		// Inline anonymous function.
		return tr.transpileAnonFn(lst)
	}

	// Method call: head starts with "."
	if strings.HasPrefix(head, ".") {
		return tr.transpileMethodCall(lst)
	}

	// Regular function call (possibly qualified: fmt.Println, etc.)
	return tr.transpileFuncCall(lst)
}

// transpileBinaryOp generates an infix expression for a binary operator.
func (tr *Transpiler) transpileBinaryOp(op string, args []ast.Node) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("binary op %q: need 2 arguments, got %d", op, len(args))
	}
	left, err := tr.transpileExpr(args[0])
	if err != nil {
		return "", err
	}
	right, err := tr.transpileExpr(args[1])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("(%s %s %s)", left, op, right), nil
}

// transpileIfExpr generates an IIFE for an if expression.
func (tr *Transpiler) transpileIfExpr(lst *ast.List) (string, error) {
	args := lst.Args()
	if len(args) < 2 {
		return "", fmt.Errorf("if (expr): need condition and then-branch")
	}
	cond, err := tr.transpileExpr(args[0])
	if err != nil {
		return "", err
	}
	thenExpr, err := tr.transpileExpr(args[1])
	if err != nil {
		return "", err
	}

	if len(args) >= 3 {
		elseExpr, err := tr.transpileExpr(args[2])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() interface{} { if %s { return %s }; return %s }()", cond, thenExpr, elseExpr), nil
	}
	return fmt.Sprintf("func() interface{} { if %s { return %s }; return nil }()", cond, thenExpr), nil
}

// transpileFuncCall generates a standard Go function call or struct literal.
func (tr *Transpiler) transpileFuncCall(lst *ast.List) (string, error) {
	fn := lst.Head().String()

	// Track qualified imports (well-known packages only).
	if idx := strings.Index(fn, "."); idx > 0 {
		pkg := fn[:idx]
		if wellKnownPackages[pkg] {
			tr.autoImports[pkg] = true
		}
	}

	var argStrs []string
	for _, a := range lst.Args() {
		s, err := tr.transpileExpr(a)
		if err != nil {
			return "", err
		}
		argStrs = append(argStrs, s)
	}

	// If the callee is a known struct type, generate a struct literal: Type{args}.
	if tr.structTypes[fn] {
		return fmt.Sprintf("%s{%s}", fn, strings.Join(argStrs, ", ")), nil
	}

	return fmt.Sprintf("%s(%s)", fn, strings.Join(argStrs, ", ")), nil
}

// transpileMethodCall generates a Go method call: (.MethodName receiver args…)
func (tr *Transpiler) transpileMethodCall(lst *ast.List) (string, error) {
	method := strings.TrimPrefix(lst.HeadValue(), ".")
	args := lst.Args()
	if len(args) == 0 {
		return "", fmt.Errorf(".%s: need at least a receiver argument", method)
	}
	recv, err := tr.transpileExpr(args[0])
	if err != nil {
		return "", err
	}
	var argStrs []string
	for _, a := range args[1:] {
		s, err := tr.transpileExpr(a)
		if err != nil {
			return "", err
		}
		argStrs = append(argStrs, s)
	}
	return fmt.Sprintf("%s.%s(%s)", recv, method, strings.Join(argStrs, ", ")), nil
}

// transpileAnonFn generates an anonymous function expression.
func (tr *Transpiler) transpileAnonFn(lst *ast.List) (string, error) {
	args := lst.Args()
	// (fn () body…) or (fn (params…) rettype body…)
	var paramList *ast.List
	var retType string
	idx := 0

	// Skip optional name — anonymous fns triggered via transpileListExpr always
	// have fn as the head, so args[0] is either params or name.
	if idx < len(args) {
		if pl, ok := args[idx].(*ast.List); ok {
			paramList = pl
			idx++
		} else {
			// First arg is a name atom — skip it for anon fn.
			idx++
			if idx < len(args) {
				if pl, ok := args[idx].(*ast.List); ok {
					paramList = pl
					idx++
				}
			}
		}
	}
	if paramList == nil {
		paramList = ast.NewList()
	}

	if idx < len(args) && isTypeNode(args[idx]) {
		retType = tr.transpileType(args[idx])
		idx++
	}

	bodyNodes := args[idx:]
	params := tr.transpileParamPairs(paramList.Nodes)

	autoReturn := retType != "" && retType != "()"
	bodyLines, err := tr.transpileBlock(bodyNodes, 2, autoReturn, retType)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("func(")
	sb.WriteString(params)
	sb.WriteString(")")
	if retType != "" {
		sb.WriteString(" ")
		sb.WriteString(retType)
	}
	sb.WriteString(" {\n")
	for _, l := range bodyLines {
		sb.WriteString(indent(2))
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	sb.WriteString(indent(1))
	sb.WriteString("}")
	return sb.String(), nil
}

// --------------------------------------------------------------------------
// Slice and map literals
// --------------------------------------------------------------------------

func (tr *Transpiler) transpileSliceLit(lst *ast.List) (string, error) {
	var elems []string
	for _, n := range lst.Nodes {
		s, err := tr.transpileExpr(n)
		if err != nil {
			return "", err
		}
		elems = append(elems, s)
	}
	return "[]interface{}{" + strings.Join(elems, ", ") + "}", nil
}

func (tr *Transpiler) transpileMapLit(lst *ast.List) (string, error) {
	nodes := lst.Nodes
	var pairs []string
	for i := 0; i+1 < len(nodes); i += 2 {
		k, err := tr.transpileExpr(nodes[i])
		if err != nil {
			return "", err
		}
		v, err := tr.transpileExpr(nodes[i+1])
		if err != nil {
			return "", err
		}
		pairs = append(pairs, k+": "+v)
	}
	return "map[string]interface{}{" + strings.Join(pairs, ", ") + "}", nil
}

// --------------------------------------------------------------------------
// Type transpilation
// --------------------------------------------------------------------------

// transpileType converts a Glide type node to a Go type string.
func (tr *Transpiler) transpileType(node ast.Node) string {
	switch n := node.(type) {
	case *ast.Atom:
		return glideTypeToGo(n.Value)

	case *ast.List:
		switch n.Kind {
		case ast.ListBracket:
			// [T] → []T
			if len(n.Nodes) == 1 {
				return "[]" + tr.transpileType(n.Nodes[0])
			}
			if len(n.Nodes) == 0 {
				return "[]interface{}"
			}
		case ast.ListParen:
			// (T1 T2) → multiple return: (T1, T2)
			var parts []string
			for _, child := range n.Nodes {
				parts = append(parts, tr.transpileType(child))
			}
			return "(" + strings.Join(parts, ", ") + ")"
		}
		return n.String()
	}
	return node.String()
}

// glideTypeToGo maps Glide primitive type names to Go equivalents.
func glideTypeToGo(t string) string {
	switch t {
	case "int", "int64", "int32", "int16", "int8",
		"uint", "uint64", "uint32", "uint16", "uint8",
		"float64", "float32",
		"bool", "string", "byte", "rune", "error",
		"any", "interface{}":
		return t
	case "void":
		return ""
	}
	// Pointer or slice types passed as atoms: *User, []int, map[string]int
	if strings.HasPrefix(t, "*") || strings.HasPrefix(t, "[]") || strings.HasPrefix(t, "map[") {
		return t
	}
	// Assume it's a user-defined type name.
	return t
}

// isTypeNode reports whether a node looks like a type annotation.
func isTypeNode(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.Atom:
		t := n.Value
		switch t {
		case "int", "int64", "int32", "int16", "int8",
			"uint", "uint64", "uint32", "uint16", "uint8",
			"float64", "float32",
			"bool", "string", "byte", "rune", "error",
			"any", "interface{}", "void":
			return true
		}
		// Pointer or slice type: *T, []T, map[k]v
		if strings.HasPrefix(t, "*") || strings.HasPrefix(t, "[]") || strings.HasPrefix(t, "map[") {
			return true
		}
		// Multi-return tuple: (T1 T2) represented as paren list would be caught below
		return false
	case *ast.List:
		if n.Kind == ast.ListBracket {
			return true // [T] is a slice type
		}
		if n.Kind == ast.ListParen && len(n.Nodes) >= 1 {
			// (T1 T2) could be a multi-return type
			for _, child := range n.Nodes {
				if !isTypeNode(child) {
					return false
				}
			}
			return true
		}
	}
	return false
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

// toGoIdentifier converts a human-readable test name like "math operations"
// into a valid Go identifier like "MathOperations".
func toGoIdentifier(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	var sb strings.Builder
	for _, w := range words {
		if len(w) > 0 {
			r := []rune(w)
			r[0] = unicode.ToUpper(r[0])
			sb.WriteString(string(r))
		}
	}
	return sb.String()
}

// TestFile returns the accumulated test functions as a Go _test.go file.
// If there are no test functions the empty string is returned.
func (tr *Transpiler) TestFile(src string) (string, error) {
	nodes, err := parser.ParseString(src)
	if err != nil {
		return "", err
	}
	for _, node := range nodes {
		if lst, ok := node.(*ast.List); ok {
			switch lst.HeadValue() {
			case "test":
				out, err := tr.transpileTest(lst)
				if err != nil {
					return "", err
				}
				tr.testFuncs = append(tr.testFuncs, out)
			case "bench":
				out, err := tr.transpileBench(lst)
				if err != nil {
					return "", err
				}
				tr.testFuncs = append(tr.testFuncs, out)
			}
		}
	}
	if len(tr.testFuncs) == 0 {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("package main\n\nimport \"testing\"\n\n")
	for _, f := range tr.testFuncs {
		sb.WriteString(f)
		sb.WriteString("\n\n")
	}
	formatted, err := format.Source([]byte(sb.String()))
	if err != nil {
		return sb.String(), fmt.Errorf("go/format test: %w", err)
	}
	return string(formatted), nil
}
