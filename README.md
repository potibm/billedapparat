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
- Dex IdP: https://dex.billedapparat.test

## Authentication (OIDC)

Billedapparat supports OpenID Connect (OIDC) authentication for the admin panel. When enabled, all `/api/admin` endpoints are protected by Bearer token validation.

### How it works

1. The backend exposes its OIDC configuration via the `/api/config` endpoint
2. The React Admin frontend reads this config and initializes the `oidc-client-ts` UserManager
3. Users are redirected to the configured Identity Provider (e.g., Dex) to log in
4. After successful authentication, the frontend receives an access token
5. The access token is attached to every API request via the `Authorization: Bearer <token>` header
6. The backend verifies the token against the OIDC provider on every admin request

### Configuration

Authentication is configured in `backend/config/config.yaml` (or `config.local.yaml`):

```yaml
auth:
  type: oidc
  name: Dex # Display name shown on the login button
  authority: https://dex.billedapparat.test/dex
  client_id: react-admin-client
  skip_tls_verify: false # Only set to true in local development
```

You can also override these values via environment variables:

| Variable                   | Description                                 | Example                              |
| -------------------------- | ------------------------------------------- | ------------------------------------ |
| `APP_AUTH_TYPE`            | Authentication type (only `oidc` supported) | `oidc`                               |
| `APP_AUTH_NAME`            | Display name of the Identity Provider       | `Dex`                                |
| `APP_AUTH_AUTHORITY`       | OIDC issuer URL                             | `https://dex.billedapparat.test/dex` |
| `APP_AUTH_CLIENT_ID`       | OIDC client ID                              | `react-admin-client`                 |
| `APP_AUTH_SKIP_TLS_VERIFY` | Skip TLS verification (dev only!)           | `false`                              |

> **Security Warning:** Never set `skip_tls_verify` to `true` in production. The application will refuse to start in production mode if this flag is enabled.

### Local Development with Dex

The included `docker-compose.yaml` provides a local [Dex](https://dexidp.io/) instance pre-configured with a static user:

- **Email:** `admin@example.com`
- **Password:** `password`

The Dex configuration is located at `infra/dex-config.yaml`. You can modify the static users, clients, or connect Dex to an external identity provider (LDAP, GitHub, Google, etc.) by editing this file.

### Audit Trail

When OIDC authentication is active, the backend automatically tracks the user ID on all database operations:

- `created_by` — set on insert
- `modified_by` — updated on every change
- `deleted_by` — set on soft delete

These fields are stored on all GORM models (slides, news, timetable events, filter rules).

## Documentation

Please refer to the [collectors](doc/collectors.md) documentation. Further information will be added soon.
