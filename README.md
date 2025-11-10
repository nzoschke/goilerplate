# goilerplate

Production-ready Go boilerplate with templ, HTMX, and TailwindCSS.

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Make](https://www.gnu.org/software/make/)
- [TailwindCSS CLI](https://tailwindcss.com/blog/standalone-cli)
- [Docker](https://docs.docker.com/get-docker/)

> **Note:** Development tools like `templ` and `air` are automatically managed via Go 1.25's `tool` directive and don't need manual installation.

## Quick Start

```bash
# Clone
git clone https://github.com/templui/goilerplate.git
cd goilerplate

# Setup environment
cp .env.example .env

# Start (installs dependencies, generates code, starts server)
make dev
```

Open [http://localhost:7331](http://localhost:7331)

## Documentation

[Docs](https://goilerplate.com/docs)

## License

[License](LICENSE)
