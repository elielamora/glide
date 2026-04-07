import React, { useState } from "react";
import { Link } from "gatsby";
import Layout from "../components/Layout";
import CodeBlock from "../components/CodeBlock";

export const Head = () => (
  <>
    <title>Glide — The Lisp that speaks Go</title>
    <meta
      name="description"
      content="Glide is a readable Lisp that compiles to idiomatic Go. Get Go's performance and concurrency with Lisp's expressive power — no boilerplate, built-in testing."
    />
    <meta property="og:title" content="Glide — The Lisp that speaks Go" />
    <meta
      property="og:description"
      content="A modern Lisp that compiles to idiomatic Go. Express more. Write less."
    />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </>
);

const HERO_CODE = `; Factorial in Glide — compiles to a Go for loop
fn factorial (n int) int
    loop (current n acc 1)
        if (<= current 1)
            acc
            recur (- current 1) (* acc current)

fn main ()
    for i in 1..=10
        fmt.Printf "%d! = %d\\n" i (factorial i)`;

const FEATURES = [
  {
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="w-7 h-7" aria-hidden="true">
        <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
      </svg>
    ),
    title: "Go Performance",
    description:
      "Every Glide program compiles to clean, gofmt-formatted Go source and runs as a native binary. No interpreter, no VM, no runtime penalty.",
  },
  {
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="w-7 h-7" aria-hidden="true">
        <path strokeLinecap="round" strokeLinejoin="round" d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
      </svg>
    ),
    title: "Expressive Syntax",
    description:
      "Prefix S-expressions eliminate operator precedence bugs. Significant indentation replaces noise parentheses. Write what you mean.",
  },
  {
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="w-7 h-7" aria-hidden="true">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
    ),
    title: "Built-in Testing",
    description:
      "test, assert, and bench are first-class language constructs — not a framework, not a library. Just write tests and run glide test.",
  },
  {
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="w-7 h-7" aria-hidden="true">
        <path strokeLinecap="round" strokeLinejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
      </svg>
    ),
    title: "Tail-Call Loops",
    description:
      "loop/recur compiles to a plain Go for loop. Write recursive-style logic with zero stack overhead — no overflow, no trampoline.",
  },
  {
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="w-7 h-7" aria-hidden="true">
        <path strokeLinecap="round" strokeLinejoin="round" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
      </svg>
    ),
    title: "Auto Imports",
    description:
      "Use fmt.Println, math.Sqrt, os.ReadFile — common standard library packages are imported automatically. Focus on your code.",
  },
  {
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="w-7 h-7" aria-hidden="true">
        <path strokeLinecap="round" strokeLinejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0" />
      </svg>
    ),
    title: "Go Concurrency",
    description:
      "Goroutines with go expr. Channels with chan. The full power of Go's concurrency model expressed in two words.",
  },
];

const EXAMPLES = [
  {
    label: "Hello World",
    code: `fn main ()
    fmt.Println "Hello, Glider! 🚀"`,
    note: "No package main. No imports. Glide handles the ceremony.",
  },
  {
    label: "Fibonacci",
    code: `fn fib (n int) int
    if (< n 2)
        n
        + (fib (- n 1)) (fib (- n 2))

fn main ()
    for i in 0..=10
        fmt.Printf "fib(%d) = %d\\n" i (fib i)`,
    note: "if is an expression — both branches return a value.",
  },
  {
    label: "Structs",
    code: `type Point struct
    X float64
    Y float64

fn distance (a Point b Point) float64
    let dx (- b.X a.X)
    let dy (- b.Y a.Y)
    math.Sqrt (+ (* dx dx) (* dy dy))

fn main ()
    let p (Point 3.0 4.0)
    fmt.Printf "%.1f\\n" (distance (Point 0 0) p)`,
    note: "Struct literals need no field names — positional by default.",
  },
  {
    label: "Threading",
    code: `; Without threading — read inside-out:
fmt.Println (strings.ToUpper (strings.TrimSpace "  hello  "))

; With -> — read top to bottom:
-> "  hello  "
    strings.TrimSpace
    strings.ToUpper
    fmt.Println`,
    note: "-> threads a value as the first argument through each step.",
  },
  {
    label: "Concurrency",
    code: `fn worker (id int results (chan string))
    <- results (fmt.Sprintf "worker %d done" id)

fn main ()
    let results (chan string 5)
    for i in 0..5
        go (worker i results)
    for i in 0..5
        fmt.Println (<- results)`,
    note: "go launches a goroutine. <- sends to / receives from a channel.",
  },
  {
    label: "Testing",
    code: `fn hypotenuse (a float64 b float64) float64
    math.Sqrt (+ (* a a) (* b b))

test "3-4-5 triangle"
    assert (== (hypotenuse 3.0 4.0) 5.0)

test "isosceles right triangle"
    let h (hypotenuse 1.0 1.0)
    assert (> h 1.41)
    assert (< h 1.42)`,
    note: "test and assert are keywords. Run with: glide test file.glide",
  },
];

const INSTALL_STEPS = [
  {
    step: "01",
    title: "Clone & Build",
    code: "git clone https://github.com/elielamora/glide\ncd glide && go build -o glide .",
  },
  {
    step: "02",
    title: "Write Glide Code",
    code: `fn main ()\n    fmt.Println "Hello, Glider!"`,
  },
  {
    step: "03",
    title: "Run It",
    code: "glide run hello.glide",
  },
];

const IndexPage = () => {
  const [activeExample, setActiveExample] = useState(0);

  return (
    <Layout>
      {/* ── Hero ──────────────────────────────────────────────────────── */}
      <section className="relative overflow-hidden bg-gradient-to-br from-glide-950 via-glide-900 to-indigo-900 text-white">
        <div
          className="absolute inset-0 opacity-10"
          style={{
            backgroundImage:
              "radial-gradient(circle at 20% 50%, #7c3aed 0%, transparent 50%), radial-gradient(circle at 80% 20%, #4f46e5 0%, transparent 40%)",
          }}
          aria-hidden="true"
        />
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20 lg:py-28">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            <div>
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-glide-700/50 border border-glide-500/40 text-glide-200 text-xs font-semibold mb-6 uppercase tracking-wider">
                <span className="w-2 h-2 rounded-full bg-glide-400 animate-pulse" />
                Open Source · MIT License
              </div>
              <h1 className="text-5xl lg:text-6xl font-extrabold leading-tight mb-6">
                The Lisp
                <br />
                <span className="text-transparent bg-clip-text bg-gradient-to-r from-glide-300 to-indigo-300">
                  that speaks Go.
                </span>
              </h1>
              <p className="text-lg text-glide-100/80 leading-relaxed mb-8 max-w-lg">
                Glide is a modern Lisp that transpiles to idiomatic Go. Get{" "}
                <strong className="text-white">Go's performance and concurrency</strong>{" "}
                with <strong className="text-white">Lisp's expressive power</strong> — zero
                boilerplate, first-class testing, and clean readable syntax.
              </p>
              <div className="flex flex-wrap gap-4">
                <Link
                  to="/guide"
                  className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-white text-glide-700 font-bold text-sm hover:bg-glide-50 transition-all shadow-lg shadow-black/20"
                >
                  Get Started
                  <svg width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24" aria-hidden="true">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M17 8l4 4m0 0l-4 4m4-4H3" />
                  </svg>
                </Link>
                <Link
                  to="/examples"
                  className="inline-flex items-center gap-2 px-6 py-3 rounded-xl border border-glide-400/50 text-glide-100 font-bold text-sm hover:bg-white/10 transition-all"
                >
                  View Examples
                </Link>
              </div>
            </div>

            <div>
              <CodeBlock code={HERO_CODE} filename="factorial.glide" />
            </div>
          </div>
        </div>
      </section>

      {/* ── Features ──────────────────────────────────────────────────── */}
      <section className="py-20 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-14">
            <h2 className="text-3xl sm:text-4xl font-extrabold text-gray-900 mb-4">
              Why Glide?
            </h2>
            <p className="text-lg text-gray-500 max-w-2xl mx-auto">
              All the expressiveness of Lisp. All the pragmatism of Go. None of
              the boilerplate of either.
            </p>
          </div>
          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {FEATURES.map(({ icon, title, description }) => (
              <div
                key={title}
                className="group p-6 rounded-2xl border border-gray-100 bg-gray-50 hover:border-glide-200 hover:bg-glide-50 transition-all"
              >
                <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-glide-100 text-glide-700 mb-4 group-hover:bg-glide-200 transition-colors">
                  {icon}
                </div>
                <h3 className="text-lg font-bold text-gray-900 mb-2">{title}</h3>
                <p className="text-gray-600 text-sm leading-relaxed">{description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Interactive Examples ──────────────────────────────────────── */}
      <section className="py-20 bg-gray-950">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-10">
            <h2 className="text-3xl sm:text-4xl font-extrabold text-white mb-4">
              See Glide in Action
            </h2>
            <p className="text-gray-400 max-w-xl mx-auto">
              Expressive, readable, and powerful. Browse real examples.
            </p>
          </div>

          {/* Tabs */}
          <div className="flex flex-wrap gap-2 justify-center mb-8">
            {EXAMPLES.map(({ label }, i) => (
              <button
                key={label}
                onClick={() => setActiveExample(i)}
                className={`px-4 py-2 rounded-full text-sm font-medium transition-all ${
                  activeExample === i
                    ? "bg-glide-600 text-white shadow-lg shadow-glide-600/30"
                    : "bg-gray-800 text-gray-400 hover:bg-gray-700 hover:text-gray-200"
                }`}
              >
                {label}
              </button>
            ))}
          </div>

          {/* Code panel */}
          <div className="max-w-3xl mx-auto">
            <CodeBlock
              code={EXAMPLES[activeExample].code}
              filename={`${EXAMPLES[activeExample].label.toLowerCase().replace(/\s+/g, "-")}.glide`}
            />
            <p className="mt-4 text-center text-sm text-gray-400 italic">
              {EXAMPLES[activeExample].note}
            </p>
          </div>
        </div>
      </section>

      {/* ── Transpile preview ─────────────────────────────────────────── */}
      <section className="py-20 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h2 className="text-3xl sm:text-4xl font-extrabold text-gray-900 mb-4">
              Clean Go Output
            </h2>
            <p className="text-gray-500 max-w-2xl mx-auto">
              Glide transpiles to code you'd be proud to have written yourself.
              Run <code className="font-mono text-glide-700 bg-glide-50 px-1.5 py-0.5 rounded">glide transpile</code> to
              inspect the generated Go at any time.
            </p>
          </div>
          <div className="grid md:grid-cols-2 gap-6 max-w-5xl mx-auto">
            <div>
              <p className="text-xs font-semibold uppercase tracking-wider text-gray-400 mb-3">
                Glide source
              </p>
              <CodeBlock
                code={`fn factorial (n int) int
    loop (current n acc 1)
        if (<= current 1)
            acc
            recur (- current 1) (* acc current)`}
                filename="factorial.glide"
              />
            </div>
            <div>
              <p className="text-xs font-semibold uppercase tracking-wider text-gray-400 mb-3">
                Generated Go
              </p>
              <CodeBlock
                language="go"
                code={`func factorial(n int) int {
    current, acc := n, 1
    for {
        if current <= 1 {
            return acc
        }
        current, acc = current-1, acc*current
    }
}`}
                filename="factorial.go"
              />
            </div>
          </div>
        </div>
      </section>

      {/* ── Install ───────────────────────────────────────────────────── */}
      <section className="py-20 bg-gray-50" id="install">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-14">
            <h2 className="text-3xl sm:text-4xl font-extrabold text-gray-900 mb-4">
              Up and running in seconds
            </h2>
            <p className="text-gray-500">
              You need Go 1.21+ installed. That's it.
            </p>
          </div>
          <div className="space-y-6">
            {INSTALL_STEPS.map(({ step, title, code }) => (
              <div key={step} className="flex gap-6 items-start">
                <div className="flex-shrink-0 w-10 h-10 rounded-full bg-glide-600 text-white flex items-center justify-center font-bold text-sm">
                  {step}
                </div>
                <div className="flex-1">
                  <p className="font-semibold text-gray-900 mb-2">{title}</p>
                  <CodeBlock language="bash" code={code} />
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── CTA ───────────────────────────────────────────────────────── */}
      <section className="py-20 bg-gradient-to-br from-glide-700 to-indigo-700 text-white">
        <div className="max-w-3xl mx-auto px-4 text-center">
          <h2 className="text-4xl font-extrabold mb-4">
            Ready to glide?
          </h2>
          <p className="text-glide-100 text-lg mb-8">
            Start with the developer guide or dive straight into examples.
          </p>
          <div className="flex flex-wrap justify-center gap-4">
            <Link
              to="/guide"
              className="px-8 py-3 rounded-xl bg-white text-glide-700 font-bold hover:bg-glide-50 transition-all shadow-xl"
            >
              Read the Guide →
            </Link>
            <Link
              to="/examples"
              className="px-8 py-3 rounded-xl border border-white/40 text-white font-bold hover:bg-white/10 transition-all"
            >
              Browse Examples
            </Link>
          </div>
        </div>
      </section>
    </Layout>
  );
};

export default IndexPage;
