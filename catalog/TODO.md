# Catalog Domain - TODO

> Última atualização: 2026-02-19
> Ref. arquitetural: `docs/ARCHITECTURE.md`

---

## 🎯 Próxima Sessão — Foco Prioritário

### Para concluir o Catalog (em ordem sugerida)

1. **Fase 2 — Testes (QA)** — _pré-requisito para fechar o Catalog_
2. **Débitos de Domínio restantes** — _corrigir antes de iniciar próximo módulo_
3. **Fase 3 — Messaging (Outbox Worker)** — _pré-requisito para o Saga funcionar_
4. **CQRS — ValidateOrder** — _explorar separação de leitura/escrita em MenuAppService_

### Gate para iniciar próximo módulo (Ordering)

- [ ] Testes de Handler (unit) implementados
- [ ] Testes E2E (integração completa) implementados
- [ ] Débitos de domínio críticos resolvidos (ver seção abaixo)

---

## 🚧 Backlog de Implementação (Pendente)

### 🟠 Fase 2: Quality Assurance & Integration

- [ ] **Testes de Unidade para Handlers REST** (Mocking Service)
  - `RestaurantHandler`: criar restaurante, erros de validação, not found
  - `MenuHandler`: criar menu, ativar, arquivar, adicionar categoria/item
- [ ] **Testes de Integração (E2E) para Fluxo Completo**:
  - Criar Restaurante → Criar Menu → Adicionar Itens → Ativar → Buscar Ativo
- [ ] **Validar Robustez do Server Lifecycle**:
  - Graceful Shutdown com requisições em andamento (drain)
  - Startup Lento/Timeouts (NotifyReady channel)

### 🟡 Fase 3: gRPC Server (Finalização)

- [ ] Adicionar Health Checks gRPC (`grpc.health.v1`)

### 🔵 Fase 4: Deployment (Kubernetes & Istio)

- [ ] Criar `Dockerfile` (Multi-stage build: builder + runner distroless/alpine)
- [ ] Configurar `.dockerignore`
- [ ] Pipeline de CI/CD (GitHub Actions)
- [ ] Kubernetes Manifests (`Deployment`, `Service`, `ConfigMap`, `Secret`)
- [ ] Istio Configuration (`Gateway`, `VirtualService`, `DestinationRule`, `PeerAuthentication`)

---

## 🛠 Débitos Técnicos & Refatoração (Restantes)

### Domínio & Regras de Negócio (Todos resolvidos ✅)

> Os bugs de domínio originalmente listados foram verificados no código atual e estão todos resolvidos:
> `Menu.Activate()` order, eventos duplicados, `processChildren`, erros silenciados em repositórios,
> e side-effect em `Category.AddItem`. Ver seção "Concluídas" para detalhes.

### Code Review & Clean Architecture (Todos resolvidos ✅)

> Todos os itens de clean architecture foram resolvidos nesta sessão. Ver seção "Concluídas".

- [ ] **CQRS — `MenuAppService.ValidateOrder`** *(Prioridade Máxima)*:
  - *Problema:* `MenuAppService` acumula métodos misturando Commands e Queries. `ValidateOrder` é uma query de leitura complexa e instanciar entidades pesadas no Repositório afeta a performance.
  - *Ação:* Criar `CatalogQueryService` e `CatalogQueryRepository` usando SQL otimizado para retornar structs flat.
- [ ] **Mapeamento Semântico de Erros (Error Types):**
  - *Problema:* O arquivo `errormap.go` acopla a camada REST aos pacotes internos, usando um switch-case gigante sobre instâncias de erro (`errors.Is`). Erros crus de infra podem gerar `500` não mapeados.
  - *Ação:* Criar Custom Error Types (`NotFoundError`, `ValidationError`) no Core. Alterar `errormap.go` para usar `errors.As` validando tipos universais, desaclopando o REST dos pacotes internos (violando menos o OCP).
- [ ] **Resiliência do Outbox Processor (DLQ & Retries):**
  - *Problema:* Se o RabbitMQ estiver indiponível, o processor fica num loop infinito travando a fila (`processBatch` falha e volta ao banco).
  - *Ação:* Criar coluna de `retry_count`. Implementar Dead Letter Queue local (ignorar a mensagem se tentar 3x) para o worker prosseguir.

---

## ✅ Tarefas Concluídas

### Infraestrutura & Base
- [x] Configurar `docker-compose.yml` (MySQL)
- [x] Variáveis de ambiente e Config
- [x] Makefile (`up`, `down`, `logs`, `migrations`)
- [x] Migrations (`golang-migrate`) com Schema Relacional (Surrogate Keys + UUIDs)
- [x] Repositórios MySQL (`Restaurant`, `Menu`, `Outbox`) com Testcontainers
- [x] Schema padronizado para snake_case (`created_at`/`updated_at`) — migrations `20260218000001`, `20260218000002`

### Domínio (Core)
- [x] Agregados: `Restaurant` e `Menu` completos
- [x] Entidades: `Category`, `ItemMenu`
- [x] Value Objects: `Address`, `IDs`, `Money` (Imutáveis)
- [x] Domain Events (Fat Events: `ItemMenuCreated`, `MenuActivated`, etc.)
- [x] Specifications Pattern
- [x] Domain Services (`MenuAssignmentService`)
- [x] `Menu.ValidateItems()` — lógica de validação de itens movida do Service para o Agregado
- [x] `Restaurant.UpdateMenu` ownership garantido via `MenuAssignmentService`

### Aplicação & API
- [x] `MenuAppService` e `RestaurantAppService`
- [x] Handlers HTTP (REST) mapeados
- [x] Server wiring e Graceful Shutdown
- [x] gRPC Service Definition (`.proto`) e Server Implementation
- [x] Unit of Work Pattern
- [x] Testes Unitários (> 95% cobertura de domínio)

### Refatorações (Realizadas em 2026-02-19)
- [x] Remoção de dependências de framework no domínio
- [x] Saga Pattern: Fairy Tale + Orchestrated Compensations (Decisão)
- [x] Outbox Pattern foundation + `OutboxRepository.SaveEvents` (elimina duplicação)
- [x] Violação de Demeter em `Menu` → `Category` corrigida
- [x] Vazamento de Transação em `UnitOfWork` corrigido
- [x] `UnitOfWork`: removido `tx.Rollback()` explícito redundante (somente `defer` mantido)
- [x] Status Codes HTTP corrigidos + mapeamento de erros de domínio para 422
- [x] `RestaurantStatus.String()` corrigido: `"CLOSE"` → `"CLOSED"`
- [x] `GrpcListener` config corrigido: `"9090"` → `":9090"`
- [x] Código não utilizado removido: `getEnvInt()`, `ErrValidation`, `QueryInsertPublishedEvent`
- [x] Variável global `Validate` desexportada → `validate`
- [x] Typos corrigidos: `handleAddMenuCategorie` → `handleAddCategory`, `restauntIdStr` → `restaurantIdStr`
- [x] Typo proto corrigido: `ValidateRestauranteAndItemsRespose` → `ValidateRestaurantAndItemsResponse`
- [x] DTO tags: `binding:` → `validate:` em todos os request structs
- [x] Violações da Lei de Demeter corrigidas em `mappers.go`, `events.go` e repositórios (`.ID().String()` → `.String()`)
- [x] Error mapper extraído: `errormap.go` com `httpStatusFor(err)` separado de `utils.go`
- [x] `handleAppError` recebe `*http.Request` — contexto de request (method, path) incluído nos logs de erro
- [x] ACL `RestaurantService`: parsing de UUID movido para os handlers; interface usa `valueobjects.RestaurantID`
- [x] `ValidateOrder`: adicionada verificação `rest.CanAcceptOrder()` antes de buscar menu ativo
- [x] Testes de `ValidateOrder` adicionados (restaurante fechado, não encontrado, sem menu, itens válidos/inválidos)

### Messaging & EDA (Outbound)
- [x] Implementar `EventPublisher` Adapter (RabbitMQ via amqp091-go integrado)
- [x] Outbox Worker funcional: Polling na tabela `outbox_events` com Lock Otimista
- [x] Mapeada resiliência e "At-Least-Once delivery" (Idempotency keys prontas para o consumidor)
