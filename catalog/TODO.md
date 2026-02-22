# Catalog Domain - TODO

> Última atualização: 2026-02-21
> Ref. arquitetural: `docs/ARCHITECTURE.md`

---

## 🎯 Próximo Passo

### 1. Correções do Code Review (~30min)

- [ ] `server_test.go` — Mock implementa `MenuService` (9 métodos, 8 com `panic`), mas o gRPC server agora depende de `CatalogQueryService` (1 método). Simplificar o mock.
- [ ] `menu_service_test.go` — `ItemID{}` zero-value no setup. Trocar por `NewItemID()`.
- [ ] `menu_service_test.go` — `NewMenu()` criado e descartado na sequência por `Restore()`. Remover linha morta.
- [ ] `menu_service_test.go` — Apenas 2 cenários para `ActiveMenu`. Adicionar: menu não encontrado, restaurante não encontrado, menu já ativo, assignment falha.

### 2. Testes de Handler REST (Unit)

- [ ] `RestaurantHandler`: criar restaurante, erros de validação, not found
- [ ] `MenuHandler`: criar menu, ativar, arquivar, adicionar categoria/item

### 3. Teste E2E (Integração Completa)

- [ ] Criar Restaurante → Criar Menu → Adicionar Itens → Ativar → Buscar Ativo

### 4. Robustez do Server Lifecycle

- [ ] Graceful Shutdown com requisições em andamento (drain)
- [ ] Startup Lento/Timeouts (NotifyReady channel)

---

## 🛠 Débitos Técnicos

- [ ] **Error Types** — `errormap.go` acopla REST aos pacotes internos via `errors.Is`. Criar `NotFoundError`/`ValidationError` no Core e usar `errors.As`.
- [ ] **Outbox DLQ** — Se RabbitMQ cair, processor trava em loop. Adicionar `retry_count` + Dead Letter Queue local.
- [ ] **Infra (`server.go` / `rabbitmq_publisher.go`)** — Goroutines do gRPC vazam silenciosamente; `close(NotifyReady)` sem `sync.Once`; `PublishWithContext` sem timeout.
- [ ] Health Checks gRPC (`grpc.health.v1`)

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
