# Billedapparat: Authentication (OIDC)

Billedapparat supports OpenID Connect (OIDC) authentication for the admin panel. When enabled, all `/api/admin` endpoints are protected by Bearer token validation.

## How it works

1. The backend exposes its OIDC configuration via the `/api/config` endpoint
2. The React Admin frontend reads this config and initializes the `oidc-client-ts` UserManager
3. Users are redirected to the configured Identity Provider (e.g., Dex) to log in
4. After successful authentication, the frontend receives an access token
5. The access token is attached to every API request via the `Authorization: Bearer <token>` header
6. The backend verifies the token against the OIDC provider on every admin request

## Configuration

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

## Local Development with Dex

The included `docker-compose.yaml` provides a local [Dex](https://dexidp.io/) instance pre-configured with a static user:

- **Email:** `admin@example.com`
- **Password:** `password`

The Dex configuration is located at `infra/dex-config.yaml`. You can modify the static users, clients, or connect Dex to an external identity provider (LDAP, GitHub, Google, etc.) by editing this file.

## Audit Trail

When OIDC authentication is active, the backend automatically tracks the user ID on all database operations:

- `created_by` — set on insert
- `modified_by` — updated on every change
- `deleted_by` — set on soft delete

These fields are stored on all GORM models (slides, news, timetable events, filter rules).
