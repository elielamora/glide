import React from "react";
import { Link } from "gatsby";
import Layout from "../components/Layout";
import CodeBlock from "../components/CodeBlock";

export const Head = () => (
  <>
    <title>Examples — Glide</title>
    <meta
      name="description"
      content="Real Glide code examples: hello world, fibonacci, structs, error handling, concurrency, threading macros, and built-in testing."
    />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </>
);

/* ── Example data ───────────────────────────────────────────────────────── */
const EXAMPLES = [
  {
    category: "Getting Started",
    items: [
      {
        title: "Hello, World",
        description:
          "The simplest Glide program. No package declaration, no imports — Glide adds them automatically.",
        filename: "hello.glide",
        code: `fn main ()
    fmt.Println "Hello, Glider! 🚀"`,
        output: "Hello, Glider! 🚀",
      },
      {
        title: "Variables & Constants",
        description:
          "Use let to bind local variables, set to reassign, and const for compile-time constants.",
        filename: "variables.glide",
        code: `const Pi 3.14159

fn circleArea (r float64) float64
    * Pi (* r r)

fn main ()
    let r 5.0
    let area (circleArea r)
    fmt.Printf "Area of circle with r=%.0f is %.2f\\n" r area`,
        output: "Area of circle with r=5 is 78.54",
      },
    ],
  },
  {
    category: "Functions",
    items: [
      {
        title: "Fibonacci (recursive)",
        description:
          "Recursive function definition. if is an expression — both branches return a value.",
        filename: "fibonacci.glide",
        code: `fn fib (n int) int
    if (< n 2)
        n
        + (fib (- n 1)) (fib (- n 2))

fn main ()
    for i in 0..=10
        fmt.Printf "fib(%2d) = %d\\n" i (fib i)`,
        output: `fib( 0) = 0
fib( 1) = 1
fib( 2) = 1
fib( 3) = 2
fib( 5) = 5
fib(10) = 55`,
      },
      {
        title: "Closures",
        description:
          "Functions are first-class values. makeAdder returns a closure that captures n.",
        filename: "closures.glide",
        code: `fn makeAdder (n int) func(int) int
    fn (x int) int
        + x n

fn makeMultiplier (n int) func(int) int
    fn (x int) int
        * x n

fn main ()
    let addTen  (makeAdder 10)
    let triple  (makeMultiplier 3)
    fmt.Println (addTen 5)     ; 15
    fmt.Println (triple 7)     ; 21
    fmt.Println (addTen (triple 4))  ; 22`,
        output: "15\n21\n22",
      },
      {
        title: "Multiple Return Values",
        description:
          "Functions can return multiple values. Destructure them with let.",
        filename: "multi_return.glide",
        code: `fn divmod (a int b int) (int int)
    (return (/ a b) (% a b))

fn main ()
    let (q r) (divmod 17 5)
    fmt.Printf "17 ÷ 5 = %d remainder %d\\n" q r`,
        output: "17 ÷ 5 = 3 remainder 2",
      },
    ],
  },
  {
    category: "Loops",
    items: [
      {
        title: "Factorial with loop/recur",
        description:
          "loop/recur is Glide's tail-call construct. It compiles to a plain Go for loop — zero stack overhead.",
        filename: "factorial.glide",
        code: `fn factorial (n int) int
    loop (current n acc 1)
        if (<= current 1)
            acc
            recur (- current 1) (* acc current)

fn main ()
    for i in 1..=12
        fmt.Printf "%2d! = %d\\n" i (factorial i)`,
        output: " 1! = 1\n 2! = 2\n 3! = 6\n...12! = 479001600",
      },
      {
        title: "each — slice iteration",
        description:
          "each iterates over a slice giving you both the index and the value.",
        filename: "each.glide",
        code: `fn main ()
    let langs ["Go" "Glide" "Rust" "Zig"]
    each [i lang] in langs
        fmt.Printf "%d. %s\\n" (+ i 1) lang`,
        output: "1. Go\n2. Glide\n3. Rust\n4. Zig",
      },
    ],
  },
  {
    category: "Types & Structs",
    items: [
      {
        title: "Structs and Methods",
        description:
          "Define types with type…struct. Add methods with fn (recv Type) .MethodName.",
        filename: "structs.glide",
        code: `type Point struct
    X float64
    Y float64

fn distance (a Point b Point) float64
    let dx (- b.X a.X)
    let dy (- b.Y a.Y)
    math.Sqrt (+ (* dx dx) (* dy dy))

fn (p Point) .String () string
    fmt.Sprintf "(%.1f, %.1f)" p.X p.Y

fn main ()
    let origin (Point 0.0 0.0)
    let target (Point 3.0 4.0)
    fmt.Printf "Distance from %v to %v = %.1f\\n"
        (.String origin)
        (.String target)
        (distance origin target)`,
        output: "Distance from (0.0, 0.0) to (3.0, 4.0) = 5.0",
      },
    ],
  },
  {
    category: "Error Handling",
    items: [
      {
        title: "try — error propagation",
        description:
          "try evaluates an expression and propagates any non-nil error to the caller, like Rust's ? operator.",
        filename: "try.glide",
        code: `fn readFile (path string) (string error)
    let data (try (os.ReadFile path))
    (return (string data) nil)

fn main ()
    let (content err) (readFile "README.md")
    if (!= err nil)
        fmt.Println "Error:" err
        fmt.Printf "Read %d bytes\\n" (len content)`,
        output: "Read 842 bytes",
      },
      {
        title: "guard — inline error handling",
        description:
          "guard checks an error value inline and executes the recovery block if non-nil.",
        filename: "guard.glide",
        code: `fn safeDiv (a float64 b float64) (float64 error)
    if (== b 0.0)
        (return 0.0 (errors.New "division by zero"))
        (return (/ a b) nil)

fn main ()
    let (result err) (safeDiv 10.0 0.0)
    guard err
        fmt.Println "Caught error:" err
        (return)
    fmt.Printf "Result: %.2f\\n" result`,
        output: "Caught error: division by zero",
      },
    ],
  },
  {
    category: "Concurrency",
    items: [
      {
        title: "Goroutines & Channels",
        description:
          "go launches a goroutine. chan creates a buffered channel. <- sends and receives.",
        filename: "concurrency.glide",
        code: `fn worker (id int results (chan string))
    ; Simulate work
    let msg (fmt.Sprintf "worker %d done" id)
    <- results msg

fn main ()
    let results (chan string 5)
    for i in 0..5
        go (worker i results)
    for i in 0..5
        fmt.Println (<- results)`,
        output: "worker 2 done\nworker 0 done\nworker 4 done\n...",
      },
    ],
  },
  {
    category: "Threading Macros",
    items: [
      {
        title: "-> (thread first)",
        description:
          "-> threads a value as the first argument through each step. Read top to bottom instead of inside-out.",
        filename: "threading.glide",
        code: `fn main ()
    ; Without threading — read inside-out:
    let before (strings.ToUpper (strings.TrimSpace "  hello world  "))
    fmt.Println before

    ; With -> — read top to bottom:
    let after
        -> "  hello world  "
            strings.TrimSpace
            strings.ToUpper
    fmt.Println after`,
        output: "HELLO WORLD\nHELLO WORLD",
      },
    ],
  },
  {
    category: "Testing",
    items: [
      {
        title: "Built-in Test Blocks",
        description:
          "test and assert are keywords. Run the file with glide test — no framework, no configuration.",
        filename: "math_test.glide",
        code: `test "basic arithmetic"
    assert (== (+ 2 2) 4)
    assert (== (* 3 5) 15)
    assert (== (- 10 3) 7)
    assert (== (/ 10 2) 5)

test "comparisons"
    assert (> 5 3)
    assert (< 1 10)
    assert (>= 5 5)
    assert (!= 1 2)

test "nested expressions"
    ; (1 + 2) * 3 = 9
    assert (== (* (+ 1 2) 3) 9)`,
        output:
          "--- PASS: TestBasic_basic_arithmetic (0.00s)\n--- PASS: TestBasic_comparisons (0.00s)\n--- PASS: TestBasic_nested_expressions (0.00s)\nPASS",
      },
      {
        title: "Benchmarks",
        description:
          "bench blocks run via glide test and integrate with Go's benchmark framework.",
        filename: "bench_test.glide",
        code: `fn fibIter (n int) int
    loop (i n a 0 b 1)
        if (<= i 0)
            a
            recur (- i 1) b (+ a b)

bench "fibonacci iterative"
    let _ (fibIter 30)

test "fib(10) = 55"
    assert (== (fibIter 10) 55)`,
        output: "BenchmarkBasic_fibonacci_iterative\n637 ns/op",
      },
    ],
  },
];

/* ── Category badge ─────────────────────────────────────────────────────── */
const categoryColors = {
  "Getting Started": "bg-green-100 text-green-700",
  Functions: "bg-blue-100 text-blue-700",
  Loops: "bg-orange-100 text-orange-700",
  "Types & Structs": "bg-purple-100 text-purple-700",
  "Error Handling": "bg-red-100 text-red-700",
  Concurrency: "bg-teal-100 text-teal-700",
  "Threading Macros": "bg-indigo-100 text-indigo-700",
  Testing: "bg-yellow-100 text-yellow-700",
};

const ExamplesPage = () => (
  <Layout>
    {/* Page header */}
    <div className="bg-gradient-to-br from-glide-950 to-indigo-900 text-white py-14">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <h1 className="text-4xl sm:text-5xl font-extrabold mb-4">
          Glide by Example
        </h1>
        <p className="text-glide-200 text-lg max-w-2xl mx-auto">
          Runnable programs covering every major feature. Copy any example,
          paste it into a <code className="font-mono text-glide-100">.glide</code> file,
          and run it with{" "}
          <code className="font-mono text-glide-100">glide run</code>.
        </p>
      </div>
    </div>

    <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      {EXAMPLES.map(({ category, items }) => (
        <section key={category} className="mb-16">
          <h2 className="text-2xl font-extrabold text-gray-900 mb-6 flex items-center gap-3">
            <span
              className={`px-3 py-1 rounded-full text-sm font-semibold ${categoryColors[category] || "bg-gray-100 text-gray-700"}`}
            >
              {category}
            </span>
          </h2>

          <div className="space-y-10">
            {items.map(({ title, description, filename, code, output }) => (
              <div
                key={title}
                className="rounded-2xl border border-gray-200 overflow-hidden shadow-sm hover:shadow-md transition-shadow"
              >
                {/* Card header */}
                <div className="px-6 py-4 bg-gray-50 border-b border-gray-200">
                  <h3 className="text-lg font-bold text-gray-900">{title}</h3>
                  <p className="text-gray-600 text-sm mt-1">{description}</p>
                </div>

                {/* Code */}
                <div className="p-6 bg-white">
                  <CodeBlock code={code} filename={filename} />

                  {/* Output */}
                  <div className="mt-4">
                    <p className="text-xs font-semibold uppercase tracking-wider text-gray-400 mb-2">
                      Output
                    </p>
                    <pre className="bg-gray-900 rounded-lg p-4 text-sm text-green-400 font-mono leading-relaxed overflow-x-auto">
                      {output}
                    </pre>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </section>
      ))}

      {/* CTA */}
      <div className="mt-8 p-8 rounded-2xl bg-gradient-to-br from-glide-700 to-indigo-700 text-white text-center">
        <h3 className="text-2xl font-bold mb-2">Want to learn more?</h3>
        <p className="text-glide-100 mb-6">
          The developer guide covers every language feature in depth.
        </p>
        <Link
          to="/guide"
          className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-white text-glide-700 font-bold hover:bg-glide-50 transition-all shadow-lg"
        >
          Read the Developer Guide →
        </Link>
      </div>
    </div>
  </Layout>
);

export default ExamplesPage;
