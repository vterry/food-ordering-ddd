# Catalog Domain - TODO

> Última atualização: 2026-02-14
> Ref. arquitetural: `docs/ARCHITECTURE.md`

---

## 🔴 Fase 1: Infraestrutura Local & Persistência

### Ambiente de Desenvolvimento

- [x] Configurar `docker-compose.yml` para MySQL dedicado ao Catalog
- [x] Definir variáveis de ambiente (`.env`) e Config (via Environment)
- [x] Criar Makefile para gestão de containers (`up`, `down`, `logs`) e Migrations

### Migrations (MySQL)

- [x] Configurar ferramenta de migração (`golang-migrate`)
- [x] Criar migration inicial (`restaurants`, `menus`, `categories`, `items`, `outbox_events`)
- [x] Definir schema relacional com Surrogate Keys e UUIDs:
  - `restaurants` (id PK, uuid UNIQUE, dados flatten, status, active_menu_uuid)
  - `menus` (id PK, uuid UNIQUE, restaurant_id UUID, status)
  - `categories` (id PK, uuid UNIQUE, menu_id FK CASCADE)
  - `items` (id PK, uuid UNIQUE, category_id FK CASCADE, price_cents)
  - `outbox_events` (id PK, uuid UNIQUE, payload JSON)

### Repositories (Adapters)

- [x] Implementar `RestaurantRepositoryMySQL`
- [x] Implementar `MenuRepositoryMySQL`
- [x] Implementar `OutboxRepositoryMySQL`
- [x] Testes de Integração com Testcontainers (ou banco local)

---

## 🟠 Fase 2: API & Wiring (Administrative - Sync)

### HTTP Handlers (net/http Standard Lib)

- [x] Mapear rotas de Restaurante (`POST /`, `GET /:id`, `PATCH /:id/open`, `PATCH /:id/close`)
- [x] Mapear rota de Storefront (`GET /restaurants/:id/active-menu` - Menu Ativo com Itens)
- [x] Mapear rotas de Menu (`POST /`, `PATCH /activate`)
- [x] Wire up: Injeção de dependência no `main.go` e `server.go`
- [x] Graceful Shutdown

### Quality Assurance (API & Integration)

- [ ] Implementar Testes de Unidade para Handlers (Mocking Service)
- [ ] Implementar Testes de Integração (E2E) para Fluxo Completo:
  - Criar Restaurante -> Criar Menu -> Adicionar Itens -> Ativar -> Buscar Ativo
- [ ] Validar Robustez do Server Lifecycle:
  - Graceful Shutdown com requisições em andamento (drain)
  - Startup Lento/Timeouts (NotifyReady channel)

---

## 🟠 Fase 2: Domain Events Enrichment (Foundation)

### Event-Carried State Transfer (Fat Events)
- [x] Refatorar eventos para incluir payload completo (evitar callbacks do consumidor):
  - [x] `ItemMenuCreated`: Adicionar Name, Price, CategoryID
  - [x] `MenuActivated`: Adicionar RestaurantID, list of ItemIDs (snapshot)
  - [x] `RestaurantCreated`: Adicionar Address, Status
  - [x] `ItemPriceChanged`: Adicionar OldPrice, NewPrice
- [x] Atualizar `OutboxRepository` para serializar novos payloads
- [x] Atualizar Testes de Unidade afetados

---

## 🟡 Fase 3: gRPC Server & Internal API (Sync Validation)

### gRPC Service Definition
- [x] Definir `catalog_service.proto`:
  - `CheckRestaurantOpen(restaurantID)`
  - `GetItemDetails(itemIDs)` (Batch retrieval para snapshot de preços)
- [x] Gerar stubs Go (`protoc-gen-go`)
- [x] Criar adaptadores: `GrpcMetadataService` (implementa a interface definida no proto)

### Implementação do Server
- [x] Implementar métodos gRPC chamando os UseCases/Services existentes
- [x] Configurar servidor gRPC no `main.go` (porta separada ou cmux)
- [ ] Adicionar Health Checks gRPC

---

## 🟡 Fase 3: Messaging (Outbound)

### Publisher Infrastructure
- [ ] Implementar `EventPublisher` (RabbitMQ/Kafka)
- [ ] Worker de polling para tabela `outbox`
- [ ] Garantir "At-Least-Once delivery"

---

## 🔵 Fase 4: Deployment (Kubernetes & Istio)

### Containerização

- [ ] Criar `Dockerfile` (Multi-stage build: builder + runner distroless/alpine)
- [ ] Configurar `.dockerignore`
- [ ] Pipeline de CI/CD (GitHub Actions) para build e push da imagem

### Kubernetes Manifests (k8s/)

- [ ] `Deployment`: Configurar réplicas, resources (requests/limits) e probes
- [ ] `Service`: Expor aplicação internamente (ClusterIP)
- [ ] `ConfigMap`: Gerenciar configurações de ambiente
- [ ] `Secret`: Gerenciar credenciais sensíveis (DB, Broker)

### Istio Configuration

- [ ] `Gateway`: Configurar entrada de tráfego no mesh
- [ ] `VirtualService`: Definir roteamento, retries e timeouts
- [ ] `DestinationRule`: Configurar políticas de tráfego (Load Balancing, Connection Pool)
- [ ] `PeerAuthentication`: Configurar mTLS (modo STRICT ou PERMISSIVE)

---

## 🟢 Concluído

### Domain Layer

- [x] Restaurant Aggregate completo
- [x] Menu Aggregate completo
- [x] Category Entity
- [x] ItemMenu Entity
- [x] Value Objects (Address, IDs) - Imutáveis e encapsulados
- [x] Domain Events (Fat Events)
- [x] Specifications (validações)
- [x] Domain Service (`MenuAssignmentService`) - Lógica no domínio

### Application Layer

- [x] Mappers Otimizados (sem alocações extras)
- [x] `MenuAppService` implementado
- [x] `RestaurantAppService` com Inversão de Dependência (`MenuAssigner`)
- [x] Interfaces `Input Ports` definidas

### Testes Unitários (Cobertura > 95%)

- [x] Domain/Menu (95.7%)
- [x] Domain/Restaurant (100%)
- [x] Domain/Services (100%)
- [x] Correção de bugs em ValueObjects (`Address`)
- [x] Validação de regras de negócio complexas
- [x] Application/MenuService (Cenários de Sucesso e Rollback)

### Testes de Integração (Repositories)

- [x] OutboxRepository: FindUnpublishedEvents, MarkAsPublished
- [x] MenuRepository: Save, FindByRestaurantId
- [x] RestaurantRepository: Save, FindById, FindAll
- [x] Testcontainers + table-driven tests

### Refatoração Arquitetural

- [x] Remoção de dependências de framework (`uuid`) no domínio
- [x] Definição de interface no consumidor (`MenuAssigner` no app layer)
- [x] Ajuste de contratos DTO/Domain (consistência de campos)
- [x] Implementação do padrão Unit of Work para consistência entre agregados

### Decisões Arquiteturais

- [x] Saga Pattern: Fairy Tale + Orchestrated Compensations (Hybrid)
- [x] Catalog interfaces: Sync (REST/gRPC for CRUD + validation), Async (Outbox events)
- [x] Knowledge base criada em `docs/ARCHITECTURE.md`

### Technical Debt / Refactoring

- [ ] Error Handling: Refatorar `handleAppError` para usar arquitetura limpa (App Errors na camada de aplicação vs. Domain Errors), removendo o acoplamento do REST com o Domínio.
