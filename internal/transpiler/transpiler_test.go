package transpiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

// transpileFile calls TranspileFile and returns the Go source.
func transpileFile(t *testing.T, src string) string {
	t.Helper()
	out, err := TranspileFile(src)
	if err != nil {
		t.Fatalf("TranspileFile error: %v\nsource:\n%s", err, out)
	}
	return out
}

// mustContain fails the test if the Go source does not contain the needle.
func mustContain(t *testing.T, goSrc, needle string) {
	t.Helper()
	if !strings.Contains(goSrc, needle) {
		t.Errorf("expected Go source to contain %q\ngot:\n%s", needle, goSrc)
	}
}

// mustNotContain fails the test if the Go source contains the needle.
func mustNotContain(t *testing.T, goSrc, needle string) {
	t.Helper()
	if strings.Contains(goSrc, needle) {
		t.Errorf("expected Go source NOT to contain %q\ngot:\n%s", needle, goSrc)
	}
}

// compileAndRun writes goSrc to a temp dir, compiles it with 'go run', and
// returns stdout + stderr.  The test fails if compilation fails.
func compileAndRun(t *testing.T, goSrc string) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go binary not in PATH, skipping execution test")
	}

	dir := t.TempDir()
	srcFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcFile, []byte(goSrc), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	// Derive the Go version from the running toolchain (e.g. "go1.24.1" → "1.24.1").
	goVer := strings.TrimPrefix(runtime.Version(), "go")

	goMod := "module tmpglide\n\ngo " + goVer + "\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0600); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	cmd := exec.Command("go", "run", srcFile)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run failed: %v\noutput:\n%s\nsource:\n%s", err, out, goSrc)
	}
	return string(out)
}

// --------------------------------------------------------------------------
// Literal and variable tests
// --------------------------------------------------------------------------

func TestTranspile_letInt(t *testing.T) {
	goSrc := transpileFile(t, "let x 5\nfmt.Println x")
	mustContain(t, goSrc, "x :=")
	mustContain(t, goSrc, "5")
}

func TestTranspile_letExplicitType(t *testing.T) {
	goSrc := transpileFile(t, "let age int 30")
	mustContain(t, goSrc, "var age int = 30")
}

func TestTranspile_letString(t *testing.T) {
	goSrc := transpileFile(t, `let name "Alice"`)
	mustContain(t, goSrc, `"Alice"`)
}

func TestTranspile_constDecl(t *testing.T) {
	goSrc := transpileFile(t, "const pi 3.14159")
	mustContain(t, goSrc, "const pi = 3.14159")
}

func TestTranspile_setAssign(t *testing.T) {
	goSrc := transpileFile(t, "let age int 30\nset age 31")
	mustContain(t, goSrc, "age = 31")
}

// --------------------------------------------------------------------------
// Arithmetic expression tests
// --------------------------------------------------------------------------

func TestTranspile_addExpr(t *testing.T) {
	goSrc := transpileFile(t, "let result (+ 1 2)")
	mustContain(t, goSrc, "(1 + 2)")
}

func TestTranspile_mulExpr(t *testing.T) {
	goSrc := transpileFile(t, "let result (* 3 4)")
	mustContain(t, goSrc, "(3 * 4)")
}

func TestTranspile_nestedArith(t *testing.T) {
	// + 1 (* 2 3)  ≡  (+ 1 (* 2 3))  ≡  1 + (2 * 3)
	goSrc := transpileFile(t, "+ 1\n    * 2 3")
	mustContain(t, goSrc, "(1 + (2 * 3))")
}

func TestTranspile_compareOps(t *testing.T) {
	goSrc := transpileFile(t, "let r (> 5 3)")
	mustContain(t, goSrc, "(5 > 3)")
}

func TestTranspile_eqOp(t *testing.T) {
	goSrc := transpileFile(t, "let r (== 1 1)")
	mustContain(t, goSrc, "(1 == 1)")
}

// --------------------------------------------------------------------------
// Function definition tests
// --------------------------------------------------------------------------

func TestTranspile_simpleFn(t *testing.T) {
	src := "fn add (a int b int) int\n    + a b"
	goSrc := transpileFile(t, src)
	mustContain(t, goSrc, "func add(")
	mustContain(t, goSrc, "a int")
	mustContain(t, goSrc, "b int")
	mustContain(t, goSrc, "return (a + b)")
}

func TestTranspile_fnWithLet(t *testing.T) {
	src := "fn double (x int) int\n    let result (* x 2)\n    result"
	goSrc := transpileFile(t, src)
	mustContain(t, goSrc, "result :=")
	mustContain(t, goSrc, "(x * 2)")
	mustContain(t, goSrc, "return result")
}

func TestTranspile_voidFn(t *testing.T) {
	src := "fn greet ()\n    fmt.Println \"hello\""
	goSrc := transpileFile(t, src)
	mustContain(t, goSrc, "func greet()")
	mustNotContain(t, goSrc, "return fmt.Println")
}

// --------------------------------------------------------------------------
// If expression tests
// --------------------------------------------------------------------------

func TestTranspile_ifStmt(t *testing.T) {
	src := `fn classify (age int) string
    if (> age 18)
        "Adult"
        "Minor"`
	goSrc := transpileFile(t, src)
	// go/format removes the outer parens from conditions, so expect `age > 18`
	mustContain(t, goSrc, "age > 18")
	mustContain(t, goSrc, `return "Adult"`)
	mustContain(t, goSrc, `return "Minor"`)
}


func TestTranspile_ifLetExpr(t *testing.T) {
	src := `let status
    if (> 20 18)
        "Adult"
        "Minor"`
	goSrc := transpileFile(t, src)
	mustContain(t, goSrc, `"Adult"`)
	mustContain(t, goSrc, `"Minor"`)
}

// --------------------------------------------------------------------------
// For / each loop tests
// --------------------------------------------------------------------------

func TestTranspile_forInclusiveRange(t *testing.T) {
	src := "for i in 1..=5\n    fmt.Println i"
	goSrc := transpileFile(t, src)
	mustContain(t, goSrc, "for i := 1; i <= 5; i++")
}

func TestTranspile_forExclusiveRange(t *testing.T) {
	src := "for i in 0..10\n    fmt.Println i"
	goSrc := transpileFile(t, src)
	mustContain(t, goSrc, "for i := 0; i < 10; i++")
}

func TestTranspile_each(t *testing.T) {
	src := `let names ["Alice" "Bob"]
each [i name] in names
    fmt.Println i name`
	goSrc := transpileFile(t, src)
	mustContain(t, goSrc, "for i, name := range")
}

// --------------------------------------------------------------------------
// Type / struct tests
// --------------------------------------------------------------------------

func TestTranspile_structType(t *testing.T) {
	src := "type User struct\n    Name string\n    Age  int"
	goSrc := transpileFile(t, src)
	mustContain(t, goSrc, "type User struct {")
	// go/format may pad field names for alignment, so check name and type separately
	mustContain(t, goSrc, "Name")
	mustContain(t, goSrc, "string")
	mustContain(t, goSrc, "Age")
	mustContain(t, goSrc, "int")
}

// --------------------------------------------------------------------------
// loop / recur (tail-recursive loop)
// --------------------------------------------------------------------------

func TestTranspile_loopRecur(t *testing.T) {
	src := `fn factorial (n int) int
    loop (current n acc 1)
        if (<= current 1)
            acc
            recur (- current 1) (* acc current)`
	goSrc := transpileFile(t, src)
	mustContain(t, goSrc, "for {")
	mustContain(t, goSrc, "continue")
	mustContain(t, goSrc, "return acc")
}

// --------------------------------------------------------------------------
// Use / imports
// --------------------------------------------------------------------------

func TestTranspile_use(t *testing.T) {
	src := "use net.http web\nweb.ListenAndServe \":8080\" nil"
	goSrc := transpileFile(t, src)
	mustContain(t, goSrc, `"net/http"`)
}

// --------------------------------------------------------------------------
// Test block generation
// --------------------------------------------------------------------------

func TestTranspile_testBlock(t *testing.T) {
	src := `test "math operations"
    assert (== (+ 2 2) 4)
    assert (!= (+ 2 2) 5)`
	tr := New()
	testSrc, err := tr.TestFile(src)
	if err != nil {
		t.Fatalf("TestFile error: %v\nsrc:\n%s", err, testSrc)
	}
	mustContain(t, testSrc, "func TestMathOperations")
	mustContain(t, testSrc, "(t *testing.T)")
	mustContain(t, testSrc, "(2 + 2) == 4")
	mustContain(t, testSrc, "(2 + 2) != 5")
}

// --------------------------------------------------------------------------
// Threading macro
// --------------------------------------------------------------------------

func TestTranspile_threadFirst(t *testing.T) {
	src := `let result
    -> "hello"
       strings.ToUpper
       fmt.Println`
	goSrc := transpileFile(t, src)
	mustContain(t, goSrc, "strings.ToUpper")
	mustContain(t, goSrc, "fmt.Println")
}

// --------------------------------------------------------------------------
// End-to-end: compile and run generated Go IR
// --------------------------------------------------------------------------

func TestE2E_helloWorld(t *testing.T) {
	src := `fn main ()
    fmt.Println "Hello, Glide!"`
	goSrc := transpileFile(t, src)
	output := compileAndRun(t, goSrc)
	if !strings.Contains(output, "Hello, Glide!") {
		t.Errorf("expected 'Hello, Glide!' in output, got: %q", output)
	}
}

func TestE2E_arithmetic(t *testing.T) {
	src := `fn main ()
    let x (+ 3 4)
    fmt.Println x`
	goSrc := transpileFile(t, src)
	output := compileAndRun(t, goSrc)
	if !strings.Contains(output, "7") {
		t.Errorf("expected 7 in output, got: %q", output)
	}
}

func TestE2E_factorial(t *testing.T) {
	src := `fn factorial (n int) int
    loop (current n acc 1)
        if (<= current 1)
            acc
            recur (- current 1) (* acc current)

fn main ()
    let result (factorial 5)
    fmt.Println result`
	goSrc := transpileFile(t, src)
	output := compileAndRun(t, goSrc)
	if !strings.Contains(output, "120") {
		t.Errorf("expected 120 in output, got: %q\ngenerated Go:\n%s", output, goSrc)
	}
}

func TestE2E_stringConcat(t *testing.T) {
	src := `fn greet (name string) string
    + "Hello, " name

fn main ()
    fmt.Println (greet "World")`
	goSrc := transpileFile(t, src)
	output := compileAndRun(t, goSrc)
	if !strings.Contains(output, "Hello, World") {
		t.Errorf("expected 'Hello, World' in output, got: %q\ngenerated Go:\n%s", output, goSrc)
	}
}

func TestE2E_forLoop(t *testing.T) {
	src := `fn main ()
    for i in 1..=3
        fmt.Println i`
	goSrc := transpileFile(t, src)
	output := compileAndRun(t, goSrc)
	for _, want := range []string{"1", "2", "3"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in output, got: %q\ngenerated Go:\n%s", want, output, goSrc)
		}
	}
}

func TestE2E_ifElse(t *testing.T) {
	src := `fn classify (n int) string
    if (> n 0)
        "positive"
        "non-positive"

fn main ()
    fmt.Println (classify 5)
    fmt.Println (classify (- 0 1))`
	goSrc := transpileFile(t, src)
	output := compileAndRun(t, goSrc)
	if !strings.Contains(output, "positive") {
		t.Errorf("expected 'positive' in output, got: %q\ngenerated Go:\n%s", output, goSrc)
	}
	if !strings.Contains(output, "non-positive") {
		t.Errorf("expected 'non-positive' in output, got: %q\ngenerated Go:\n%s", output, goSrc)
	}
}

func TestE2E_struct(t *testing.T) {
	src := `type Point struct
    X int
    Y int

fn main ()
    let p (Point 3 4)
    fmt.Println p.X p.Y`
	goSrc := transpileFile(t, src)
	// Struct literal generation: Point{3, 4}
	mustContain(t, goSrc, "type Point struct")
	output := compileAndRun(t, goSrc)
	if !strings.Contains(output, "3") || !strings.Contains(output, "4") {
		t.Errorf("expected '3 4' in output, got: %q\ngenerated Go:\n%s", output, goSrc)
	}
}
