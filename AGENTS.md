# FlySoft Flight Integration Service

This repository is a Go 1.25 backend microservice. The locked stack is Gin, GORM, PostgreSQL, golang-migrate, `log/slog`, and swaggo. Swagger dependencies and generated docs are added in Phase 6, not earlier.

`docs/PLAN.md` is the source of truth for architecture, API contracts, phase scope, and acceptance gates.

Layering rule: HTTP handlers only bind requests, call services, and map responses or errors. Business logic belongs in `internal/services`. Dependencies flow `handlers -> services -> {repository, providers}`. Leaf packages such as `domain`, `money`, `pricing`, and `apperror` should stay dependency-light.

Money rule: represent money as `int64` cents everywhere. Do not use floating point for money math. Float formatting is allowed only at the JSON display boundary defined by the money package in a later phase.

Errors use one unified envelope:

```json
{"success":false,"error":{"code":"...","message":"..."}}
```

Success responses use `{"success":true,"data":...}` or `{"success":true,"booking":...}` depending on the endpoint.
