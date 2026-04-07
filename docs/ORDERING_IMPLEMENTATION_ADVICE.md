# 📝 Plano de Execução Detalhado — Ordering Context

> **Status**: Design Tático / Início da Implementação  
> **Mentor**: DDD Performance & Distributed Sagas  
> **Data de Início sugerida**: 2026-04-06

---

## 🧐 1. Questionamento Socrático (The Architecture "STOP" Phase)

Antes de iniciar a codificação, estas perguntas de design tático devem ser respondidas para garantir a integridade do Agregado e da Saga:

### **Pergunta 1: Agregado e Identidade de Itens**
*   **Contexto**: Se `OrderItem` for apenas um Value Object (snapshot imutável), como lidamos com personalizações específicas do cliente?
*   **Desafio**: Como o `Order` protege a invariante de que o preço do item **não pode ser alterado** pelo usuário (fraude), permitindo ao mesmo tempo que o usuário adicione notas (ex: "sem cebola")? O `OrderItem` deve ter uma identidade interna para ser rastreável em logs?

### **Pergunta 2: Stale Data vs. Consistência Eventual**
*   **Contexto**: Validação gRPC síncrona no Catalog.
*   **Desafio**: O Catalog diz que o item custa R$ 50 no momento da validação. Se o preço mudar para R$ 60 milissegundos antes do `Order.Save()`, qual preço prevalece? Se o Ordering "garantir" o preço capturado, como evitamos abusos de clientes que seguram o checkout por horas com um preço desatualizado?

### **Pergunta 3: Orquestração Híbrida e Resiliência**
*   **Contexto**: Uso de comandos (Orquestração) para compensações financeiras (Voids/Refunds).
*   **Desafio**: Ao enviar um comando `VoidPayment`, o Ordering deve aguardar o evento `PaymentVoided` para marcar o pedido como `CANCELLED` no banco, ou ele faz o cancelamento local e "confia" na entrega garantida do broker? Como garantimos idempotência se o comando for enviado 2x após uma falha de rede?

---

## 📅 2. Cronograma de Implementação (Dia a Dia)

### **Dia 1: O Coração do Domínio (Agregado & VOs)**
*   **Foco**: `core/domain/order` e `core/domain/valueobjects`.
*   **Ações**:
    - [ ] Criar `OrderID`, `CustomerID`, `RestaurantID` (Common Pattern).
    - [ ] Implementar `DeliveryAddress` (VO) com validações de campos obrigatórios.
    - [ ] Estruturar o Aggregate Root `Order` e a entidade/VO `OrderItem`.
    - [ ] **Regra de Ouro**: Implementar o cálculo automático de `TotalAmount` no construtor.

### **Dia 2: Comportamento, Estados e Eventos**
*   **Foco**: Rich Domain Model & State Transitions.
*   **Ações**:
    - [ ] Implementar métodos de transição: `MarkAsPaid()`, `Confirm()`, `Cancel()`, `Fail()`.
    - [ ] Validar invariantes de estado (ex: não permitir cancelar pedidos em estados terminais).
    - [ ] Definir os Domain Events: `OrderPlaced`, `OrderConfirmed`, `OrderPaid`, `OrderCancelled`, `OrderFailed`.
    - [ ] Criar testes unitários para o `Order` cobrindo todas as 9 transições previstas.

### **Dia 3: Application Layer & Ports (gRPC Client)**
*   **Foco**: Use Cases e Interfaces de Saída.
*   **Ações**:
    - [ ] Criar `PlaceOrderService` (Interage com Catalog via gRPC e salva o agredado).
    - [ ] Definir as Ports: `CatalogValidationPort`, `OrderRepository`, `OutboxRepository`.
    - [ ] Implementar o client gRPC para o Catalog (consumindo o `.proto` corporativo).

### **Dia 4: Infraestrutura e Persistência (Anti Dual-Write)**
*   **Foco**: MySQL & Unit of Work.
*   **Ações**:
    - [ ] Criar migrações SQL para as tabelas `orders` e `order_items`.
    - [ ] Implementar o `OrderRepository` em Go (MySQL adapter).
    - [ ] **Crucial**: Garantir que o `Save()` do agregador insira os eventos na tabela `outbox_events` na mesma transação JDBC/SQL do pedido.

### **Dia 5: O Maestro da Saga (Caminho Feliz)**
*   **Foco**: Messaging & Saga Orchestration.
*   **Ações**:
    - [ ] Implementar o `OrderSagaHandler` (EventListener para eventos externos).
    - [ ] Criar os consumidores para `PaymentAuthorized` e `RestaurantOrderAccepted`.
    - [ ] Implementar proteção contra duplicidade usando a tabela `processed_events` (Idempotência).

### **Dia 6: Compensações e Resiliência (Compensating Transactions)**
*   **Foco**: Fluxos de Erro e Reversão.
*   **Ações**:
    - [ ] Implementar reações para `PaymentFailed` e `RestaurantOrderRejected`.
    - [ ] Codar a emissão de Comandos de Compensação (`VoidPayment`, `RefundPayment`, `CancelDelivery`).
    - [ ] Testar cenários onde o restaurante rejeita o pedido após o pagamento autorizado.

### **Dia 7: Deployment, Observabilidade e Testes E2E**
*   **Foco**: K8s, Istio e Qualidade Final.
*   **Ações**:
    - [ ] Configurar o `Helm Chart` (`values.yaml`) para o serviço de Ordering.
    - [ ] Adicionar o módulo ao `skaffold.yaml` raiz.
    - [ ] Configurar `AuthorizationPolicy` do Istio (Zero Trust).
    - [ ] Implementar testes de integração E2E com `TestContainers`.

---

## 🚀 Checklist de Prontidão

- [ ] Respondido o Questionamento Socrático?
- [ ] Diagrama de Estados do Pedido validado?
- [ ] Schema do banco suporta BigInt PK + UUID Business ID?
- [ ] Outbox Processor está configurado para Claim-Based Locking?
