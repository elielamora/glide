# glide
The go list interpreter development experience

## Documentation

| Document | Description |
|----------|-------------|
| [Demo & Showcase](docs/demo.md) | Compelling examples and a language tour for new gliders |
| [Developer Guide](docs/guide.md) | Complete reference: syntax, types, loops, concurrency, testing, and more |

## Website

A full showcase website lives in [`website/`](website/). Built with Gatsby + Tailwind CSS, it includes:

- **Homepage** — marketing landing page with interactive code examples
- **[/guide](website/src/pages/guide.js)** — developer guide with sidebar navigation
- **[/examples](website/src/pages/examples.js)** — runnable examples for every language feature

```sh
cd website
npm install
npm run develop    # dev server on http://localhost:8000
npm run build      # static output in website/public/
```

## Quick Start

```sh
git clone https://github.com/elielamora/glide
cd glide
go build -o glide .
glide run examples/hello.glide
```

## CLI

```
glide run       <file.glide>   Transpile and run
glide build     <file.glide>   Transpile and compile to binary
glide transpile <file.glide>   Print generated Go source to stdout
glide test      <file.glide>   Run test blocks
```
