# Food Ordering DDD - Claude Code Project Guide

## Project Overview

Food delivery platform implementing **DDD**, **Sagas**, **Outbox Pattern**, and **Event-Driven Architecture** in Go.

**Study project** — the goal is learning distributed architecture patterns, not shipping to production.

## Architecture

- **Go Workspace**: `go.work` with modules `./catalog` and `./common`
- **Go Version**: 1.25.5
- **Database**: MySQL (one instance per bounded context)
- **Communication**: Hybrid — gRPC for sync validation, Event Bus for async lifecycle
- **Saga Pattern**: Parallel Saga (Orchestrated) — Ordering is the orchestrator
- **Consistency**: Eventual, with compensation flows
- **Events**: Fat Events (Event-Carried State Transfer)

## Bounded Contexts

| Context        | Module Dir     | Status         |
|----------------|----------------|----------------|
| Catalog        | `catalog/`     | In Progress    |
| Ordering       | (not created)  | Not Started    |
| Payment        | (not created)  | Not Started    |
| Restaurant     | (not created)  | Not Started    |
| Delivery       | (not created)  | Not Started    |

## Key Documentation

- `docs/PROJECT_OVERVIEW.md` — Domain model, aggregates, events, flows
- `docs/ARCHITECTURE.md` — Technical decisions, saga analysis, context mapping
- `docs/EVENTS.md` — Event definitions
- `docs/SystemDesign.png` — System topology
- `catalog/TODO.md` — Catalog implementation roadmap
- `catalog/IDEAS.md` — Architectural trade-offs (repository patterns)

## Code Structure (per Bounded Context)

Each bounded context follows **Hexagonal Architecture**:

```
<context>/
├── cmd/main.go
├── internal/
│   ├── core/
│   │   ├── domain/        # Aggregates, entities, VOs, domain events, domain services
│   │   ├── app/           # Application services (use cases), mappers
│   │   └── ports/         # Input ports (DTOs) and output ports (repository interfaces)
│   ├── adapters/
│   │   ├── input/
│   │   │   ├── rest/      # HTTP handlers (std lib net/http)
│   │   │   └── grpc/      # gRPC server + proto definitions
│   │   └── output/
│   │       └── repository/ # MySQL implementations, DAOs, Unit of Work
│   └── infra/
│       ├── config/        # Env-based configuration
│       ├── db/            # MySQL connection + migrations
│       └── server/        # HTTP/gRPC server lifecycle
```

## Conventions

### Go
- **Standard library HTTP** (net/http with ServeMux) — no frameworks
- **Structured logging** with `slog` (JSON format)
- **Testcontainers** for integration tests
- **go-playground/validator** for input validation
- **Surrogate keys** (auto-increment) + **UUID business IDs**
- **Money** stored as cents (int64)

### DDD Patterns
- Aggregates protect invariants via domain methods (Rich Domain Model)
- Value Objects are immutable
- Domain events are raised inside aggregates and collected
- Application services coordinate use cases, depend on port interfaces
- **Unit of Work** pattern for transactional consistency
- **Outbox Pattern**: events written atomically with aggregate in `outbox_events` table

### Testing
- Domain/service unit tests with testify
- Repository integration tests with Testcontainers + MySQL
- Target: 95%+ coverage on domain layer

### Build & Run
```bash
cd catalog && make build        # Compile
cd catalog && make run          # Build + run
cd catalog && make test         # Run all tests
cd catalog && make migrate-up   # Apply migrations
cd catalog && make proto        # Regenerate gRPC code
```

## Known Technical Debt (Catalog)

See `catalog/TODO.md` for full list. Key items:
- Anti-corruption layer: adapters import domain types directly
- Validation logic leaking from domain to service layer
- Missing handler-level tests
- Outbox worker/publisher not yet implemented

## DDD Mentor Mode

When discussing architecture (not writing code), follow the workflow in `.agent/workflows/ddd-mentor.md`:
1. Validate against documentation before any implementation
2. Ask Socratic questions about modeling decisions
3. Generate skeleton-first (interfaces + structs), not full implementations
4. Review for Rich Domain Model, immutability, side effects, idempotency
