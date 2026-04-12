# Note Service

Go microservice for notes, implemented close to the reference repository:

- `handler -> service -> storage`
- MongoDB as primary storage
- custom `internal/apperror` middleware
- CRUD for notes
- partial update with special handling for `tags`
