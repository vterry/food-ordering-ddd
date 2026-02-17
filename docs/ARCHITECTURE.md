# Food Ordering DDD - Architecture Knowledge Base

## Bounded Contexts

| Context        | Role                                                       | Status         |
| -------------- | ---------------------------------------------------------- | -------------- |
| **Catalog**    | Manages restaurants, menus, categories, items              | 🟡 In Progress |
| **Ordering**   | Manages order lifecycle (create, confirm, cancel, deliver) | 🔴 Not Started |
| **Payment**    | Handles payment authorization, refunds                     | 🔴 Not Started |
| **Restaurant** | Accepts/rejects orders from restaurant perspective         | 🔴 Not Started |
| **Delivery**   | Assigns couriers, tracks delivery                          | 🔴 Not Started |
| **Kitchen**    | Manages ticket issuance and preparation                    | 🟡 In Progress |

---

## Order Creation Flow (Saga)

Reference diagram: `docs/images/CriaçãoPedido.png`

### Happy Path (Orchestrated)

```
1. Client → POST /orders/checkout → Ordering Context (Orchestrator)

2. Ordering executes PARALLEL Validations (gRPC/Internal):
   - Thread A (Catalog): ValidateItems(restaurantId, itemIds) -> Returns Price Snapshot
   - Thread B (Payment): ValidatePaymentMethod(cardToken)
   - Thread C (Kitchen): CheckAvailability(restaurantId) (Optional)

3. Aggregation (Join):
   - If ANY fails: Return Error immediately (fail-fast)
   - If ALL pass: 

4. Order Creation (Local Transaction):
   - Create Order(PENDING)
   - Persist Order Snapshot (prices from Catalog)
   - Publish: OrderCreated [outbox]

5. Async Fulfillment (Choreography starts here or continues Orchestrated):
   - Payment listens to OrderCreated -> Authorize()
   - Restaurant listens to OrderPaid -> Accept()
   - Delivery listens to OrderConfirmed -> Dispatch()
```

### Compensation Flows (3 Error Scenarios)

```
ERRO 1: PaymentFailed
  - Payment publishes PaymentFailed
  - Ordering calls Cancel("PAYMENT_FAILED")
  - Publishes: OrderCancelled [outbox]
  - Saga terminates

ERRO 2: RestaurantOrderRejected
  - Restaurant publishes RestaurantOrderRejected
  - Ordering calls Cancel("RESTAURANT_REJECTED")
  - Ordering sends RefundPayment command
  - Payment calls Refund() → publishes PaymentRefunded
  - Compensation complete

ERRO 3: DeliveryFailed
  - Delivery publishes DeliveryStatusChanged(FAILED)
  - Ordering calls MarkFailed("DELIVERY_FAILED")
  - Ordering sends RefundPayment(partial)
  - Partial refund policy applies
```

---

## Saga Pattern Analysis (Based on "Software Architecture: The Hard Parts")

### Decision Matrix

| Dimension          | Our Choice                       | Rationale                         |
| ------------------ | -------------------------------- | --------------------------------- |
| **Communication**  | Hybrid (gRPC + Events)           | Performance for validation, Decoupling for lifecycle |
| **Coordination**   | Orchestration (Smart Endpoint)   | Centralized control in Ordering   |
| **Consistency**    | Eventual                         | No distributed transactions       |
| **Error Handling** | Backward Recovery (Compensation) | Orchestrator triggers compensations |

### Pattern Classification: Parallel Saga (Orchestrated) 🎻

- **Orchestration + Parallel Sync Validation + Async Fulfillment**
- **Ordering Context** acts as the **Orchestrator** (Maestro).
- **Parallel Execution:** When an order is placed, the orchestrator triggers parallel validations (Catalog, Payment, localized Inventory).

### Why this pattern?

| Pattern                                       | Why Selected/Rejected                     |
| --------------------------------------------- | ----------------------------------------- |
| **Parallel Saga** (Selected)                  | Reduces latency by running independent validations concurrently via gRPC. |
| **Fairy Tale** (Rejected)                     | Choreography becomes hard to track with multiple validation steps. |
| **Epic Saga** (Rejected)                      | Too much synchronous coupling for the entire lifecycle. |

### ⚠️ PENDING DECISION: Hybrid Approach

**Context:** Compensation flows become progressively complex:

- PaymentFailed → 1 step (simple cancel)
- RestaurantRejected → 2 steps (cancel + refund)
- DeliveryFailed → 2+ steps (mark failed + partial refund with business policy)

**Consideration:** When implementing the **Ordering Context**, evaluate whether compensations
should use **orchestration** instead of choreography. The book suggests a **Hybrid Approach**:
choreography on the happy path, orchestration on compensations.

**Decision Point:** This should be revisited when implementing the Ordering module.
The Catalog module is NOT affected by this decision — it only:

1. Exposes a sync API for item/restaurant validation
2. Publishes domain events via Outbox pattern

---

## Context Mapping

### Catalog Context Responsibilities

- **Upstream (Publisher):** Publishes MenuActivated, MenuArchived, RestaurantCreated events
- **Downstream (Queried):** Responds to sync queries from Ordering (validate items, price snapshot)
- **Integration Style:** Outbox Pattern for events, REST API for queries

### Inter-Context Communication Patterns

- **Outbox Pattern:** All contexts use transactional outbox for event publishing.
- **Event-Carried State Transfer:** Events include full payloads (Fat Events) so consumers can build local read models without querying back.
- **gRPC (Internal):** Used for synchronous, low-latency validation calls within the cluster (Ordering -> Catalog).

---

## Technical Decisions Log

| Date       | Decision                                            | Context      | Status            |
| ---------- | --------------------------------------------------- | ------------ | ----------------- |
| 2026-02-10 | MySQL as primary database                           | All contexts | ✅ Active         |
| 2026-02-10 | Surrogate keys (auto-increment) + UUID business IDs | All contexts | ✅ Active         |
| 2026-02-11 | Transactional Outbox for event publishing           | All contexts | ✅ Active         |
| 2026-02-14 | Testcontainers for integration tests                | Catalog      | ✅ Active         |
| 2026-02-16 | Parallel Saga (Orchestrated Validation)             | Ordering     | ✅ Active         |
| 2026-02-16 | gRPC for internal sync validation                   | Catalog      | ✅ Active         |
| 2026-02-16 | Event-Carried State Transfer (Fat Events)           | All contexts | ✅ Active         |
