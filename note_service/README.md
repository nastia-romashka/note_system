# Note Service

Go microservice for notes, implemented close to the reference repository:

- application code lives in `app/`
- runtime files for local testing stay at service root: `config.yml`, `app_test.http`, `Dockerfile`
- `handler -> service -> storage`
- MongoDB as primary storage
- custom `internal/apperror` middleware
- CRUD for notes
- partial update with special handling for `tags`
