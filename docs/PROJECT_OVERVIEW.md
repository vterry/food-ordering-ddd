# Sistema de Delivery de Comida – Case para DDD, Sagas, Outbox e EDA

## 1. Visão geral do domínio

Domínio: plataforma de delivery que conecta **cliente**, **restaurante**, **entregador** e **plataforma de pagamento**, cuidando de pedido, pagamento, catálogo e logística.[web:23][web:27]

Escopo do case:

- Cliente: cria pedido, paga, acompanha status.
- Restaurante: recebe pedidos, aceita/rejeita, prepara.
- Logística: atribui entregador, acompanha retirada e entrega.
- Pagamento: autoriza, captura, reembolsa.
- Catálogo: restaurantes, menus, itens, disponibilidade e preços.[web:23][web:26]

Objetivos técnicos:

- Exercitar **DDD (tático e estratégico)**.
- Implementar **Sagas** para transações distribuídas.[web:25][web:82]
- Implementar **Outbox pattern** para mensagens confiáveis.[web:81][web:84]
- Modelar **Event-Driven Architecture (EDA)** entre microserviços.[web:64][web:88]

---

## 2. Bounded Contexts

### 2.1. Customer Ordering Context

Responsável pelo ciclo de vida do pedido do ponto de vista do cliente.[web:23][web:24]

- Responsabilidade:
  - Criar pedidos.
  - Orquestrar a saga (Payment, Restaurant, Delivery).
  - Manter estados:
    - `PENDING → PAID → CONFIRMED → IN_DELIVERY → DELIVERED / CANCELLED / FAILED`.

### 2.2. Restaurant Management Context

Gerencia como o restaurante recebe e prepara pedidos.[web:24][web:30]

- Responsabilidade:
  - `RestaurantOrder` interno do restaurante.
  - Aceitar ou rejeitar pedidos.
  - Acompanhar preparo.
  - Estados:
    - `QUEUED → PREPARING → READY_FOR_PICKUP → HANDED_TO_COURIER / REJECTED`.

### 2.3. Delivery Logistics Context

Gerencia entregas físicas.[web:24][web:79]

- Responsabilidade:
  - Criar entrega (`Delivery`).
  - Atribuir entregador.
  - Rastrear:
    - `ASSIGNED → PICKED_UP → ON_ROUTE → DELIVERED / FAILED`.

### 2.4. Payment Context

Cuida de pagamentos.[web:27][web:80]

- Responsabilidade:
  - Iniciar, autorizar, capturar e reembolsar.
  - Estados:
    - `INITIATED → AUTHORIZED → CAPTURED → FAILED / REFUNDED`.

### 2.5. Catalog Context

Gerencia restaurantes, menus e itens.[web:22][web:26][web:53]

- Responsabilidade:
  - Restaurantes (status, horários, áreas).
  - Menus, categorias, itens, opções, preços.
  - Disponibilidade em tempo real de itens.
  - Fonte de verdade para o que o Ordering pode vender.

---

## 3. Modelagem tática

### 3.1. Customer Ordering

**Agregado: Order**

- `Order (AR)`
  - `orderId`
  - `customerId`
  - `restaurantId`
  - `deliveryAddress: DeliveryAddress (VO)`
  - `items: List<OrderItem>`
  - `totalAmount: Money (VO)`
  - `status: OrderStatus`
    - `PENDING`, `PAID`, `CONFIRMED`, `IN_DELIVERY`, `DELIVERED`, `CANCELLED`, `FAILED`
  - `paymentId`
  - `deliveryId`
  - `createdAt`, `updatedAt`.[web:27]

- `OrderItem`
  - `orderItemId` (opcional)
  - `productId` (menuItemId)
  - `name`
  - `unitPrice: Money`
  - `quantity`
  - `totalPrice: Money`
  - `chosenOptions` (nome + preço).

- `DeliveryAddress (VO)`
  - `street`, `number`, `neighborhood`, `city`, `state`, `zipCode`, `complement`.
  - Imutável, validado.

- `Money (VO)`
  - `amount`
  - `currency`.

- Invariantes:
  - Total = soma dos itens.
  - Itens não podem ser alterados após `CONFIRMED`.
  - Mudança de status apenas via métodos de domínio.

- Máquina de estados (`OrderStatus`):
  - `PENDING → PAID` (pagamento autorizado)
  - `PENDING → CANCELLED` (pagamento falhou, ou cliente cancelou)
  - `PAID → CONFIRMED` (restaurante aceitou + pagamento capturado)
  - `PAID → CANCELLED` (restaurante rejeitou, ou cliente cancelou antes do aceite)
  - `CONFIRMED → IN_DELIVERY` (entregador saiu em rota)
  - `IN_DELIVERY → DELIVERED` (entrega concluída)
  - `IN_DELIVERY → CANCELLED` (cliente recusou receber o pedido → refund total)
  - `IN_DELIVERY → FAILED` (falha na entrega — incidente operacional)
  - `PAID → FAILED` (falha na captura do pagamento — raro, mas possível)

- Distinção semântica:
  - `CANCELLED`: ação deliberada (cliente cancelou, restaurante rejeitou, pagamento recusado).
  - `FAILED`: impedimento técnico/operacional (entrega falhou, captura falhou).

---

### 3.2. Restaurant Management

**Agregado: RestaurantOrder**

- `RestaurantOrder (AR)`
  - `restaurantOrderId`
  - `restaurantId`
  - `customerOrderId`
  - `items: List<RestaurantOrderItem>`
  - `status: RestaurantOrderStatus`
    - `QUEUED`, `PREPARING`, `READY_FOR_PICKUP`, `REJECTED`
  - `estimatedPreparationTime`
  - `notes`.[web:24][web:30]

- Invariantes:
  - Transições válidas:
    - `QUEUED → PREPARING → READY_FOR_PICKUP`.
    - `QUEUED → REJECTED`.

---

### 3.3. Delivery Logistics

**Agregado: Delivery**

- `Delivery (AR)`
  - `deliveryId`
  - `courierId`
  - `customerOrderId`
  - `restaurantOrderId`
  - `pickupAddress: Address`
  - `dropoffAddress: Address`
  - `status: DeliveryStatus`
    - `ASSIGNED`, `PICKED_UP`, `ON_ROUTE`, `DELIVERED`, `FAILED`
  - `assignedAt`, `pickedUpAt`, `deliveredAt`
  - `estimatedDeliveryTime`.[web:26][web:79]

- Invariantes:
  - `PICKED_UP` só depois de `ASSIGNED`.
  - `DELIVERED` só depois de `ON_ROUTE`.

---

### 3.4. Payment

**Agregado: Payment**

- `Payment (AR)`
  - `paymentId`
  - `customerId`
  - `orderId`
  - `amount: Money`
  - `status: PaymentStatus`
    - `INITIATED`, `AUTHORIZED`, `CAPTURED`, `FAILED`, `VOIDED`, `REFUNDED`
  - `method`
  - `externalTransactionId`
  - `createdAt`, `updatedAt`.[web:27][web:80]

- Máquina de estados (`PaymentStatus`):
  - `INITIATED → AUTHORIZED` (gateway aprovou autorização)
  - `INITIATED → FAILED` (gateway recusou autorização)
  - `AUTHORIZED → CAPTURED` (captura executada após aceite do restaurante)
  - `AUTHORIZED → FAILED` (captura falhou — autorização expirada ou erro no gateway)
  - `AUTHORIZED → VOIDED` (cancelamento antes da captura — libera reserva, custo zero)
  - `CAPTURED → REFUNDED` (estorno total após captura)

- Distinção operacional:
  - **Void**: libera autorização antes da captura. Sem custo financeiro.
  - **Refund**: estorno real após captura. Pode ter custo de chargeback.

---

### 3.5. Catalog

**Agregado: Restaurant**

- `Restaurant (AR)`
  - `restaurantId`
  - `name`
  - `legalName`
  - `document (VO)`
  - `address: Address (VO)`
  - `status` (`ACTIVE`, `INACTIVE`)
  - `openingHours: OperatingHours (VO)`
  - `deliveryAreas: List<DeliveryArea (VO)>`
  - `activeMenuId`.[web:26][web:53]

**Agregado: Menu**

- `Menu (AR)`
  - `menuId`
  - `restaurantId`
  - `name`
  - `status` (`DRAFT`, `ACTIVE`, `ARCHIVED`)
  - `categories: List<MenuCategory>`
  - `validFrom`, `validTo`.[web:53]

- `MenuCategory`
  - `categoryId`
  - `name`
  - `position`
  - `items: List<MenuItem>`.

- `MenuItem`
  - `menuItemId`
  - `name`
  - `description`
  - `basePrice: Money`
  - `status` (`AVAILABLE`, `UNAVAILABLE`, `TEMP_UNAVAILABLE`)
  - `tags`
  - `options: List<MenuItemOptionGroup>`.[web:53][web:58]

- Regras:
  - Menu só é editável em `DRAFT`.
  - `activateMenu()` exige pelo menos uma categoria com itens disponíveis.
  - Ativação desativa menus ativos anteriores.

---

## 4. Eventos de domínio por contexto

### 4.1. Customer Ordering – eventos publicados

- `OrderPlaced`
- `OrderPaid`
- `OrderCancelled`
- `OrderConfirmed`
- `OrderInDelivery`
- `OrderDelivered`
- `OrderFailed`.[web:24][web:88]

### 4.2. Payment – eventos publicados

- `PaymentInitiated`
- `PaymentAuthorized`
- `PaymentCaptured`
- `PaymentCaptureFailed`
- `PaymentFailed`
- `PaymentVoided`
- `PaymentRefunded`.[web:80][web:81]

### 4.3. Restaurant – eventos publicados

- `RestaurantOrderCreated`
- `RestaurantOrderAccepted`
- `RestaurantOrderRejected`
- `RestaurantOrderStatusChanged`.[web:24]

### 4.4. Delivery – eventos publicados

- `DeliveryCreated` / `DeliveryAssigned`
- `DeliveryStatusChanged` (incluindo `PICKED_UP`, `ON_ROUTE`, `DELIVERED`, `FAILED`).[web:79][web:83]

### 4.5. Catalog – eventos publicados

- `RestaurantCreated`
- `RestaurantActivated`
- `RestaurantDeactivated`
- `MenuCreated`
- `MenuActivated`
- `MenuArchived`
- `MenuItemAvailabilityChanged`
- `MenuItemPriceChanged`.[web:22][web:53][web:58]

---

## 5. Fluxos funcionais – passo a passo para implementação

### 5.1. Fluxo A – Criação e processamento de pedido (Saga principal)

Objetivo: da criação do pedido até a entrega concluída.

#### A.1. PlaceOrder (Ordering)

1. Endpoint `POST /orders` recebe `PlaceOrderCommand`.
2. Validar dados básicos (cliente, endereço, restaurante).
3. Consultar `CatalogService`:
   - Validar restaurante ativo/aberto.
   - Validar itens e opções (IDs, disponibilidade, preço).[web:22][web:54]
4. Criar agregado `Order`:
   - Status inicial: `PENDING`.
   - Itens com snapshot do catálogo.
   - Calcular `totalAmount`.
5. Persistir `Order` e registrar domain event `OrderPlaced`.
6. Escrever `OrderPlaced` na tabela `outbox_events`.
7. Worker de outbox publica `OrderPlaced` no broker.

---

#### A.2. Payment reage a OrderPlaced

1. Consumer em Payment lê `OrderPlaced`.
2. Invoca caso de uso `InitiatePayment(orderId, customerId, amount)`.
3. Criar agregado `Payment`:
   - Status: `INITIATED`.
4. Chamar gateway de pagamento (síncrono ou assíncrono).
5. Se aprovado:
   - `payment.Authorize(externalTransactionId)`.
   - Status: `AUTHORIZED`.
   - Domain event: `PaymentAuthorized`.
6. Se recusado:
   - `payment.Fail(reason)`.
   - Status: `FAILED`.
   - Domain event: `PaymentFailed`.
7. Persistir `Payment` e escrever evento na outbox.
8. Worker publica `PaymentAuthorized` ou `PaymentFailed`.

---

#### A.3. Ordering reage a PaymentAuthorized / PaymentFailed

**PaymentAuthorized**

1. Consumer em Ordering lê `PaymentAuthorized`.
2. Carrega `Order` por `orderId`.
3. `order.MarkAsPaid(paymentId)`:
   - `PENDING → PAID`.
   - Domain event: `OrderPaid`.
4. Persistir `Order` e escrever `OrderPaid` na outbox.
5. Worker publica `OrderPaid`.

**PaymentFailed**

1. Consumer em Ordering lê `PaymentFailed`.
2. Carrega `Order`.
3. `order.Cancel("PAYMENT_FAILED")`:
   - `PENDING → CANCELLED`.
4. Persistir e escrever `OrderCancelled` na outbox.
5. Worker publica `OrderCancelled`.

---

#### A.4. Restaurant reage a OrderPaid

1. Consumer em Restaurant lê `OrderPaid`.
2. Invoca `CreateRestaurantOrder(orderPaidEvent)`.
3. Criar `RestaurantOrder`:
   - Status: `QUEUED`.
   - Referência a `customerOrderId`.
4. Persistir `RestaurantOrder` e registrar `RestaurantOrderCreated`.
5. Escrever evento na outbox e publicar.
6. Backoffice do restaurante exibe pedidos em fila.
7. Operador do restaurante usa endpoints:
   - `POST /restaurant-orders/{id}/accept`.
   - `POST /restaurant-orders/{id}/reject`.

---

#### A.5. Restaurante aceita ou rejeita pedido

**Aceitar**

1. Endpoint `POST /restaurant-orders/{id}/accept` recebe `estimatedPreparationTime`.
2. Carrega `RestaurantOrder`.
3. `restaurantOrder.Accept(estimatedPreparationTime)`:
   - `QUEUED → PREPARING`.
   - Domain event: `RestaurantOrderAccepted`.
4. Persistir e escrever evento na outbox.
5. Worker publica `RestaurantOrderAccepted`.

**Rejeitar**

1. Endpoint `POST /restaurant-orders/{id}/reject` recebe `reason`.
2. Carrega `RestaurantOrder`.
3. `restaurantOrder.Reject(reason)`:
   - `QUEUED → REJECTED`.
   - Domain event: `RestaurantOrderRejected`.
4. Persistir e escrever evento.
5. Worker publica `RestaurantOrderRejected`.

---

#### A.6. Ordering reage a RestaurantOrderAccepted / Rejected

**RestaurantOrderAccepted (abordagem otimista — confirma e captura em paralelo)**

1. Consumer em Ordering lê `RestaurantOrderAccepted`.
2. Carrega `Order`.
3. `order.Confirm(restaurantOrderId)`:
   - `PAID → CONFIRMED`.
   - Domain event: `OrderConfirmed`.
4. Persistir `Order` + outbox.
5. Publicar `OrderConfirmed`.
6. **Em paralelo**, Ordering envia comando `CapturePayment(orderId)` para Payment.
7. Payment executa captura no gateway:
   - Se OK: `payment.Capture()` → `AUTHORIZED → CAPTURED` → domain event `PaymentCaptured`.
   - Se falhou: `payment.FailCapture(reason)` → `AUTHORIZED → FAILED` → domain event `PaymentCaptureFailed`.
8. Se `PaymentCaptureFailed`:
   - Ordering recebe evento e executa compensação:
     - `order.Fail("CAPTURE_FAILED")` → `CONFIRMED → FAILED`.
     - Envia comando `CancelDelivery(orderId)` para Delivery (se já criada).
     - Notifica restaurante para cancelar preparo.

> **Nota**: falha de captura após autorização bem-sucedida é rara (o valor já foi reservado no cartão), mas não impossível (autorização expirada, problema no gateway). A abordagem otimista prioriza latência sobre consistência imediata.

**RestaurantOrderRejected**

1. Consumer em Ordering lê `RestaurantOrderRejected`.
2. Carrega `Order`.
3. `order.Cancel("RESTAURANT_REJECTED")`:
   - `PAID → CANCELLED`.
4. Persistir + outbox.
5. Publicar `OrderCancelled`.
6. Ordering envia comando `VoidPayment(orderId)` para Payment.
   - Payment executa void (libera autorização sem custo): `AUTHORIZED → VOIDED`.
   - Domain event: `PaymentVoided`.

---

#### A.7. Delivery reage a OrderConfirmed

1. Consumer em Delivery lê `OrderConfirmed`.
2. Invoca `CreateDelivery(orderConfirmedEvent)`.
3. Criar `Delivery`:
   - `PENDING_ASSIGNMENT` ou `ASSIGNED`.
4. `CourierAssignmentService.Assign(delivery)` define `courierId`.
5. `delivery.AssignCourier(courierId)`:
   - Status: `ASSIGNED`.
   - Domain event: `DeliveryAssigned`.
6. Persistir e escrever na outbox.
7. Worker publica `DeliveryCreated` / `DeliveryAssigned`.

---

#### A.8. Entregador retira e entrega (Delivery)

**Retirada – PICKED_UP**

1. Endpoint `POST /deliveries/{id}/picked-up`.
2. Carrega `Delivery`.
3. `delivery.MarkPickedUp()`:
   - `ASSIGNED → PICKED_UP`.
   - Domain event: `DeliveryPickedUp`.
4. Persistir + outbox.
5. Publicar `DeliveryStatusChanged(PICKED_UP)`.

**Sai em rota – ON_ROUTE**

1. Endpoint `POST /deliveries/{id}/on-route` (ou acoplado ao passo anterior).
2. `delivery.MarkOnRoute()`:
   - `PICKED_UP → ON_ROUTE`.
   - Domain event: `DeliveryOnRoute`.
3. Persistir + outbox e publicar `DeliveryStatusChanged(ON_ROUTE)`.

**Entrega concluída – DELIVERED**

1. Endpoint `POST /deliveries/{id}/delivered`.
2. Carrega `Delivery`.
3. `delivery.MarkDelivered()`:
   - `ON_ROUTE → DELIVERED`.
   - Domain event: `DeliveryDelivered`.
4. Persistir + outbox.
5. Publicar `DeliveryDelivered`.

---

#### A.9. Ordering reage aos eventos de Delivery

**DeliveryStatusChanged(ON_ROUTE)**

1. Consumer em Ordering lê `DeliveryStatusChanged(ON_ROUTE)`.
2. Carrega `Order`.
3. `order.MarkInDelivery()`:
   - `CONFIRMED → IN_DELIVERY`.
   - Domain event: `OrderInDelivery`.
4. Persistir + outbox.
5. Publicar `OrderInDelivery`.

**DeliveryDelivered**

1. Consumer em Ordering lê `DeliveryDelivered`.
2. Carrega `Order`.
3. `order.MarkAsDelivered(deliveryId, deliveredAt)`:
   - `IN_DELIVERY → DELIVERED`.
   - Domain event: `OrderDelivered`.
4. Persistir + outbox.
5. Publicar `OrderDelivered`.

---

### 5.2. Fluxo B – Cancelamento

Casos principais:

#### B.1. Falha de pagamento

- Já coberto em A.3:
  - `PaymentFailed` → `OrderCancelled`.

#### B.2. Restaurante rejeita pedido

- Já coberto em A.5 e A.6:
  - `RestaurantOrderRejected` → `OrderCancelled` + `VoidPayment` (pagamento ainda não capturado).

#### B.3. Cliente cancela antes de aceitação

1. Endpoint `POST /orders/{id}/cancel`.
2. Carrega `Order`.
3. Regras:
   - Se `status == PENDING`:
     - `order.Cancel("CUSTOMER_REQUEST")`.
     - Persistir + `OrderCancelled`.
     - Nenhuma ação financeira (pagamento ainda não autorizado).
   - Se `status == PAID`:
     - `order.Cancel("CUSTOMER_REQUEST")`.
     - Persistir + `OrderCancelled`.
     - Ordering envia comando `VoidPayment(orderId)` para Payment (libera autorização).
4. Worker publica `OrderCancelled`.

> **Nota**: antes da captura (antes do aceite do restaurante), o cancelamento é financeiramente barato — basta um Void para liberar a reserva no cartão, sem custo de estorno.

#### B.4. Cancelamento durante entrega (cliente recusa)

1. Cliente cancela após `IN_DELIVERY` via endpoint `POST /orders/{id}/cancel`.
2. Ordering valida que `status == IN_DELIVERY`.
3. `order.Cancel("CUSTOMER_REFUSED")`:
   - `IN_DELIVERY → CANCELLED`.
   - Domain event: `OrderCancelled`.
4. Ordering envia comando `CancelDelivery(orderId)` para Delivery.
5. Delivery:
   - `delivery.MarkFailed("CUSTOMER_CANCELLED")`.
   - Publica `DeliveryStatusChanged(FAILED)`.
6. Ordering envia comando `RefundPayment(orderId)` para Payment.
   - **Refund total** (decisão de simplificação para o case).
   - Payment executa estorno: `CAPTURED → REFUNDED` → domain event `PaymentRefunded`.

> **Nota sobre Void vs Refund**: a operação financeira depende de quando o cancelamento acontece em relação à captura do pagamento:
> - **Antes do aceite do restaurante** (PENDING, PAID): pagamento apenas autorizado → **Void** (libera reserva, custo zero).
> - **Depois do aceite** (CONFIRMED, IN_DELIVERY): pagamento já capturado → **Refund** (estorno real).

#### B.5. Falha na entrega (incidente operacional)

1. Delivery publica `DeliveryStatusChanged(FAILED)` por motivo operacional (ex: acidente, endereço não encontrado).
2. Consumer em Ordering lê o evento.
3. `order.Fail("DELIVERY_FAILED")`:
   - `IN_DELIVERY → FAILED`.
   - Domain event: `OrderFailed`.
4. Ordering envia comando `RefundPayment(orderId)` para Payment.
   - **Refund total** (decisão de simplificação).
5. Persistir + outbox e publicar `OrderFailed`.

---

### 5.3. Fluxo C – Manutenção de Catálogo

#### C.1. Criar restaurante e menu

1. `CreateRestaurant`
   - Cria `Restaurant` (INACTIVE ou ACTIVE).
   - Emite `RestaurantCreated`.
2. `ActivateRestaurant`
   - Marca `status = ACTIVE`.
   - Emite `RestaurantActivated`.
3. `CreateMenu`
   - Cria `Menu` em `DRAFT`.
   - Emite `MenuCreated`.
4. `AddCategory`, `AddMenuItem`
   - Adicionam categorias/itens dentro de `Menu`.
5. `ActivateMenu(menuId)`
   - Valida que há categorias e itens AVAILABLE.
   - `DRAFT → ACTIVE`.
   - Desativa outros menus do restaurante (`ARCHIVED`).
   - Emite `MenuActivated`.[web:53][web:58]

#### C.2. Disponibilidade de itens

1. `SetMenuItemAvailability(menuId, menuItemId, status)`
   - Atualiza status (`AVAILABLE`, `TEMP_UNAVAILABLE`).
   - Emite `MenuItemAvailabilityChanged`.
2. Read-model de catálogo consome eventos:
   - Atualiza listagem e exibição de cardápio.
3. Ordering usa `CatalogService` para respeitar disponibilidade ao criar pedidos.

---

### 5.4. Fluxo D – Read Models e UI (CQRS)

- Projeção para listagem de restaurantes:
  - Consome `RestaurantCreated`, `RestaurantActivated`, `MenuActivated`, `MenuItemAvailabilityChanged`.[web:22][web:54]
- Projeção para “meus pedidos” do cliente:
  - Consome `OrderPlaced`, `OrderConfirmed`, `OrderInDelivery`, `OrderDelivered`, `OrderCancelled`.
- Projeção para painel do restaurante:
  - Consome `RestaurantOrderCreated`, `RestaurantOrderStatusChanged`.
- Projeção para app do entregador:
  - Consome `DeliveryCreated`, `DeliveryStatusChanged`.[web:54][web:57]

---

## 6. Outbox pattern (padrão transversal)

Para cada microserviço (Ordering, Payment, Restaurant, Delivery, Catalog):[web:81][web:84][web:104]

1. Transação de escrita:
   - Persiste o agregado.
   - Insere registro em `outbox_events` com:
     - `id`
     - `aggregate_type`, `aggregate_id`
     - `event_type`
     - `payload JSON`
     - `occurred_at`
     - `processed = false`.
2. Worker de outbox:
   - Lê registros `processed = false`.
   - Publica no broker (Kafka, RabbitMQ, NATS...).
   - Marca `processed = true`.
3. Em caso de falha:
   - Retentativas.
   - Monitoria para eventos “encalhados”.

---

## 7. Ordem sugerida de implementação

1. **Catálogo + Ordering (monolito modular)**:
   - Implementar agregados `Order`, `Restaurant`, `Menu` e casos de uso de `PlaceOrder` e `CreateMenu`.
2. **Extrair Payment como microserviço**:
   - Implementar eventos `OrderPlaced`, `PaymentAuthorized`, `PaymentFailed` e outbox.
3. **Adicionar Restaurant Management**:
   - `RestaurantOrderCreated`, `RestaurantOrderAccepted`, `RestaurantOrderRejected`.
4. **Adicionar Delivery**:
   - `DeliveryCreated`, `DeliveryStatusChanged`, e integração com Ordering.
5. **Completar fluxos de cancelamento e catálogo dinâmico**.
6. **Instrumentar Sagas e EDA**:
   - Logs, métricas, dashboards por evento e por saga.[web:90][web:93]

Esse blueprint serve como base direta para implementação em Go ou outra linguagem, seguindo DDD, Sagas, Outbox e EDA.
