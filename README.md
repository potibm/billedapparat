# Billedapparat

![billedapparat logo](doc/billedapparat.svg)

> _Billedapparat_ is the Danish term for imaging device.

It is a party information system for the bigscreen at [demoparties](https://en.wikipedia.org/wiki/Demoscene#Parties). Based on earlier solutions (including [Partymeister](https://github.com/partymeister) and Tagwall) for [Evoke](https://www.evoke.eu/).

## Tooling

- [Go](https://go.dev)
  - [Gin Web Framework](https://gin-gonic.com)
  - [GORM](https://gorm.io)
  - [Cobra](https://cobra.dev) & [Viper](https://github.com/spf13/viper)
- [React](https://react.dev)
  - [Vite](https://vitejs.dev/)
  - [React Admin](https://marmelab.com/react-admin/)
  - [Flowbite React](https://flowbite-react.com) & [Tailwind CSS](https://tailwindcss.com)
- [SQLite](https://www.sqlite.org)
- Observability
  - [Sentry](https://sentry.io)
  - [OpenTelemetry](https://opentelemetry.io)
- Development & Ops
  - [mise](https://mise.jdx.dev/)
  - [Docker](https://www.docker.com)

## Quickstart

We use `mise` to automatically manage all tool versions (Go, Node, etc.) and project tasks.

```bash
# 1. Install mise (if not already installed)
curl https://mise.run | sh

# 2. Setup the project (installs dependencies and starts infra)
mise run setup

# 3. Start the development server (hot-reload for backend & frontend)
overmind s --timeout 10
```

## Documentation

_todo_
