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

Please refer to `mise` documentation on how to install it.

```
# 1. Setup local infrastructure
# This generates local certificates, configures dnsmasq/resolver, and updates /etc/hosts (Linux)
mise run infra:prepare

# 2. Start local services (Traefik, Dex, etc.)
mise run infra:up

# 3. Start the development server (hot-reload for backend & frontend)
overmind s --timeout 10
```

## Local Environment

Once the stack is running, you can access the applications via:

- Billedapparat Admin: https://billedapparat.test
- Dex IdP: https://dex.billedapparat.test
- OpenObserve: https://observe.billedapparat.test
- RedisInsight: https://redis.billedapparat.test

## Authentication (OIDC)

The system is configured to use OpenID Connect (OIDC) via [Dex](https://dexidp.io/). Local environment uses `react-admin-client` as client id to connect with `dex`. Make sure `skip_tls_verify` is not true in production.

To log in:
- Username: `admin@example.com`
- Password: `password`

## Documentation

Please refer to the documentation for further details.
