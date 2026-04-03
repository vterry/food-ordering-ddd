# Catalog Domain - TODO

> Última atualização: 2026-02-22
> Ref. arquitetural: `docs/ARCHITECTURE.md`

---

## 🟡 Pendências Técnicas (Sessão Atual)

- [ ] Refatorar rota `GET /restaurant/{id}/menu` para usar **CQRS/Query Store Puro**, evitando a hidratação pesada do _Aggregation Root_ `Menu` e retornando DTO direto do banco de dados otimizado.
- [ ] Enriquecer testes unitários e de integração existentes com validações rígidas de cenários de erro e exceptions de domínio.

---

## 🔵 Fase Futura: Deployment

- [ ] `Dockerfile` (multi-stage distroless)
- [ ] `.dockerignore`
- [ ] CI/CD (GitHub Actions)
- [ ] Kubernetes Manifests
- [ ] Istio Configuration

---
## ✅ Concluído

### Infraestrutura
- [x] `docker-compose.yml`, Config, Makefile, Migrations
- [x] Repositórios MySQL (`Restaurant`, `Menu`, `Outbox`) com Testcontainers
- [x] Schema snake_case padronizado

### Domínio
- [x] Agregados `Restaurant` e `Menu`, Entidades `Category`/`ItemMenu`
- [x] Value Objects imutáveis (`Address`, `IDs`, `Money`)
- [x] Domain Events (Fat Events), Specifications, Domain Services
- [x] Todos os bugs de domínio resolvidos

### Aplicação & API
- [x] `MenuAppService`, `RestaurantAppService`, Handlers REST
- [x] gRPC Server + Proto Definition
- [x] Unit of Work Pattern
- [x] Testes unitários (>95% cobertura domínio)

### CQRS — ValidateOrder ✅
- [x] `CatalogQueryService` + `CatalogQueryRepository` (interfaces dedicadas)
- [x] Query SQL otimizada (JOIN único, sem reidratar agregados)
- [x] DTOs planos (`OrderValidationData`, `OrderValidationItem`)
- [x] Placeholders dinâmicos para `IN(?)`, fix `m.id` → `m.uuid`
- [x] Wiring completo: Repository → AppService → gRPC Server
- [x] Testes unitários (6 cenários) + integração Testcontainers (6 cenários)

### Messaging & EDA
- [x] `EventPublisher` (RabbitMQ), Outbox Worker, At-Least-Once delivery

### Refatorações (2026-02-19)
- [x] Demeter fixes, UoW leak, HTTP status codes, error mapper
- [x] ACL parsing movido para handlers, DTO tags corrigidas
- [x] Código morto removido, typos corrigidos, Saga pattern definido

### Correções Code Review + Testes (2026-02-22) ✅
- [x] `server_test.go` — Mock simplificado para `CatalogQueryService` (1 método)
- [x] `menu_service_test.go` — `NewItemID()`, dead code removido, +4 cenários ActiveMenu
- [x] **Error Types refactoring** — Interfaces `NotFoundErr`/`BusinessRuleErr`/`ValidationErr` em `common/pkg`, wrappers genéricos com `Unwrap()`, domínio e repositórios wrapam sentinels, `errormap.go` desacoplado (zero imports de domínio)
- [x] **Handler REST tests** — `restaurant_handler_test.go` (17 cenários), `menu_handler_test.go` (31 cenários) usando `testify/mock` + `httptest`
- [x] **App Service tests** — `menu_service_test.go` +24 cenários (8 métodos cobertos), `restaurant_service_test.go` 15 cenários (4 métodos) — `core/app` de 21.1% → 75.4%
- [x] Cobertura: REST 92.9%, gRPC 100%, App 75.4%, Domain 81-100%

### Teste E2E — Full Menu Lifecycle (2026-02-22) ✅
- [x] HTTP E2E com Testcontainers (MySQL real) — stack completa: Handler → Service → Domain → Repository → MySQL
- [x] Fluxo: Create Restaurant → Create Menu → Add Category → Add Item → Activate → GetActive → Verify Assignment → Verify Outbox (6 events)
- [x] Localizado em `internal/e2e/full_lifecycle_test.go`

### Débitos Técnicos Corrigidos (2026-02-22) ✅
- [x] **UUID Parse Errors** — `ParseRestaurantId`/`ParseMenuId`/`ParseCategoryId`/`ParseItemId` agora wrapam com `common.NewValidationErr()` → retornam 400 (antes: 500)
- [x] **gRPC goroutine leak** — `grpc.Server` agora é criado ANTES da goroutine, evitando race condition no `Stop()`
- [x] **`close(NotifyReady)` panic** — Protegido com `sync.Once` para evitar double-close
- [x] **`PublishWithContext` sem timeout** — Adicionado `context.WithTimeout(ctx, 5s)` no publisher RabbitMQ
- [x] **Duplicate if-err** em `server.go` removido

### Débitos Técnicos Corrigidos (2026-02-22) ✅
- [x] **Outbox DLQ** — Se RabbitMQ cair, processor trava em loop. Adicionar `retry_count` + Dead Letter Queue local.
- [x] Health Checks gRPC (`grpc.health.v1`)
