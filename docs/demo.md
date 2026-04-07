# Glide — The Go Lisp You've Been Waiting For

> **Write less. Express more. Deploy as Go.**

Glide is a modern Lisp that compiles directly to idiomatic, `gofmt`-clean Go source. You get the elegance of a functional language and the raw performance of the Go runtime — no JVM, no interpreter, no GC pauses you didn't sign up for.

---

## Why Glide?

| Challenge | What most languages give you | What Glide gives you |
|-----------|------------------------------|----------------------|
| Concurrency | Callback hell or async/await | Go goroutines, expressed in two words: `go expr` |
| Error handling | Try/catch that hides control flow | `try` that makes propagation visible, `guard` for inline recovery |
| Operator precedence | Parentheses fights, or memorising tables | Prefix S-expressions — unambiguous by design |
| Build toolchain | Separate compilers, runtimes, package managers | One binary that transpiles, runs, builds, and tests |
| Readability | Lisp: too many parens. Python: significant whitespace | Glide: significant whitespace *and* only as many parens as you want |

---

## See it in Action

### Hello, World

```glide
fn main ()
    fmt.Println "Hello, Glider! 🚀"
```

```sh
$ glide run hello.glide
Hello, Glider! 🚀
```

No boilerplate. `package main` and `import "fmt"` are inferred. You write the idea, Glide writes the ceremony.

---

### Fibonacci — Elegance Meets Performance

```glide
fn fib (n int) int
    if (< n 2)
        n
        + (fib (- n 1)) (fib (- n 2))

fn main ()
    for i in 0..=10
        fmt.Printf "fib(%d) = %d\n" i (fib i)
```

```
fib(0) = 0
fib(1) = 1
fib(2) = 1
fib(3) = 2
fib(4) = 3
fib(5) = 5
fib(6) = 8
fib(7) = 13
fib(8) = 21
fib(9) = 34
fib(10) = 55
```

`if` is an expression. Recursion reads naturally. And it compiles to a Go function — so it's as fast as Go.

---

### Tail-Recursive Factorial — Zero Stack Overhead

Glide's `loop`/`recur` construct compiles to a plain Go `for` loop. No stack frames. No overflow. Pure tail-call optimisation by the compiler.

```glide
fn factorial (n int) int
    loop (current n acc 1)
        if (<= current 1)
            acc
            recur (- current 1) (* acc current)

fn main ()
    for i in 1..=12
        fmt.Printf "%2d! = %d\n" i (factorial i)
```

```
 1! = 1
 2! = 2
 3! = 6
 4! = 24
 5! = 120
 6! = 720
 7! = 5040
 8! = 40320
 9! = 362880
10! = 3628800
11! = 39916800
12! = 479001600
```

---

### Structs and Methods — Object-Oriented Without the Objects

```glide
type Point struct
    X float64
    Y float64

fn distance (p1 Point p2 Point) float64
    let dx (- p2.X p1.X)
    let dy (- p2.Y p1.Y)
    math.Sqrt (+ (* dx dx) (* dy dy))

fn (p Point) .String () string
    fmt.Sprintf "(%.1f, %.1f)" p.X p.Y

fn main ()
    let a (Point 0.0 0.0)
    let b (Point 3.0 4.0)
    fmt.Printf "Distance from %v to %v = %.1f\n"
        (.String a)
        (.String b)
        (distance a b)
```

```
Distance from (0.0, 0.0) to (3.0, 4.0) = 5.0
```

Methods live outside the type definition — in Glide you can add methods anywhere in the file. The receiver syntax `(fn (r ReceiverType) .MethodName …)` makes it immediately clear what type each method belongs to.

---

### Threading Macros — Readable Pipelines

Deeply nested function calls are a pain to read. Threading macros solve this:

```glide
; Without threading — read inside-out:
fmt.Println
    strings.ToUpper
        strings.TrimSpace "  hello, world  "

; With -> — read left to right, top to bottom:
-> "  hello, world  "
    strings.TrimSpace
    strings.ToUpper
    fmt.Println
```

```
HELLO, WORLD
```

`->` threads the value as the **first** argument. `->>` threads it as the **last** argument — perfect for slice pipelines.

---

### Concurrency — Goroutines as a First-Class Expression

```glide
fn worker (id int ch (chan string))
    <- ch (fmt.Sprintf "worker %d done" id)

fn main ()
    let results (chan string 5)
    for i in 0..5
        go (worker i results)
    for i in 0..5
        fmt.Println (<- results)
```

```
worker 3 done
worker 0 done
worker 4 done
worker 1 done
worker 2 done
```

`go expr` launches a goroutine. `<- ch val` sends. `(<- ch)` receives. The full power of Go's concurrency model, expressed concisely.

---

### Error Handling — Explicit, Not Exceptional

```glide
fn readLines (path string) ([]string error)
    let (data err) (os.ReadFile path)
    guard err
        (return nil err)
    let lines (strings.Split (string data) "\n")
    (return lines nil)

fn main ()
    let (lines err) (readLines "notes.txt")
    if (!= err nil)
        fmt.Println "Error:" err
        each [i line] in lines
            fmt.Printf "%d: %s\n" (+ i 1) line
```

No exceptions. No hidden control flow. Every error path is visible in the code.

Use `try` to propagate errors automatically (like Rust's `?`):

```glide
fn readLines (path string) ([]string error)
    let data (try (os.ReadFile path))
    (return (strings.Split (string data) "\n") nil)
```

---

### Built-in Testing — No Framework Required

```glide
; geometry_test.glide

fn hypotenuse (a float64 b float64) float64
    math.Sqrt (+ (* a a) (* b b))

test "3-4-5 triangle"
    assert (== (hypotenuse 3.0 4.0) 5.0)

test "isoceles right triangle"
    let h (hypotenuse 1.0 1.0)
    assert (> h 1.41)
    assert (< h 1.42)

test "zero legs"
    assert (== (hypotenuse 0.0 0.0) 0.0)
```

```sh
$ glide test geometry_test.glide
--- PASS: TestBasic_3-4-5_triangle (0.00s)
--- PASS: TestBasic_isoceles_right_triangle (0.00s)
--- PASS: TestBasic_zero_legs (0.00s)
PASS
```

Tests are first-class language constructs. Run them with `glide test` — no test runner, no configuration files.

---

### A Real-World Snippet — Word Frequency Counter

```glide
fn wordCount (text string) map[string]int
    let counts (map[string]int{})
    let words (strings.Fields text)
    each [_ word] in words
        let lower (strings.ToLower word)
        set (counts lower) (+ (counts lower) 1)
    counts

fn main ()
    let text "the quick brown fox jumps over the lazy dog the fox"
    let counts (wordCount text)
    ; Print in sorted order
    let keys (make []string 0 (len counts))
    for k in counts
        set keys (append keys k)
    sort.Strings keys
    each [_ k] in keys
        fmt.Printf "%-10s %d\n" k (counts k)
```

```
brown      1
dog        1
fox        2
jumps      1
lazy       1
over       1
quick      1
the        3
```

---

## What Gets Generated?

Run `glide transpile` to see the Go output. Here's the factorial example:

**Input (`factorial.glide`):**
```glide
fn factorial (n int) int
    loop (current n acc 1)
        if (<= current 1)
            acc
            recur (- current 1) (* acc current)
```

**Output (Go):**
```go
func factorial(n int) int {
    current, acc := n, 1
    for {
        if current <= 1 {
            return acc
        }
        current, acc = current-1, acc*current
    }
}
```

Clean, idiomatic Go. No runtime library. No extra allocations. The generated code is code you'd be proud to have written yourself.

---

## Getting Started in 60 Seconds

```sh
# 1. Clone and build
git clone https://github.com/elielamora/glide
cd glide && go build -o glide .

# 2. Write your first program
echo 'fn main ()\n    fmt.Println "I am a Glider!"' > first.glide

# 3. Run it
./glide run first.glide

# 4. See the generated Go
./glide transpile first.glide

# 5. Compile to a binary
./glide build first.glide && ./first
```

---

## Cheat Sheet

```glide
; Variables
let x 42
let name string "Alice"
set x (+ x 1)
const Pi 3.14159

; Functions
fn greet (name string) string
    fmt.Sprintf "Hello, %s!" name

; Methods
fn (r Rect) .Area () float64
    * r.Width r.Height

; Control
if (> x 0)
    fmt.Println "positive"
    fmt.Println "non-positive"

; Ranges
for i in 0..10        ; exclusive end
for i in 1..=10       ; inclusive end

; Slices
each [idx val] in mySlice
    fmt.Println idx val

; Tail-recursive loop
loop (n 10 acc 0)
    if (== n 0) acc (recur (- n 1) (+ acc n))

; Error handling
let data (try (os.ReadFile "file.txt"))     ; propagate
guard err (return err)                      ; inline

; Goroutine & channel
go (myFunc arg)
let ch (chan int 10)
<- ch value
let v (<- ch)

; Threading
-> value f1 f2 f3          ; thread first
->> value f1 f2 f3         ; thread last

; Testing
test "my test"
    assert (== result expected)
```

---

## Next Steps

- 📖 Read the full [Developer Guide](guide.md) for in-depth coverage of every language feature.
- 🔭 Explore the [`examples/`](../examples/) directory for ready-to-run programs.
- 🛠️ Run `glide transpile <file>` on any example to see the generated Go.
- 🐛 Found a bug or have an idea? Open an issue on [GitHub](https://github.com/elielamora/glide).

---

*Glide — because great ideas deserve elegant syntax.*
