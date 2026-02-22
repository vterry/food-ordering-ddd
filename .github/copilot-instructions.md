# Food Ordering DDD - AI Coding Agent Instructions

## Project Context

This is a **DDD study project** for a food delivery platform implementing **Domain-Driven Design**, **Saga Pattern**, **Outbox Pattern**, and **Event-Driven Architecture** in Go. The goal is learning distributed architecture patterns, not production deployment.

## Architecture Overview

### Bounded Contexts & Structure

- **Go Workspace**: `go.work` with modules `./catalog` (in progress) and `./common`
- **Database**: MySQL (one instance per bounded context)
- **Communication**: Hybrid — gRPC for sync validation, Event Bus for async lifecycle
- **Saga Pattern**: Parallel Saga (Orchestrated) — Ordering context will orchestrate flows
- **Consistency**: Eventual consistency with compensation flows

Each bounded context follows **Hexagonal Architecture**:

```
<context>/
├── cmd/main.go
├── internal/
│   ├── core/
│   │   ├── domain/           # Aggregates, entities, VOs, domain events, domain services
│   │   ├── app/              # Application services (use cases)
│   │   └── ports/            # Input (DTOs) & output (repository interfaces)
│   ├── adapters/
│   │   ├── input/            # REST (net/http) & gRPC handlers
│   │   └── output/           # Repository implementations, DAOs
│   └── infra/
│       ├── config/, db/, server/
```

## Key Technical Patterns

### Domain-Driven Design

- **Rich Domain Model**: Aggregates enforce invariants via domain methods, not anemic setters
- **Value Objects**: Immutable (e.g., `Money`, `RestaurantID`, `MenuID`)
- **Domain Events**: Raised inside aggregates via `AddEvent()`, collected by app services
- **Aggregate Root**: Single entry point for modifications (e.g., `Menu`, `Restaurant`)

Example from [catalog/internal/core/domain/menu/menu.go](catalog/internal/core/domain/menu/menu.go#L55-L63):

```go
func (m *Menu) Activate() error {
    if !m.hasItems() {
        return ErrCannotActivateEmpty
    }
    m.status = enums.MenuActive
    event := NewMenuActivated(*m)
    m.AddEvent(event)
    return nil
}
```

### Unit of Work Pattern

Transactional consistency across aggregate + outbox events:

```go
err := m.uow.Run(ctx, func(ctxTx context.Context) error {
    return m.menuRepository.Save(ctxTx, menuAggregate)
})
```

- Transaction context passed via `context.Value(txKey{})`
- Repository methods use `getExecutor()` to auto-detect transaction
- See [catalog/internal/adapters/output/repository/unit_of_work.go](catalog/internal/adapters/output/repository/unit_of_work.go)

### Outbox Pattern

- Events persisted atomically with aggregate in `outbox_events` table
- Schema: `(aggregate_id, event_type, payload, status, created_at)`
- **Not yet implemented**: Worker to poll and publish events to event bus

### Money Value Object

- Stored as **cents (int64)** to avoid floating-point precision issues
- Created via `NewMoneyFromFloat(12.50)` or `NewMoneyFromCents(1250)`
- See [common/pkg/money.go](common/pkg/money.go)

### ID Strategy

- **Surrogate keys**: Auto-increment primary keys in database
- **Business IDs**: UUIDs for domain identity (e.g., `MenuID`, `RestaurantID`)
- Base implementation in [common/pkg/base_id.go](common/pkg/base_id.go)

## Development Workflows

### Common Commands (from catalog/)

```bash
make build          # Compile to bin/catalog
make run            # Build + execute
make test           # Run all tests
make test-cover     # Generate coverage report
make migrate-up     # Apply migrations
make migrate-down   # Rollback 1 migration
make proto          # Regenerate gRPC code from proto
```

### Testing Strategy

- **Domain/Service Unit Tests**: `testify/assert`, 95%+ coverage target
- **Repository Integration Tests**: `testcontainers` with MySQL 8.0
- **Test Lifecycle**: TestMain spins up container, runs migrations, tears down
- Example: [catalog/internal/adapters/output/repository/testhelper_test.go](catalog/internal/adapters/output/repository/testhelper_test.go#L23-L49)

### Database Migrations

- Located in `internal/infra/db/migrate/migrations/`
- Format: `YYYYMMDDHHMMSS_description.up.sql` / `.down.sql`
- Create new: `make migration <description>`
- Connection string set via `DB_URL` in Makefile

## Code Conventions

### Go Standards

- **No frameworks**: Standard library `net/http` with `ServeMux`
- **Logging**: `slog` with JSON format
- **Validation**: `go-playground/validator/v10` for input DTOs
- **Errors**: Named domain errors (e.g., `ErrMenuNotEditable`)

### Clean Architecture Rules

- **No domain import leakage**: Adapters depend on ports (interfaces), NOT domain types directly _(Known debt: REST handlers import DTOs that embed domain VOs)_
- **Dependency flow**: Adapters → Ports → Domain (dependencies point inward)
- **Application services**: Orchestrate use cases, depend only on port interfaces

### Naming Conventions

- Aggregates: PascalCase (e.g., `Menu`, `Restaurant`)
- Value Objects: PascalCase with suffix (e.g., `MenuID`, `DeliveryAddress`)
- Domain Events: Past tense (e.g., `MenuCreated`, `MenuActivated`)
- Repository methods: `Save()`, `FindById()`, `FindActiveMenuByRestaurantId()`
- Application services: Suffix `AppService` (e.g., `MenuAppService`)

## Critical Files for Reference

- `docs/PROJECT_OVERVIEW.md` — Domain model, aggregates, state machines
- `docs/ARCHITECTURE.md` — Saga flows, compensation scenarios, context mapping
- `docs/EVENTS.md` — Event definitions per bounded context
- `catalog/TODO.md` — Implementation roadmap, known technical debt
- `catalog/IDEAS.md` — Architectural trade-offs and design discussions

## Current State & Known Gaps

**Catalog Context** (🟡 In Progress):

- ✅ Domain model, aggregates, repositories, app services
- ✅ REST API with handlers
- ✅ gRPC server skeleton
- ⏳ Handler-level unit tests (missing)
- ⏳ End-to-end integration tests (missing)
- ⏳ Outbox worker/publisher (not implemented)

**Other Contexts** (Ordering, Payment, Restaurant, Delivery): 🔴 Not started

## When Adding New Features

1. **Start with domain model**: Define aggregates, entities, VOs, invariants
2. **Define domain events**: What state changes need to be communicated?
3. **Create port interfaces**: Input (DTOs) and output (repository interfaces)
4. **Implement application service**: Orchestrate use case, depend on ports
5. **Implement adapter**: REST/gRPC handler or repository implementation
6. **Write tests**: Domain unit tests first, then repository integration tests
7. **Update docs**: Reflect design decisions in `docs/` or TODO files
