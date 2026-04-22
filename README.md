# Food Ordering Platform — Backend DDD & EDA

Este é um projeto de referência de uma plataforma de delivery de comida, desenvolvido como monorepo utilizando **Go**. O objetivo principal é demonstrar a aplicação prática de conceitos avançados de arquitetura distribuída, com foco em **Domain-Driven Design (DDD)**, **Event-Driven Architecture (EDA)** e padrões de resiliência.

---

## 🚀 Conceitos Core

O projeto foi construído sobre quatro pilares fundamentais:

1.  **Domain-Driven Design (DDD):** Decomposição do sistema em **Bounded Contexts** autônomos, cada um com sua própria lógica de domínio, linguagem ubíqua e isolamento de estado.
2.  **Event-Driven Architecture (EDA):** Comunicação assíncrona entre serviços utilizando **RabbitMQ**. O sistema reage a eventos de domínio para manter a consistência eventual.
3.  **Arquitetura Hexagonal (Ports & Adapters):** Desacoplamento total da lógica de negócio de detalhes de infraestrutura (bancos de dados, APIs externas, mensageria).
4.  **Saga Pattern (Orchestration):** Gerenciamento de transações distribuídas complexas (como a criação de um pedido) através de um orquestrador centralizado no `ordering-service`, garantindo fluxos de compensação em caso de falhas.
5.  **Test-Driven Development (TDD):** Desenvolvimento guiado por testes como premissa obrigatória em todos os módulos. Todo novo comportamento deve ser precedido por um teste unitário ou de integração, garantindo que o código atenda aos requisitos desde a sua concepção.


---

## 📂 Estrutura do Monorepo

O repositório está organizado por contextos delimitados, seguindo o padrão de um microserviço por pasta:

```text
.
├── common/             # Código compartilhado entre serviços (pkg/domain, pkg/errors, etc.)
├── customer/           # Bounded Context de Clientes: Cadastro, Endereços e Carrinho.
├── ordering/           # Bounded Context de Pedidos: Saga Orchestrator e Ciclo de Vida.
├── payment/            # Bounded Context de Pagamentos: Autorização e Captura.
├── restaurant/         # Bounded Context de Restaurantes: Gestão de Cardápios e Tickets.
├── delivery/           # Bounded Context de Entrega: Designação de Entregadores.
├── docs/               # Documentação detalhada, ADRs e Diagramas.
└── Makefile / Taskfile # Automação de tarefas e builds.
```

---

## 🛠 Tech Stack

| Componente | Tecnologia |
| :--- | :--- |
| **Linguagem** | Go (Golang) |
| **Banco de Dados** | MySQL (Database-per-service) |
| **Mensageria** | RabbitMQ (Topic Exchanges) |
| **Comunicação Síncrona** | gRPC (Protobuf) |
| **Interface Pública** | REST (HTTP nativo + OpenAPI) |
| **Persistence Layer** | SQLC (Type-safe SQL) |

---

## 🏗 Destaques da Arquitetura

### 1. Saga Orchestration
A criação de um pedido envolve múltiplos passos: reserva de saldo, confirmação do restaurante, captura do pagamento e entrega. O `ordering-service` atua como o **Saga Execution Coordinator (SEC)**, disparando comandos e reagindo a eventos para avançar ou compensar (rollback) a transação.

### 2. Outbox Pattern
Para garantir que um evento de domínio nunca seja perdido, utilizamos o **Transactional Outbox**. O estado do domínio e a mensagem a ser enviada são salvos na mesma transação atômica do banco de dados.

### 3. Idempotent Consumers
Todos os consumidores de eventos implementam verificações de idempotência, garantindo que o processamento duplicado de mensagens (comum em redes distribuídas) não cause efeitos colaterais indesejados.

### 4. Circuit Breaker
Implementado para proteger o sistema contra falhas em cascata, especialmente em integrações com gateways de pagamento externos.

---

## 📖 Documentação Detalhada

Para entender profundamente as decisões de design, consulte a pasta `docs/`:

*   [**Documento de Arquitetura**](docs/ARCHITECTURE.md): Visão técnica completa, C4 Models e definições de comunicação.
*   [**Documento de Requisitos**](docs/REQUIREMENTS.md): Regras de negócio e fluxos funcionais.
*   [**Decisões Arquiteturais (ADRs)**](docs/ARCHITECTURE.md#11-decisões-arquiteturais-adrs): O "porquê" por trás das escolhas tecnológicas.
*   [**Plano de Implementação**](docs/IMPLEMENTATION_PLAN.md): Roadmap de desenvolvimento faseado.

---

## 🛠 Como Executar (Em breve)

O projeto está em fase de implementação. No futuro, você poderá subir toda a infraestrutura e serviços utilizando:

```bash
docker-compose up -d
```

---
*Desenvolvido como um projeto de referência em Go.*
