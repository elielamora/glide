# Glide Developer Guide

Welcome to Glide — the readable Lisp that compiles to Go. This guide walks new **gliders** through everything they need to know to write, run, and test Glide programs.

---

## Table of Contents

1. [What is Glide?](#what-is-glide)
2. [Installation](#installation)
3. [Your First Program](#your-first-program)
4. [Syntax Fundamentals](#syntax-fundamentals)
5. [Variables and Constants](#variables-and-constants)
6. [Functions](#functions)
7. [Control Flow](#control-flow)
8. [Loops and Iteration](#loops-and-iteration)
9. [Types and Structs](#types-and-structs)
10. [Error Handling](#error-handling)
11. [Concurrency](#concurrency)
12. [Threading Macros](#threading-macros)
13. [Imports](#imports)
14. [Testing](#testing)
15. [Transpiling and Building](#transpiling-and-building)
16. [Operator Reference](#operator-reference)

---

## What is Glide?

Glide is a **Lisp-style language that transpiles to idiomatic Go**. It gives you:

- **Go's power and performance** — Glide compiles through the Go toolchain, so every program is a real Go binary.
- **Lisp's expressiveness** — prefix S-expressions make operator precedence unambiguous and composition natural.
- **Readable indentation** — like Python, Glide uses off-side-rule indentation instead of extra parentheses, so deeply nested code stays clean.
- **First-class testing** — `test` and `assert` blocks are built into the language, not a library.

---

## Installation

### Prerequisites

- Go 1.21 or newer ([golang.org](https://golang.org/dl/))

### Build from source

```sh
git clone https://github.com/elielamora/glide
cd glide
go build -o glide .
# Optionally move the binary somewhere on your PATH:
sudo mv glide /usr/local/bin/
```

### Verify

```sh
glide --help
```

---

## Your First Program

Create a file `hello.glide`:

```glide
fn main ()
    fmt.Println "Hello, Glider!"
```

Run it:

```sh
glide run hello.glide
# → Hello, Glider!
```

That's it. No `package main`, no `import "fmt"` — Glide handles both automatically.

---

## Syntax Fundamentals

### S-expressions

Every Glide expression is an **S-expression**: a head symbol followed by its arguments, either on the same line inside parentheses or on indented lines below.

```glide
; inline — parentheses group the call
(+ 1 2)

; multi-line — indentation groups the call (no extra parens needed)
+
    1
    2
```

Both forms produce the same AST. You can mix them freely:

```glide
; (1 + 2) * (3 + 4)  in infix notation
* (+ 1 2) (+ 3 4)
```

### Comments

Lines starting with `;` are comments and are ignored by the compiler.

```glide
; This is a comment
fn main ()
    ; So is this
    fmt.Println "Hi" ; and this (inline)
```

### Indentation (off-side rule)

Any line indented **more** than its parent becomes a child expression of that parent. A standard 4-space indent is recommended, but any consistent deeper indent works.

```glide
; These two are identical:
(fmt.Printf "%d\n" (* 6 7))

fmt.Printf "%d\n"
    * 6 7
```

---

## Variables and Constants

### `let` — local variable binding

```glide
fn main ()
    let x 42
    let greeting "hello"
    let pi 3.14159
    fmt.Println x greeting pi
```

You can optionally supply an explicit type:

```glide
let count int 0
let ratio float64 0.5
```

### `set` — reassignment

```glide
fn main ()
    let counter 0
    set counter (+ counter 1)
    fmt.Println counter   ; 1
```

### `const` — compile-time constants

```glide
const MaxRetries 3
const AppName "glide-app"

fn main ()
    fmt.Println AppName MaxRetries
```

---

## Functions

### Basic function

```glide
fn add (a int b int) int
    + a b

fn main ()
    let result (add 3 4)
    fmt.Println result   ; 7
```

The signature is `fn name (param type …) returnType`.

### Multiple return values

```glide
fn divide (a float64 b float64) (float64 error)
    if (== b 0.0)
        (return 0.0 (errors.New "division by zero"))
        (return (/ a b) nil)

fn main ()
    let (result err) (divide 10.0 3.0)
    if (!= err nil)
        fmt.Println "Error:" err
        fmt.Printf "%.4f\n" result
```

### Methods on structs

```glide
type Rectangle struct
    Width  float64
    Height float64

fn (r Rectangle) .Area () float64
    * r.Width r.Height

fn (r Rectangle) .Perimeter () float64
    * 2.0 (+ r.Width r.Height)

fn main ()
    let rect (Rectangle 5.0 3.0)
    fmt.Printf "Area: %.1f, Perimeter: %.1f\n"
        .Area rect
        .Perimeter rect
```

### First-class functions and closures

```glide
fn makeAdder (n int) func(int) int
    fn (x int) int
        + x n

fn main ()
    let addFive (makeAdder 5)
    fmt.Println (addFive 10)   ; 15
    fmt.Println (addFive 3)    ; 8
```

---

## Control Flow

### `if` — conditional expression

`if` is an expression in Glide: both branches return a value.

```glide
fn abs (n int) int
    if (< n 0)
        (- 0 n)
        n

fn main ()
    fmt.Println (abs -7)   ; 7
    fmt.Println (abs  3)   ; 3
```

For side-effecting branches with no else:

```glide
fn main ()
    let x 10
    if (> x 5)
        fmt.Println "big"
```

---

## Loops and Iteration

### `for` — range loop

```glide
fn main ()
    for i in 1..=5
        fmt.Printf "i = %d\n" i
```

The `..=` syntax produces an inclusive range. Use `..` for exclusive (up to but not including the end).

```glide
; Prints 0, 1, 2, 3, 4
for i in 0..5
    fmt.Println i
```

### `each` — slice iteration

```glide
fn main ()
    let fruits ["apple" "banana" "cherry"]
    each [i fruit] in fruits
        fmt.Printf "%d: %s\n" i fruit
```

### `loop` / `recur` — tail-recursive loops

`loop` establishes named bindings; `recur` jumps back to the top with new values. This compiles to a Go `for` loop — no actual recursion occurs at runtime.

```glide
fn sumTo (n int) int
    loop (i n acc 0)
        if (<= i 0)
            acc
            recur (- i 1) (+ acc i)

fn main ()
    fmt.Println (sumTo 100)   ; 5050
```

---

## Types and Structs

### Defining a struct

```glide
type Point struct
    X float64
    Y float64
```

### Creating struct literals

```glide
fn main ()
    let p (Point 1.0 2.0)
    fmt.Printf "(%v, %v)\n" p.X p.Y
```

### Embedding and composition

```glide
type Animal struct
    Name string

type Dog struct
    Animal
    Breed string

fn main ()
    let d (Dog (Animal "Rex") "Labrador")
    fmt.Println d.Name d.Breed
```

---

## Error Handling

### `try` — propagate errors automatically

`(try expr)` evaluates `expr`, and if the last return value is a non-nil `error`, it immediately returns that error from the enclosing function (similar to `?` in Rust).

```glide
fn readConfig (path string) (string error)
    let data (try (os.ReadFile path))
    (return (string data) nil)

fn main ()
    let (cfg err) (readConfig "config.toml")
    if (!= err nil)
        fmt.Println "failed:" err
        fmt.Println cfg
```

### `guard` — handle errors inline

```glide
fn processFile (path string) error
    let (data err) (os.ReadFile path)
    guard err
        (return err)
    fmt.Println (string data)
    (return nil)
```

---

## Concurrency

### `go` — launch a goroutine

```glide
fn main ()
    let wg (sync.WaitGroup{})
    for i in 0..5
        .Add &wg 1
        go
            .Done &wg
            fmt.Println "goroutine" i
    .Wait &wg
```

### `chan` — create a channel

```glide
fn producer (ch (chan int))
    for i in 0..5
        <- ch i

fn main ()
    let ch (chan int 10)
    go (producer ch)
    for v in ch
        fmt.Println v
```

---

## Threading Macros

Threading macros pipe a value through a series of transformations, making deeply nested calls easy to read.

### `->` — thread first (value inserted as first argument)

```glide
; Without threading:
fmt.Println (strings.ToUpper (strings.TrimSpace "  hello world  "))

; With ->:
-> "  hello world  "
    strings.TrimSpace
    strings.ToUpper
    fmt.Println
```

### `->>` — thread last (value inserted as last argument)

```glide
; Compute sum of squares of even numbers
let result (->> [1 2 3 4 5 6]
    (filter even?)
    (map square)
    (reduce +))
```

---

## Imports

### Auto-import

Standard-library packages used via qualified calls (`fmt.Println`, `math.Sqrt`, etc.) are imported automatically — no `use` statement needed.

### `use` — explicit import

For packages outside the auto-import list, or to give an alias:

```glide
use "net/http"
use "net/http" web    ; import as 'web'

fn main ()
    let resp (try (web.Get "https://example.com"))
    fmt.Println resp.Status
```

---

## Testing

Glide has first-class test and benchmark support.

### `test` blocks

```glide
; math_test.glide
test "addition"
    assert (== (+ 2 2) 4)
    assert (== (+ 0 -5) -5)

test "multiplication"
    assert (== (* 3 4) 12)
    assert (== (* -2 5) -10)
```

Run with:

```sh
glide test math_test.glide
```

### `bench` blocks

```glide
bench "string concat"
    let _ (+ "hello" " " "world")
```

### `assert`

`assert` takes any boolean expression. On failure it reports the exact expression and panics the test.

```glide
assert (> result 0)
assert (== (len items) 3)
```

---

## Transpiling and Building

### See the generated Go

```sh
glide transpile myprogram.glide
```

This prints the fully formatted Go source to stdout — great for understanding what Glide produces or for debugging.

### Build a binary

```sh
glide build myprogram.glide
# produces ./myprogram (or myprogram.exe on Windows)
```

### Run directly

```sh
glide run myprogram.glide
```

---

## Operator Reference

| Glide         | Go equivalent | Notes                       |
|---------------|---------------|-----------------------------|
| `(+ a b)`     | `a + b`       | Works on numbers and strings |
| `(- a b)`     | `a - b`       |                             |
| `(* a b)`     | `a * b`       |                             |
| `(/ a b)`     | `a / b`       |                             |
| `(% a b)`     | `a % b`       |                             |
| `(== a b)`    | `a == b`      |                             |
| `(!= a b)`    | `a != b`      |                             |
| `(< a b)`     | `a < b`       |                             |
| `(<= a b)`    | `a <= b`      |                             |
| `(> a b)`     | `a > b`       |                             |
| `(>= a b)`    | `a >= b`      |                             |
| `(&& a b)`    | `a && b`      |                             |
| `(\|\| a b)`  | `a \|\| b`    |                             |
| `(! a)`       | `!a`          | Logical NOT                 |
| `(& a b)`     | `a & b`       | Bitwise AND / address-of    |
| `(<- ch v)`   | `ch <- v`     | Channel send                |

---

*Happy gliding! 🚀*
