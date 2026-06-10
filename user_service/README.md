# User Service

Go microservice for users. It exposes the same HTTP API as before, but stores users in PostgreSQL.

- application code: `app/`
- database migrations: `migrations/`
- local runtime/config files: `config.yml`, `Dockerfile`
- main package: `app/cmd/main`
- layers: `app/internal/handlers/users`, `app/internal/service/users`, `app/internal/storage/db`
- Swagger UI: `/swagger`
- OpenAPI schema: `/openapi.json`
