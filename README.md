# goilerplate

Production-ready Go boilerplate with templ, HTMX, and TailwindCSS.

## Prerequisites

- [Go 1.24+](https://go.dev/dl/)
- [Make](https://www.gnu.org/software/make/)
- [templ](https://templ.guide/quick-start/installation)
- [TailwindCSS CLI](https://tailwindcss.com/blog/standalone-cli)
- [air](https://github.com/air-verse/air)
- [Docker](https://docs.docker.com/get-docker/)

## Quick Start

```bash
# Clone and install
git clone https://github.com/templui/goilerplate.git
cd goilerplate
go mod download

# Setup environment
cp .env.example .env

# Start development server
make dev
```

Open [http://localhost:7331](http://localhost:7331)

## Documentation

[Docs](https://goilerplate.com/docs)

## License

[License](LICENSE)
