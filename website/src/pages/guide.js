import React, { useState } from "react";
import Layout from "../components/Layout";
import CodeBlock from "../components/CodeBlock";

export const Head = () => (
  <>
    <title>Developer Guide — Glide</title>
    <meta
      name="description"
      content="Complete developer guide for the Glide programming language: syntax, functions, loops, structs, concurrency, error handling, and testing."
    />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </>
);

/* ── TOC items ─────────────────────────────────────────────────────────── */
const TOC = [
  { id: "what-is-glide", label: "What is Glide?" },
  { id: "installation", label: "Installation" },
  { id: "first-program", label: "Your First Program" },
  { id: "syntax", label: "Syntax Fundamentals" },
  { id: "variables", label: "Variables & Constants" },
  { id: "functions", label: "Functions" },
  { id: "control-flow", label: "Control Flow" },
  { id: "loops", label: "Loops & Iteration" },
  { id: "structs", label: "Types & Structs" },
  { id: "errors", label: "Error Handling" },
  { id: "concurrency", label: "Concurrency" },
  { id: "threading", label: "Threading Macros" },
  { id: "imports", label: "Imports" },
  { id: "testing", label: "Testing" },
  { id: "cli", label: "CLI Reference" },
  { id: "operators", label: "Operator Reference" },
];

/* ── Small section title ───────────────────────────────────────────────── */
const H2 = ({ id, children }) => (
  <h2
    id={id}
    className="text-2xl font-bold text-gray-900 mt-12 mb-4 pb-2 border-b border-gray-200"
  >
    {children}
  </h2>
);

const H3 = ({ children }) => (
  <h3 className="text-lg font-semibold text-gray-900 mt-8 mb-3">{children}</h3>
);

const P = ({ children }) => (
  <p className="text-gray-700 leading-relaxed mb-4">{children}</p>
);

const Ul = ({ items }) => (
  <ul className="list-disc list-inside space-y-1 text-gray-700 mb-4">
    {items.map((item, i) => (
      <li key={i}>{item}</li>
    ))}
  </ul>
);

const InlineCode = ({ children }) => (
  <code className="font-mono text-glide-700 bg-glide-50 px-1.5 py-0.5 rounded text-sm">
    {children}
  </code>
);

const Note = ({ children }) => (
  <div className="my-4 p-4 rounded-lg bg-blue-50 border border-blue-200 text-blue-800 text-sm leading-relaxed">
    <strong>Note: </strong>
    {children}
  </div>
);

/* ── Operator table ────────────────────────────────────────────────────── */
const OPERATORS = [
  ["(+ a b)", "a + b", "Works on numbers and strings"],
  ["(- a b)", "a - b", ""],
  ["(* a b)", "a * b", ""],
  ["(/ a b)", "a / b", ""],
  ["(% a b)", "a % b", "Modulo"],
  ["(== a b)", "a == b", ""],
  ["(!= a b)", "a != b", ""],
  ["(< a b)", "a < b", ""],
  ["(<= a b)", "a <= b", ""],
  ["(> a b)", "a > b", ""],
  ["(>= a b)", "a >= b", ""],
  ["(&& a b)", "a && b", "Logical AND"],
  ["(|| a b)", "a || b", "Logical OR"],
  ["(! a)", "!a", "Logical NOT"],
  ["(& a b)", "a & b", "Bitwise AND / address-of"],
  ["(<- ch v)", "ch <- v", "Channel send"],
];

/* ── Main page ─────────────────────────────────────────────────────────── */
const GuidePage = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
        {/* Page header */}
        <div className="mb-8">
          <h1 className="text-4xl font-extrabold text-gray-900 mb-2">
            Developer Guide
          </h1>
          <p className="text-gray-500 text-lg">
            Everything you need to write, run, and test Glide programs.
          </p>
        </div>

        <div className="flex gap-10">
          {/* ── Sidebar ──────────────────────────────────────────────── */}
          <aside className="hidden lg:block w-56 flex-shrink-0">
            <nav className="sticky top-24 space-y-1" aria-label="Guide navigation">
              {TOC.map(({ id, label }) => (
                <a
                  key={id}
                  href={`#${id}`}
                  className="block py-1 px-3 rounded-md text-sm text-gray-600 hover:text-glide-700 hover:bg-glide-50 transition-colors"
                >
                  {label}
                </a>
              ))}
            </nav>
          </aside>

          {/* ── Mobile TOC toggle ────────────────────────────────────── */}
          <div className="lg:hidden mb-6 w-full">
            <button
              onClick={() => setSidebarOpen((v) => !v)}
              className="flex items-center gap-2 px-4 py-2 rounded-lg border border-gray-200 text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
            >
              <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" viewBox="0 0 24 24" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h8m-8 6h16" />
              </svg>
              On this page
            </button>
            {sidebarOpen && (
              <nav className="mt-2 p-4 bg-gray-50 rounded-xl border border-gray-200 grid grid-cols-2 gap-1">
                {TOC.map(({ id, label }) => (
                  <a
                    key={id}
                    href={`#${id}`}
                    onClick={() => setSidebarOpen(false)}
                    className="py-1 px-2 rounded text-sm text-gray-700 hover:text-glide-700 hover:bg-glide-50 transition-colors"
                  >
                    {label}
                  </a>
                ))}
              </nav>
            )}
          </div>

          {/* ── Content ──────────────────────────────────────────────── */}
          <article className="flex-1 min-w-0">

            {/* What is Glide */}
            <H2 id="what-is-glide">What is Glide?</H2>
            <P>
              Glide is a <strong>Lisp-style language that transpiles to idiomatic Go</strong>.
              It gives you:
            </P>
            <Ul
              items={[
                "Go's power and performance — every program is a real Go binary.",
                "Lisp's expressiveness — prefix S-expressions make precedence unambiguous.",
                "Readable indentation — significant whitespace replaces excess parentheses.",
                "First-class testing — test and assert are built into the language.",
              ]}
            />

            {/* Installation */}
            <H2 id="installation">Installation</H2>
            <P>Prerequisites: Go 1.21 or newer.</P>
            <CodeBlock
              language="bash"
              code={`git clone https://github.com/elielamora/glide
cd glide
go build -o glide .
# Move to PATH (optional)
sudo mv glide /usr/local/bin/`}
            />

            {/* First program */}
            <H2 id="first-program">Your First Program</H2>
            <P>
              Create <InlineCode>hello.glide</InlineCode>:
            </P>
            <CodeBlock
              code={`fn main ()
    fmt.Println "Hello, Glider!"`}
              filename="hello.glide"
            />
            <P>Run it:</P>
            <CodeBlock language="bash" code={`glide run hello.glide\n# → Hello, Glider!`} />
            <P>
              No <InlineCode>package main</InlineCode>, no{" "}
              <InlineCode>import "fmt"</InlineCode> — Glide handles both
              automatically.
            </P>

            {/* Syntax */}
            <H2 id="syntax">Syntax Fundamentals</H2>
            <H3>S-expressions</H3>
            <P>
              Every Glide expression is an S-expression: a head symbol followed
              by its arguments, either inline in parentheses or on indented lines
              below.
            </P>
            <CodeBlock
              code={`; inline
(+ 1 2)

; multi-line — indentation groups the call
+
    1
    2

; (1 + 2) * (3 + 4)
* (+ 1 2) (+ 3 4)`}
            />

            <H3>Comments</H3>
            <CodeBlock
              code={`; This is a comment
fn main ()
    fmt.Println "Hi" ; inline comment`}
            />

            <H3>Indentation (off-side rule)</H3>
            <P>
              Any line indented more than its parent becomes a child expression.
              Four spaces is the recommended indent.
            </P>
            <CodeBlock
              code={`; These two are identical:
(fmt.Printf "%d\\n" (* 6 7))

fmt.Printf "%d\\n"
    * 6 7`}
            />

            {/* Variables */}
            <H2 id="variables">Variables & Constants</H2>
            <H3>let — local variable binding</H3>
            <CodeBlock
              code={`fn main ()
    let x 42
    let greeting "hello"
    let pi 3.14159
    ; optional explicit type:
    let count int 0
    let ratio float64 0.5`}
            />

            <H3>set — reassignment</H3>
            <CodeBlock
              code={`fn main ()
    let counter 0
    set counter (+ counter 1)
    fmt.Println counter   ; 1`}
            />

            <H3>const — compile-time constants</H3>
            <CodeBlock
              code={`const MaxRetries 3
const AppName "glide-app"

fn main ()
    fmt.Println AppName MaxRetries`}
            />

            {/* Functions */}
            <H2 id="functions">Functions</H2>
            <H3>Basic function</H3>
            <P>
              Signature: <InlineCode>fn name (param type …) returnType</InlineCode>
            </P>
            <CodeBlock
              code={`fn add (a int b int) int
    + a b

fn main ()
    fmt.Println (add 3 4)   ; 7`}
            />

            <H3>Multiple return values</H3>
            <CodeBlock
              code={`fn divide (a float64 b float64) (float64 error)
    if (== b 0.0)
        (return 0.0 (errors.New "division by zero"))
        (return (/ a b) nil)

fn main ()
    let (result err) (divide 10.0 3.0)
    if (!= err nil)
        fmt.Println "Error:" err
        fmt.Printf "%.4f\\n" result`}
            />

            <H3>Methods on structs</H3>
            <CodeBlock
              code={`type Rect struct
    Width  float64
    Height float64

fn (r Rect) .Area () float64
    * r.Width r.Height

fn main ()
    let r (Rect 5.0 3.0)
    fmt.Println (.Area r)   ; 15`}
            />

            <H3>First-class functions and closures</H3>
            <CodeBlock
              code={`fn makeAdder (n int) func(int) int
    fn (x int) int
        + x n

fn main ()
    let addFive (makeAdder 5)
    fmt.Println (addFive 10)   ; 15`}
            />

            {/* Control flow */}
            <H2 id="control-flow">Control Flow</H2>
            <P>
              <InlineCode>if</InlineCode> is an expression in Glide — both
              branches return a value.
            </P>
            <CodeBlock
              code={`fn abs (n int) int
    if (< n 0) (- 0 n) n

fn main ()
    fmt.Println (abs -7)   ; 7
    fmt.Println (abs  3)   ; 3`}
            />
            <P>Side-effecting if with no else:</P>
            <CodeBlock
              code={`if (> x 5)
    fmt.Println "big"`}
            />

            {/* Loops */}
            <H2 id="loops">Loops & Iteration</H2>
            <H3>for — range loop</H3>
            <CodeBlock
              code={`for i in 1..=5    ; inclusive end
    fmt.Printf "i = %d\\n" i

for i in 0..5     ; exclusive end (0,1,2,3,4)
    fmt.Println i`}
            />

            <H3>each — slice iteration</H3>
            <CodeBlock
              code={`let fruits ["apple" "banana" "cherry"]
each [i fruit] in fruits
    fmt.Printf "%d: %s\\n" i fruit`}
            />

            <H3>loop / recur — tail-recursive loops</H3>
            <P>
              <InlineCode>loop</InlineCode> establishes named bindings;{" "}
              <InlineCode>recur</InlineCode> jumps back to the top with new
              values. Compiles to a Go <InlineCode>for</InlineCode> loop — no
              stack frames.
            </P>
            <CodeBlock
              code={`fn sumTo (n int) int
    loop (i n acc 0)
        if (<= i 0)
            acc
            recur (- i 1) (+ acc i)

fn main ()
    fmt.Println (sumTo 100)   ; 5050`}
            />

            {/* Structs */}
            <H2 id="structs">Types & Structs</H2>
            <CodeBlock
              code={`type Point struct
    X float64
    Y float64

fn main ()
    let p (Point 1.0 2.0)
    fmt.Printf "(%v, %v)\\n" p.X p.Y`}
            />

            <H3>Embedding and composition</H3>
            <CodeBlock
              code={`type Animal struct
    Name string

type Dog struct
    Animal
    Breed string

fn main ()
    let d (Dog (Animal "Rex") "Labrador")
    fmt.Println d.Name d.Breed`}
            />

            {/* Error handling */}
            <H2 id="errors">Error Handling</H2>
            <H3>try — propagate errors automatically</H3>
            <P>
              Similar to Rust's <InlineCode>?</InlineCode> operator.{" "}
              <InlineCode>(try expr)</InlineCode> evaluates expr and immediately
              returns the error from the enclosing function if it's non-nil.
            </P>
            <CodeBlock
              code={`fn readConfig (path string) (string error)
    let data (try (os.ReadFile path))
    (return (string data) nil)`}
            />

            <H3>guard — handle errors inline</H3>
            <CodeBlock
              code={`fn processFile (path string) error
    let (data err) (os.ReadFile path)
    guard err
        (return err)
    fmt.Println (string data)
    (return nil)`}
            />

            {/* Concurrency */}
            <H2 id="concurrency">Concurrency</H2>
            <H3>go — launch a goroutine</H3>
            <CodeBlock
              code={`fn main ()
    let wg (sync.WaitGroup{})
    for i in 0..5
        .Add &wg 1
        go
            .Done &wg
            fmt.Println "goroutine" i
    .Wait &wg`}
            />

            <H3>chan — create a channel</H3>
            <CodeBlock
              code={`fn producer (ch (chan int))
    for i in 0..5
        <- ch i

fn main ()
    let ch (chan int 10)
    go (producer ch)
    for v in ch
        fmt.Println v`}
            />

            {/* Threading macros */}
            <H2 id="threading">Threading Macros</H2>
            <P>
              Threading macros pipe a value through a series of
              transformations, making deeply nested calls readable.
            </P>
            <H3>{"-> (thread first)"}</H3>
            <CodeBlock
              code={`; Without threading — read inside-out:
fmt.Println (strings.ToUpper (strings.TrimSpace "  hello  "))

; With -> — read top to bottom:
-> "  hello  "
    strings.TrimSpace
    strings.ToUpper
    fmt.Println`}
            />

            <H3>{"->> (thread last)"}</H3>
            <P>
              Inserts the value as the <em>last</em> argument at each step —
              ideal for slice pipelines.
            </P>

            {/* Imports */}
            <H2 id="imports">Imports</H2>
            <H3>Auto-import</H3>
            <P>
              Standard-library packages used via qualified calls (
              <InlineCode>fmt.Println</InlineCode>,{" "}
              <InlineCode>math.Sqrt</InlineCode>, etc.) are imported
              automatically.
            </P>
            <H3>use — explicit import</H3>
            <CodeBlock
              code={`use "net/http"
use "net/http" web    ; import as alias

fn main ()
    let resp (try (web.Get "https://example.com"))
    fmt.Println resp.Status`}
            />

            {/* Testing */}
            <H2 id="testing">Testing</H2>
            <CodeBlock
              code={`fn hypotenuse (a float64 b float64) float64
    math.Sqrt (+ (* a a) (* b b))

test "3-4-5 triangle"
    assert (== (hypotenuse 3.0 4.0) 5.0)

test "isosceles right triangle"
    let h (hypotenuse 1.0 1.0)
    assert (> h 1.41)
    assert (< h 1.42)`}
              filename="geometry_test.glide"
            />
            <CodeBlock language="bash" code={`glide test geometry_test.glide`} />

            <H3>bench — micro-benchmarks</H3>
            <CodeBlock
              code={`bench "string concat"
    let _ (+ "hello" " " "world")`}
            />

            {/* CLI */}
            <H2 id="cli">CLI Reference</H2>
            <div className="overflow-x-auto my-4">
              <table className="w-full text-sm border-collapse">
                <thead>
                  <tr className="bg-gray-100">
                    <th className="text-left px-4 py-2 font-semibold text-gray-700 border border-gray-200">Command</th>
                    <th className="text-left px-4 py-2 font-semibold text-gray-700 border border-gray-200">Description</th>
                  </tr>
                </thead>
                <tbody>
                  {[
                    ["glide run file.glide", "Transpile and run"],
                    ["glide build file.glide", "Transpile and compile to a binary"],
                    ["glide transpile file.glide", "Print generated Go source to stdout"],
                    ["glide test file.glide", "Run test and bench blocks"],
                  ].map(([cmd, desc]) => (
                    <tr key={cmd} className="hover:bg-gray-50">
                      <td className="px-4 py-2 border border-gray-200 font-mono text-glide-700 whitespace-nowrap">{cmd}</td>
                      <td className="px-4 py-2 border border-gray-200 text-gray-600">{desc}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Operators */}
            <H2 id="operators">Operator Reference</H2>
            <div className="overflow-x-auto my-4">
              <table className="w-full text-sm border-collapse">
                <thead>
                  <tr className="bg-gray-100">
                    <th className="text-left px-4 py-2 font-semibold text-gray-700 border border-gray-200">Glide</th>
                    <th className="text-left px-4 py-2 font-semibold text-gray-700 border border-gray-200">Go</th>
                    <th className="text-left px-4 py-2 font-semibold text-gray-700 border border-gray-200">Notes</th>
                  </tr>
                </thead>
                <tbody>
                  {OPERATORS.map(([glide, go_, note]) => (
                    <tr key={glide} className="hover:bg-gray-50">
                      <td className="px-4 py-2 border border-gray-200 font-mono text-glide-700 whitespace-nowrap">{glide}</td>
                      <td className="px-4 py-2 border border-gray-200 font-mono text-indigo-700 whitespace-nowrap">{go_}</td>
                      <td className="px-4 py-2 border border-gray-200 text-gray-500">{note}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <Note>
              Happy gliding! 🚀 Check out the{" "}
              <a href="/examples" className="text-glide-700 underline hover:text-glide-900">
                Examples page
              </a>{" "}
              for more real-world code.
            </Note>
          </article>
        </div>
      </div>
    </Layout>
  );
};

export default GuidePage;
