// Command glide is the Glide language toolchain CLI.
//
// Usage:
//
//	glide run <file.glide>       Transpile and run a Glide program
//	glide build <file.glide>     Transpile and compile a Glide program
//	glide transpile <file.glide> Print the generated Go source to stdout
//	glide test <file.glide>      Extract and run test blocks
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/elielamora/glide/internal/transpiler"
)

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	file := os.Args[2]

	src, err := os.ReadFile(file)
	if err != nil {
		fatalf("read %s: %v", file, err)
	}

	switch cmd {
	case "transpile":
		goSrc, err := transpiler.TranspileFile(string(src))
		if err != nil {
			fatalf("transpile: %v", err)
		}
		fmt.Print(goSrc)

	case "run":
		goSrc, err := transpiler.TranspileFile(string(src))
		if err != nil {
			fatalf("transpile: %v", err)
		}
		runGoSource(goSrc, false)

	case "build":
		goSrc, err := transpiler.TranspileFile(string(src))
		if err != nil {
			fatalf("transpile: %v", err)
		}
		outName := strings.TrimSuffix(filepath.Base(file), ".glide")
		buildGoSource(goSrc, outName)

	case "test":
		tr := transpiler.New()
		testSrc, err := tr.TestFile(string(src))
		if err != nil {
			fatalf("test transpile: %v", err)
		}
		if testSrc == "" {
			fmt.Fprintln(os.Stderr, "no test blocks found")
			os.Exit(1)
		}
		runGoSource(testSrc, true)

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", cmd)
		usage()
		os.Exit(1)
	}
}

func runGoSource(goSrc string, isTest bool) {
	dir, err := os.MkdirTemp("", "glide-*")
	if err != nil {
		fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(dir)

	srcFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcFile, []byte(goSrc), 0600); err != nil {
		fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), goModContent(), 0600); err != nil {
		fatalf("write go.mod: %v", err)
	}

	var c *exec.Cmd
	if isTest {
		c = exec.Command("go", "test", "-v", "./...")
	} else {
		c = exec.Command("go", "run", srcFile)
	}
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		os.Exit(1)
	}
}

func buildGoSource(goSrc, outName string) {
	dir, err := os.MkdirTemp("", "glide-*")
	if err != nil {
		fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(dir)

	srcFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcFile, []byte(goSrc), 0600); err != nil {
		fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), goModContent(), 0600); err != nil {
		fatalf("write go.mod: %v", err)
	}

	wd, _ := os.Getwd()
	outPath := filepath.Join(wd, outName)
	c := exec.Command("go", "build", "-o", outPath, srcFile)
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		os.Exit(1)
	}
	fmt.Printf("built %s\n", outPath)
}

func goModContent() []byte {
	// Derive the Go version from the running toolchain (e.g. "go1.24.1" → "1.24.1").
	ver := strings.TrimPrefix(runtime.Version(), "go")
	return []byte("module gliderun\n\ngo " + ver + "\n")
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "glide: "+format+"\n", args...)
	os.Exit(1)
}

func usage() {
	fmt.Fprintln(os.Stderr, `glide — the Glide language toolchain

Usage:
  glide transpile <file.glide>   Print generated Go source to stdout
  glide run       <file.glide>   Transpile and run
  glide build     <file.glide>   Transpile and compile to binary
  glide test      <file.glide>   Run test blocks`)
}
