# Billedapparat

![billedapparat logo](doc/billedapparat.svg)

> _Billedapparat_ is the Danish term for imaging device.

It is a party information system for the bigscreen at [demoparties](https://en.wikipedia.org/wiki/Demoscene#Parties). Based on earlier solutions (including [Partymeister](https://github.com/partymeister) and Tagwall) for [Evoke](https://www.evoke.eu/).

## Tooling

- **Backend**
  - [Go](https://go.dev)
    - [Gin Web Framework](https://gin-gonic.com)
    - [GORM](https://gorm.io)
    - [Cobra](https://cobra.dev) & [Viper](https://github.com/spf13/viper)
- **Frontend**
  - [React](https://react.dev)
    - [Vite](https://vitejs.dev/)
    - [React Admin](https://marmelab.com/react-admin/)
    - [Flowbite React](https://flowbite-react.com) & [Tailwind CSS](https://tailwindcss.com)
- **Database**
  - [SQLite](https://www.sqlite.org)
- **Identity & Local Infrastructure**
  - [Dex](https://dexidp.io/) (OIDC Provider)
  - [Traefik](https://traefik.io/) (Local Edge Router)
  - [mkcert](https://github.com/FiloSottile/mkcert) & dnsmasq (Local TLS & `.test` Routing)
- **Observability**
  - [Sentry](https://sentry.io)
  - [OpenTelemetry](https://opentelemetry.io)
- **Development & Ops**
  - [mise](https://mise.jdx.dev/)
  - [Docker](https://www.docker.com)

## Quickstart

We use `mise` to automatically manage all tool versions (Go, Node, etc.) and project tasks.

```bash
# 1. Install mise (if not already installed)
curl https://mise.run | sh

# 2. Setup local infrastructure
# This generates local certificates, configures dnsmasq/resolver, and updates /etc/hosts (Linux)
mise run infra:prepare

# 3. Start local services (Traefik, Dex, etc.)
mise run infra:up

# 4. Start the development server (hot-reload for backend & frontend)
overmind s --timeout 10
```

## Local Environment

Once the stack is running, you can access the applications via:

- Billedapparat Admin: https://billedapparat.test

## Documentation

Please refer to the documentation on

* [collectors](doc/collectors.md)
* [authentication](doc/auth.md)
* [setting up a global traefik service](doc/global-traefik.md)

Further information will be added soon.
