# Documento de Arquitetura — Food Ordering Platform

> **Versão:** 1.0  
> **Data:** 2026-04-20  
> **Status:** Em Revisão

---

## 1. Visão Geral da Arquitetura

O sistema adota uma arquitetura de **microserviços orientada a eventos** (EDA — Event-Driven Architecture), estruturada com base nos princípios de **Domain-Driven Design (DDD)**. Cada serviço encapsula um **Bounded Context** autônomo, com seu próprio banco de dados (MySQL), comunicando-se com os demais por meio de eventos assíncronos via **RabbitMQ** e, quando necessário, por chamadas síncronas via **gRPC**.

### Princípios Arquiteturais

1. **Autonomia de Serviço** — Cada serviço é dono de seus dados e pode ser implantado independentemente.
2. **Consistência Eventual** — Operações cross-context utilizam Saga Pattern com compensação.
3. **Comunicação Assíncrona por Padrão** — Eventos são a forma primária de comunicação; gRPC é usado apenas quando latência sub-100ms é requisito.
4. **Resiliência por Design** — Circuit breakers, retries, dead-letter queues e idempotência em todos os pontos de integração.
5. **Observabilidade End-to-End** — Correlation IDs propagados em toda a cadeia de chamadas.

---

## 2. Bounded Contexts e Serviços

O sistema é decomposto em **5 Bounded Contexts** principais, cada um mapeando para um serviço independente:

| Bounded Context | Serviço | Responsabilidade |
|-----------------|---------|------------------|
| **Ordering** | `ordering-service` | Orquestração do ciclo de vida do pedido (Saga Orchestrator) |
| **Restaurant** | `restaurant-service` | Gestão de cardápios e processamento de pedidos pelo restaurante |
| **Payment** | `payment-service` | Autorização, captura e reembolso de pagamentos |
| **Delivery** | `delivery-service` | Designação de entregadores e rastreamento de entregas |
| **Customer** | `customer-service` | Cadastro de clientes, endereços e cartões (tokenizados) |

### 2.1 Diagrama de Contexto (C4 — Nível 1)

📎 **Diagrama:** [diagrams/c4-context.puml](diagrams/c4-context.puml)

### 2.2 Diagrama de Containers (C4 — Nível 2)

📎 **Diagrama:** [diagrams/c4-container.puml](diagrams/c4-container.puml)

---

## 3. Arquitetura Interna dos Serviços (Hexagonal Architecture)

Cada serviço segue a **Arquitetura Hexagonal** (Ports & Adapters), organizada em camadas concêntricas:

📎 **Diagrama:** [diagrams/hexagonal-architecture.puml](diagrams/hexagonal-architecture.puml)

### 3.1 Estrutura de Diretórios por Serviço

```
services/<service-name>/
├── cmd/
│   └── main.go                  # Entrypoint
├── internal/
│   ├── core/
│   │   ├── domain/              # Aggregates, Entities, Value Objects, Domain Events
│   │   ├── ports/               # Interfaces (Input & Output Ports)
│   │   └── services/            # Application Services (Use Cases)
│   ├── adapters/
│   │   ├── inbound/
│   │   │   ├── http/            # HTTP Handlers (net/http)
│   │   │   ├── grpc/            # gRPC Handlers
│   │   │   └── messaging/       # RabbitMQ Consumers
│   │   └── outbound/
│   │       ├── persistence/     # MySQL Repositories (SQLC)
│   │       ├── messaging/       # RabbitMQ Publishers
│   │       └── external/        # Clients para APIs externas
│   └── config/                  # Configuração do serviço
├── api/
│   ├── proto/                   # Protobuf definitions
│   └── openapi/                 # OpenAPI specs (se HTTP)
├── db/
│   ├── migrations/              # SQL migrations
│   ├── queries/                 # SQL queries (SQLC input)
│   └── sqlc.yaml                # SQLC configuration
└── Dockerfile
```

---

## 4. Design de Domínio por Bounded Context

### 4.1 Ordering Context (Saga Orchestrator)

O **Ordering Context** é o coração do sistema. Ele orquestra todo o ciclo de vida do pedido usando o **Saga Pattern (Orchestration)**.

#### Aggregates

| Aggregate | Entities / VOs | Responsabilidade |
|-----------|---------------|------------------|
| **Order** | `OrderItem (VO)`, `OrderStatus (VO)`, `Money (VO)` | Ciclo de vida completo do pedido |

#### Domain Events (Publicados)

| Evento | Quando Ocorre | Dados |
|--------|---------------|-------|
| `OrderCreated` | Cliente confirma o carrinho e submete o pedido | orderId, customerId, restaurantId, items[], totalAmount |
| `OrderCancelled` | Pedido é cancelado (por cliente, restaurante ou sistema) | orderId, reason, cancelledBy |
| `OrderConfirmed` | Restaurante aceita e pagamento é capturado | orderId |

#### Commands (Enviados para outros contextos)

| Comando | Destino | Propósito |
|---------|---------|-----------|
| `AuthorizePayment` | Payment Service | Reservar valor no cartão |
| `CapturePayment` | Payment Service | Efetivar cobrança |
| `ReleasePayment` | Payment Service | Liberar reserva |
| `RefundPayment` | Payment Service | Solicitar reembolso |
| `CreateTicket` | Restaurant Service | Enviar pedido ao restaurante |
| `CancelTicket` | Restaurant Service | Cancelar pedido no restaurante |
| `ScheduleDelivery` | Delivery Service | Solicitar designação de entregador |
| `CancelDelivery` | Delivery Service | Cancelar entrega |

### 4.2 Restaurant Context

#### Aggregates

| Aggregate | Entities / VOs | Responsabilidade |
|-----------|---------------|------------------|
| **Restaurant** | `Address (VO)`, `OperatingHours (VO)` | Dados do estabelecimento |
| **Menu** | `MenuItem (Entity)`, `Money (VO)`, `Category (VO)` | Cardápios e itens |
| **Ticket** | `TicketItem (VO)`, `TicketStatus (VO)` | Representação local do pedido recebido |

#### Domain Events (Publicados)

| Evento | Quando Ocorre |
|--------|---------------|
| `TicketConfirmed` | Restaurante aceita o pedido |
| `TicketRejected` | Restaurante recusa o pedido |
| `TicketReady` | Pedido pronto para coleta |
| `TicketCancelled` | Pedido cancelado no restaurante |

### 4.3 Payment Context

#### Aggregates

| Aggregate | Entities / VOs | Responsabilidade |
|-----------|---------------|------------------|
| **Payment** | `PaymentStatus (VO)`, `Money (VO)`, `CardToken (VO)` | Gerenciar estado do pagamento |

#### Domain Events (Publicados)

| Evento | Quando Ocorre |
|--------|---------------|
| `PaymentAuthorized` | Reserva aprovada pelo gateway |
| `PaymentAuthorizationFailed` | Reserva recusada pelo gateway |
| `PaymentCaptured` | Captura efetuada com sucesso |
| `PaymentCaptureFailed` | Captura falhou |
| `PaymentRefunded` | Reembolso processado |
| `PaymentReleased` | Reserva liberada |

### 4.4 Delivery Context

#### Aggregates

| Aggregate | Entities / VOs | Responsabilidade |
|-----------|---------------|------------------|
| **Delivery** | `CourierInfo (VO)`, `DeliveryStatus (VO)`, `Address (VO)` | Ciclo de vida da entrega |

#### Domain Events (Publicados)

| Evento | Quando Ocorre |
|--------|---------------|
| `DeliveryScheduled` | Entregador designado |
| `DeliveryPickedUp` | Entregador coletou o pedido |
| `DeliveryCompleted` | Pedido entregue ao cliente |
| `DeliveryRefused` | Cliente recusou o recebimento |
| `DeliveryCancelled` | Entrega cancelada |

### 4.5 Customer Context

#### Aggregates

| Aggregate | Entities / VOs | Responsabilidade |
|-----------|---------------|------------------|
| **Customer** | `Name (VO)`, `Email (VO)`, `Phone (VO)` | Dados do cliente |
| **Address** | `Street (VO)`, `City (VO)`, `ZipCode (VO)` | Endereços de entrega |
| **Cart** | `CartItem (VO)`, `Money (VO)` | Carrinho de compras ativo |

> **Nota:** O Cart é um aggregate de vida curta (transiente). Ele existe enquanto o cliente monta seu pedido e é "consumido" ao criar a Order.

---

## 5. Fluxos Principais — Diagramas de Sequência

### 5.1 Fluxo Principal: Criação de Pedido (Happy Path)

📎 **Diagrama:** [diagrams/seq-order-happy-path.puml](diagrams/seq-order-happy-path.puml)

### 5.2 Fluxo de Compensação: Falha no Pagamento

📎 **Diagrama:** [diagrams/seq-payment-failure.puml](diagrams/seq-payment-failure.puml)

### 5.3 Fluxo de Compensação: Recusa do Restaurante

📎 **Diagrama:** [diagrams/seq-restaurant-rejection.puml](diagrams/seq-restaurant-rejection.puml)

### 5.4 Fluxo de Compensação: Cliente Recusa na Entrega

📎 **Diagrama:** [diagrams/seq-delivery-refusal.puml](diagrams/seq-delivery-refusal.puml)

### 5.5 Fluxo de Cancelamento pelo Cliente

📎 **Diagrama:** [diagrams/seq-client-cancellation.puml](diagrams/seq-client-cancellation.puml)

---

## 6. Saga de Criação de Pedido — Máquina de Estados

O Ordering Service implementa uma **Saga Orchestrator** (SEC — Saga Execution Coordinator). O diagrama abaixo mostra a máquina de estados completa:

📎 **Diagrama:** [diagrams/saga-state-machine.puml](diagrams/saga-state-machine.puml)

---

## 7. Topologia de Mensageria (RabbitMQ)

O RabbitMQ é configurado com **topic exchanges** para permitir roteamento flexível:

📎 **Diagrama:** [diagrams/rabbitmq-topology.puml](diagrams/rabbitmq-topology.puml)

### 7.1 Estratégia de Exchanges

| Exchange | Tipo | Serviço Owner | Propósito |
|----------|------|---------------|-----------|
| `ordering.exchange` | Topic | Ordering | Publicação de comandos para outros serviços |
| `payment.exchange` | Topic | Payment | Publicação de eventos de pagamento |
| `restaurant.exchange` | Topic | Restaurant | Publicação de eventos de ticket |
| `delivery.exchange` | Topic | Delivery | Publicação de eventos de entrega |

### 7.2 Convenção de Routing Keys

```
<context>.<aggregate>.<action>

Exemplos:
  payment.authorize         → Comando para autorizar pagamento
  payment.authorized        → Evento de pagamento autorizado
  restaurant.ticket.create  → Comando para criar ticket
  restaurant.ticket.confirmed → Evento de ticket confirmado
```

---

## 8. Padrões de Infraestrutura

### 8.1 Outbox Pattern

Para garantir consistência entre o estado do aggregate e a publicação de eventos, todos os serviços implementam o **Outbox Pattern**:

📎 **Diagrama:** [diagrams/outbox-pattern.puml](diagrams/outbox-pattern.puml)

#### Schema da Tabela Outbox

```sql
CREATE TABLE outbox_messages (
    id              CHAR(36)     PRIMARY KEY,  -- UUID
    aggregate_type  VARCHAR(100) NOT NULL,
    aggregate_id    CHAR(36)     NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    payload         JSON         NOT NULL,
    correlation_id  CHAR(36)     NOT NULL,
    created_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at    TIMESTAMP    NULL
);

CREATE INDEX idx_outbox_unpublished ON outbox_messages(published_at, created_at);
```

### 8.2 Idempotent Consumer

Para garantir idempotência no processamento de mensagens:

📎 **Diagrama:** [diagrams/idempotent-consumer.puml](diagrams/idempotent-consumer.puml)

#### Schema da Tabela de Idempotência

```sql
CREATE TABLE processed_messages (
    message_id   CHAR(36)  PRIMARY KEY,
    processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### 8.3 Circuit Breaker (para chamadas ao Payment Gateway)

📎 **Diagrama:** [diagrams/circuit-breaker.puml](diagrams/circuit-breaker.puml)

---

## 9. Modelo de Dados Simplificado por Serviço

### 9.1 Ordering Service

📎 **Diagrama:** [diagrams/er-ordering.puml](diagrams/er-ordering.puml)

### 9.2 Restaurant Service

📎 **Diagrama:** [diagrams/er-restaurant.puml](diagrams/er-restaurant.puml)

---

## 10. Comunicação entre Serviços — Resumo

📎 **Diagrama:** [diagrams/communication-matrix.puml](diagrams/communication-matrix.puml)

### 10.1 Resumo de Protocolos

| De → Para | Protocolo | Padrão | Justificativa |
|-----------|-----------|--------|---------------|
| Ordering → Payment | Async (RabbitMQ) | Command/Event | Operações de pagamento são long-running |
| Ordering → Restaurant | Async (RabbitMQ) | Command/Event | Confirmação do restaurante é assíncrona por natureza |
| Ordering → Delivery | Async (RabbitMQ) | Command/Event | Designação de entregador é assíncrona |
| Ordering → Customer | Sync (gRPC) | Request/Response | Consulta de dados do cliente é rápida e necessária antes de criar o pedido |
| API Gateway → Serviços | HTTP / gRPC | Request/Response | Interface pública para clientes |

---

## 11. Decisões Arquiteturais (ADRs)

### ADR-001: Saga Orchestration vs Choreography

**Status:** Aceito

**Contexto:** O fluxo de criação de pedido envolve 4 serviços e múltiplos passos de compensação. Precisamos de uma estratégia para coordenar transações distribuídas.

**Decisão:** Usar **Saga Orchestration** com o Ordering Service como SEC (Saga Execution Coordinator).

**Alternativas Consideradas:**
- **Choreography** — Cada serviço reage a eventos e emite novos eventos. Mais simples para fluxos pequenos, mas gera *event spaghetti* quando há múltiplos caminhos de compensação.
- **Hybrid** — Happy path via choreography, compensação via orchestration. Adiciona complexidade sem ganho significativo para nosso caso.

**Consequências:**
- ✅ Lógica de compensação centralizada e explícita
- ✅ Facilita debugging e monitoramento
- ❌ Ordering Service se torna um single point of coordination
- ❌ Maior acoplamento de conhecimento (Ordering conhece todos os passos)

**Trade-off:** Centralização do fluxo e facilidade de debug são priorizados sobre independência total dos serviços.

---

### ADR-002: MySQL como Datastore Único

**Status:** Aceito

**Contexto:** Precisamos de um banco relacional com suporte ACID e boa integração com SQLC.

**Decisão:** Usar **MySQL** como banco de dados para todos os serviços, cada um com seu schema/database isolado.

**Alternativas Consideradas:**
- **PostgreSQL** — Mais features (JSON, full-text), mas MySQL foi especificado como requisito.
- **Multi-database** — Usar diferentes bancos por serviço. Desnecessário neste estágio.

**Consequências:**
- ✅ Stack simplificada (um tipo de banco)
- ✅ Excelente suporte do SQLC para MySQL
- ❌ Limitações em queries complexas comparado ao PostgreSQL
- ❌ JSON support menos maduro que PostgreSQL

---

### ADR-003: Outbox Pattern para Event Publishing

**Status:** Aceito

**Contexto:** Precisamos garantir que eventos sejam publicados atomicamente junto com mudanças de estado, evitando inconsistências entre banco e mensageria.

**Decisão:** Implementar **Outbox Pattern** com polling para todos os serviços.

**Alternativas Consideradas:**
- **Transactional Outbox + CDC (Debezium)** — Mais eficiente mas adiciona dependência de infraestrutura (Kafka Connect).
- **Event Sourcing** — Garante consistência por design mas muda radicalmente o modelo de persistência.
- **Dual Write** — Escrever no banco e publicar no RabbitMQ separadamente. **Rejeitado** por risco de inconsistência.

**Consequências:**
- ✅ Consistência garantida entre estado e eventos
- ✅ Simplicidade de implementação
- ❌ Latência adicional do polling (mitigável com intervalo baixo)
- ❌ Carga extra no banco de dados pelo polling

---

### ADR-004: gRPC para Comunicação Síncrona

**Status:** Aceito

**Contexto:** Algumas operações necessitam de resposta síncrona com baixa latência (ex: consultar dados do cliente durante criação do pedido).

**Decisão:** Usar **gRPC** para comunicação síncrona entre serviços internos.

**Alternativas Consideradas:**
- **HTTP REST** — Mais simples, mas sem tipagem de contrato e overhead de serialização JSON.
- **GraphQL** — Apropriado para BFF, não para comunicação serviço-a-serviço.

**Consequências:**
- ✅ Contratos tipados via Protobuf
- ✅ Performance superior (HTTP/2, serialização binária)
- ✅ Geração de código em Go
- ❌ Debugging mais complexo (não é legível como JSON)
- ❌ Requer geração de código e tooling adicional

---

## 12. Riscos e Mitigações

| # | Risco | Impacto | Probabilidade | Mitigação |
|---|-------|---------|---------------|-----------|
| 1 | Ordering Service como SPOF | Alto | Média | Múltiplas instâncias, health checks, auto-recovery |
| 2 | Mensagens perdidas no RabbitMQ | Alto | Baixa | Publisher confirms, persistent queues, DLQ |
| 3 | Inconsistência Outbox-RabbitMQ | Médio | Baixa | Idempotent consumers, at-least-once delivery |
| 4 | Falha no Payment Gateway | Alto | Média | Circuit breaker, retry com backoff, fallback |
| 5 | Saga stuck em estado intermediário | Alto | Baixa | Timeout por step, compensação automática, monitoramento |
| 6 | Race condition: cliente cancela durante confirmação do restaurante | Médio | Média | Optimistic locking (version field), claim-based locking |

---

## 13. Roadmap de Implementação Sugerido

| Fase | Módulo | Foco |
|------|--------|------|
| **1** | Customer Service | Cadastro, endereços, carrinho |
| **2** | Restaurant Service | Cardápios, itens, CRUD |
| **3** | Payment Service | Autorização, captura, reembolso (mock gateway) |
| **4** | Ordering Service | Saga Orchestrator, máquina de estados |
| **5** | Delivery Service | Designação, rastreamento |
| **6** | Infraestrutura | Outbox poller, idempotent consumer, DLQ |
| **7** | Integração | Testes E2E, fluxos de compensação |
| **8** | Observabilidade | Correlation IDs, logging, métricas |

---

## Índice de Diagramas

| # | Diagrama | Arquivo | Tipo |
|---|----------|---------|------|
| 1 | Contexto (C4 Nível 1) | [c4-context.puml](diagrams/c4-context.puml) | Contexto |
| 2 | Containers (C4 Nível 2) | [c4-container.puml](diagrams/c4-container.puml) | Container |
| 3 | Arquitetura Hexagonal | [hexagonal-architecture.puml](diagrams/hexagonal-architecture.puml) | Componente |
| 4 | Criação de Pedido (Happy Path) | [seq-order-happy-path.puml](diagrams/seq-order-happy-path.puml) | Sequência |
| 5 | Falha no Pagamento | [seq-payment-failure.puml](diagrams/seq-payment-failure.puml) | Sequência |
| 6 | Recusa do Restaurante | [seq-restaurant-rejection.puml](diagrams/seq-restaurant-rejection.puml) | Sequência |
| 7 | Recusa na Entrega | [seq-delivery-refusal.puml](diagrams/seq-delivery-refusal.puml) | Sequência |
| 8 | Cancelamento pelo Cliente | [seq-client-cancellation.puml](diagrams/seq-client-cancellation.puml) | Sequência |
| 9 | Saga — Máquina de Estados | [saga-state-machine.puml](diagrams/saga-state-machine.puml) | Estado |
| 10 | Topologia RabbitMQ | [rabbitmq-topology.puml](diagrams/rabbitmq-topology.puml) | Componente |
| 11 | Outbox Pattern | [outbox-pattern.puml](diagrams/outbox-pattern.puml) | Sequência |
| 12 | Idempotent Consumer | [idempotent-consumer.puml](diagrams/idempotent-consumer.puml) | Sequência |
| 13 | Circuit Breaker | [circuit-breaker.puml](diagrams/circuit-breaker.puml) | Estado |
| 14 | ER — Ordering Service | [er-ordering.puml](diagrams/er-ordering.puml) | ER |
| 15 | ER — Restaurant Service | [er-restaurant.puml](diagrams/er-restaurant.puml) | ER |
| 16 | Matriz de Comunicação | [communication-matrix.puml](diagrams/communication-matrix.puml) | Componente |
