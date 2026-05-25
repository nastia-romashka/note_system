# Project Structure

The repository now separates service code from delivery files.

## Conventions

- each service keeps application code inside `app/`
- service root keeps runtime and infra files such as `Dockerfile`, `config.yml|config.yaml`, `.env`, and `*.http`
- the whole stack is orchestrated by the root `docker-compose.yml`
- the shared deployment playbook lives in `deploy/deploy-service.yml`
- CI/CD lives in `.github/workflows/publish.yml`

## Examples

- `api_service/app` contains Go source code and `go.mod`
- `api_service/config.yml` stays available for local runs and `test.http`
- `note_service/app_test.http` still targets the note service directly
- `web_client/app` contains the Vite app, while `web_client/Dockerfile` builds the production image
