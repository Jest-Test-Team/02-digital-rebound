# Source — Digital Rebound

Phase 1 MVP vertical slice: `ingest → validate → calculate → explain → review`.

## Run locally

```bash
# from repository root
make run
# or
cd src && SCHEMA_PATH=../schemas/event.schema.json go run ./cmd/server
```

API listens on `:8081` by default (`ADDR` env).

## Key packages

| Package | Responsibility |
| --- | --- |
| `internal/dto` | Request binding / validation structs |
| `internal/vo` | Response structs (OpenAPI-aligned) |
| `internal/rules` | Pinned explainable rebound metrics (`rebound-rules@0.1.0`) |
| `internal/store` | File-based append-only evidence + assessments |
| `internal/service` | Schema validation, analysis, review annotations |
| `internal/handler` | Gin HTTP handlers |

Do not introduce ML or production-data connectors in Phase 1.
