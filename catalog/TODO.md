# Catalog Domain - TODO

> Última atualização: 2026-02-17
> Ref. arquitetural: `docs/ARCHITECTURE.md`

---

## 🚧 Backlog de Implementação (Pendente)

### 🟠 Fase 2: Quality Assurance & Integration
- [ ] Implementar Testes de Unidade para Handlers (Mocking Service)
- [ ] Implementar Testes de Integração (E2E) para Fluxo Completo:
  - Criar Restaurante -> Criar Menu -> Adicionar Itens -> Ativar -> Buscar Ativo
- [ ] Validar Robustez do Server Lifecycle:
  - Graceful Shutdown com requisições em andamento (drain)
  - Startup Lento/Timeouts (NotifyReady channel)

### 🟡 Fase 3: gRPC Server (Finalização)
- [ ] Adicionar Health Checks gRPC

### 🟡 Fase 3: Messaging (Outbound)
- [ ] Implementar `EventPublisher` (RabbitMQ/Kafka)
- [ ] Worker de polling para tabela `outbox`
- [ ] Garantir "At-Least-Once delivery" (Idempotency keys no consumidor)

### 🔵 Fase 4: Deployment (Kubernetes & Istio)
- [ ] Criar `Dockerfile` (Multi-stage build: builder + runner distroless/alpine)
- [ ] Configurar `.dockerignore`
- [ ] Pipeline de CI/CD (GitHub Actions)
- [ ] Kubernetes Manifests (`Deployment`, `Service`, `ConfigMap`, `Secret`)
- [ ] Istio Configuration (`Gateway`, `VirtualService`, `DestinationRule`, `PeerAuthentication`)

---

## 🛠 Débitos Técnicos & Refatoração (Prioridade Alta)

### Code Review & Clean Architecture
- [ ] **Camada Anti-Corrupção (Input Parsing)**:
  - *Problema:* `MenuAppService` recebe strings e faz parsing para VOs. Adapters (`utils.go`) importam domínio diretamente.
  - *Solução:* Mover parsing para o Handler. Remover imports de domínio em `adapters/input/rest`.
- [ ] **Error Handling Refactoring**:
  - *Problema:* `handleAppError` acopla HTTP Status com Erros de Domínio.
  - *Solução:* Criar estratégia de mapeamento de erros agnóstica (Middleware ou Mapper).
- [ ] **Validação de DTOs**:
  - *Problema:* Tags `binding` (Gin) usadas com `go-playground/validator`.
  - *Solução:* Substituir por tags `validate`.
- [ ] **Violações da Lei de Demeter**:
  - *Problema:* Acesso a campos embeddados (ex: `restaurant.RestaurantID.ID().String()`).
  - *Solução:* Usar method promotion (`restaurant.String()`). Afeta Mappers, Events e Repositories.
- [ ] **Rollback Redundante**:
  - *Problema:* `UnitOfWork` chama `Rollback` explicitamente além do `defer`.
  - *Solução:* Remover chamada explícita.

### Domínio & Regras de Negócio
- [ ] **Domain Logic Leak**:
  - *Problema:* `validateSingleItem` no Service contém regras de negócio.
  - *Solução:* Mover para `Menu.ValidateItems(...)`.
- [ ] **Invariante de Agregado (Restaurant)**:
  - *Problema:* `Restaurant.UpdateMenu` não valida ownership do menu.
  - *Solução:* Adicionar validação.
- [ ] **Erros Silenciados em Repositórios**:
  - *Problema:* Erros de parsing (`ParseMenuStatus`, `ParseUUID`) ignorados no scan do banco.
  - *Solução:* Tratar e propagar erros de corrupção de dados.
- [ ] **Side-Effect em Category.AddItem**:
  - *Problema:* `PullEvent` chamado em cópia de valor/struct.
  - *Solução:* Revisar lógica de eventos em coleções.

### Observações Menores
- [ ] Renomear `handleAddMenuCategorie` para `handleAddCategory` (typo).
- [ ] Renomear `restauntIdStr` para `restaurantIdStr` (typo).
- [ ] Remover variável global `Validate` em `utils.go`.
- [ ] Adicionar contexto de request nos logs de erro interno.

---

## ✅ Tarefas Concluídas

### Infraestrutura & Base
- [x] Configurar `docker-compose.yml` (MySQL)
- [x] Variáveis de ambiente e Config
- [x] Makefile (`up`, `down`, `logs`, `migrations`)
- [x] Migrations (`golang-migrate`) com Schema Relacional (Surrogate Keys + UUIDs)
- [x] Repositórios MySQL (`Restaurant`, `Menu`, `Outbox`) com Testcontainers

### Domínio (Core)
- [x] Agregados: `Restaurant` e `Menu` completos
- [x] Entidades: `Category`, `ItemMenu`
- [x] Value Objects: `Address`, `IDs`, `Money` (Imutáveis)
- [x] Domain Events (Fat Events: `ItemMenuCreated`, `MenuActivated`, etc.)
- [x] Specifications Pattern
- [x] Domain Services (`MenuAssignmentService`)

### Aplicação & API
- [x] `MenuAppService` e `RestaurantAppService`
- [x] Handlers HTTP (REST) mapeados
- [x] Server wiring e Graceful Shutdown
- [x] gRPC Service Definition (`.proto`) e Server Implementation
- [x] Unit of Work Pattern
- [x] Testes Unitários (> 95% cobertura de domínio)

### Refatorações (Realizadas)
- [x] Remoção de dependências de framework no domínio
- [x] Saga Pattern: Fairy Tale + Orchestrated Compensations (Decisão)
- [x] Outbox Pattern foundation
- [x] Violação de Demeter em `Menu` -> `Category` corrigida
- [x] Vazamento de Transação em `UnitOfWork` corrigido
- [x] Status Codes HTTP corrigidos
- [x] Mapeamento de erros de domínio para 422
