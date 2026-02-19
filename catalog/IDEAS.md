# Banco de Ideias & Otimizações
> Espaço para documentar trade-offs arquiteturais e ideias de otimização para estudo futuro.

## 1. Reidratação de Agregados via SQL Joins (One-to-Many)

**Situação:**
Ao carregar múltiplos agregados (ex: `Menu`) que possuem entidades filhas (ex: `Category`, `Item`), uma query SQL com `JOIN` retorna múltiplas linhas para o mesmo agregado raiz. Precisamos agrupar essas linhas para reconstruir a árvore de objetos.

**Opções de Implementação:**

### A. Map Accumulation (Escolhida para MVP)
Carregar todas as linhas em memória e usar um Hash Map (`map[uuid]*MetaStruct`) para agrupar.
- **Prós:** Código simples, fácil de ler, menos propenso a bugs de estado.
- **Contras:** Maior consumo de memória (aloca tudo antes de processar).
- **Ideal para:** Paginação pequena (ex: 10-50 itens).

### B. Control Break (Streaming/Flush)
Processar linha a linha e detectar mudança de ID (`if currentID != lastID`). Quando muda, "fecha" (flush) o anterior.
- **Prós:** Baixo consumo de memória (processa 1 por vez).
- **Contras:** Código complexo (gestão de estado variáveis soltas), fácil esquecer o "último flush".
- **Ideal para:** Relatórios ou exportações massivas (milhares de linhas).

### C. Stateful Reader Object
Encapsular a lógica do "Control Break" em um struct auxiliar (`MenuRehydrator`).
- **Prós:** Código limpo no repositório, testável isoladamente.
- **Contras:** Boilerplate (mais structs/métodos para manter).
- **Ideal para:** Quando a lógica de reidratação é muito complexa e reusada.

---

## 2. Decorator Pattern para Desacoplar Outbox dos Repositories

> Registrado em: 2026-02-18
> Ref: *Implementing Domain-Driven Design* (Vaughn Vernon), *Software Architecture: The Hard Parts* (Neal Ford)

### Problema Atual

A lógica de persistência de eventos na tabela `outbox_events` está **duplicada** dentro de cada aggregate repository (`RestaurantRepository`, `MenuRepository`). Cada `Save()` faz duas coisas: persiste o estado do agregado **e** serializa+insere os domain events no outbox.

**Exemplo concreto** (código atual em `restaurant_repository.go`):

```go
func (r *RestaurantRepository) Save(ctx context.Context, agg *restaurant.Restaurant) error {
    executor := getExecutor(ctx, r.db)

    // 1. Persiste o estado do agregado
    _, err := executor.ExecContext(ctx, QueryUpsertRestaurant,
        agg.ID().String(), agg.Name(), /* ... campos ... */)
    if err != nil {
        return fmt.Errorf("failed to upsert restaurant: %w", err)
    }

    // 2. Persiste eventos — ESTA LÓGICA SE REPETE EM CADA REPOSITÓRIO
    for _, event := range agg.PullEvent() {
        payload, err := json.Marshal(event)
        if err != nil {
            return fmt.Errorf("failed to marshal event: %w", err)
        }
        _, err = executor.ExecContext(ctx, QueryInsertOutboxEvent,
            event.EventID().String(),
            agg.ID().String(),
            "Restaurant",
            event.EventName(),
            payload,
            event.OccurredOn())
        if err != nil {
            return fmt.Errorf("failed to save outbox event: %w", err)
        }
    }
    return nil
}
```

**Consequências:**
- **Shotgun Surgery**: Qualquer mudança no formato do outbox (ex: adicionar `correlation_id`) exige editar N repositórios.
- **SRP violado**: O repository do agregado conhece detalhes de serialização de eventos.
- **Testabilidade limitada**: Impossível testar persistência de outbox isoladamente.

### Ideia: Decorator Pattern (Composição)

A ideia é separar **completamente** a persistência de estado da persistência de eventos usando composição. O repositório "puro" persiste apenas o agregado. Um decorator envolve o repositório e adiciona a responsabilidade de outbox.

**Repositório puro** (só estado do agregado):

```go
// restaurant_repository.go — limpo, sem outbox
func (r *RestaurantRepository) Save(ctx context.Context, agg *restaurant.Restaurant) error {
    executor := getExecutor(ctx, r.db)
    _, err := executor.ExecContext(ctx, QueryUpsertRestaurant, /* apenas campos do agregado */)
    return err
}
```

**Decorator** (adiciona outbox por composição):

```go
// event_publishing_repository.go
type EventPublishingRestaurantRepo struct {
    inner    output.RestaurantRepository  // repositório puro
    outbox   output.OutboxRepository      // responsável por outbox
}

func NewEventPublishingRestaurantRepo(inner output.RestaurantRepository, outbox output.OutboxRepository) *EventPublishingRestaurantRepo {
    return &EventPublishingRestaurantRepo{inner: inner, outbox: outbox}
}

func (r *EventPublishingRestaurantRepo) Save(ctx context.Context, agg *restaurant.Restaurant) error {
    // 1. Delega persistência de estado ao inner
    if err := r.inner.Save(ctx, agg); err != nil {
        return err
    }
    // 2. Persiste eventos via OutboxRepository
    return r.outbox.SaveEvents(ctx, agg.ID().String(), "Restaurant", agg.PullEvent())
}

// FindById e outros métodos apenas delegam ao inner
func (r *EventPublishingRestaurantRepo) FindById(ctx context.Context, id valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
    return r.inner.FindById(ctx, id)
}
```

**Wiring** (na composição do servidor):

```go
// server.go
pureRestaurantRepo := repository.NewRestaurantRepository(db)
outboxRepo := repository.NewOutboxRepository(db)
restaurantRepo := repository.NewEventPublishingRestaurantRepo(pureRestaurantRepo, outboxRepo)
// restaurantRepo satisfaz output.RestaurantRepository — transparente para o app service
```

### Trade-offs

| Aspecto | Abordagem Atual (inline) | Decorator |
|---|---|---|
| Simplicidade | ✅ Menos structs | ❌ Mais boilerplate de delegação |
| DRY | ❌ Lógica duplicada | ✅ Centralizada |
| Testabilidade | ⚠️ Outbox testado junto | ✅ Cada camada testável isoladamente |
| SRP | ❌ Repository faz 2 coisas | ✅ Cada struct tem 1 responsabilidade |
| Transparência | ✅ Tudo visível num lugar | ⚠️ Comportamento dividido entre structs |
| Atomicidade | ✅ Via UoW/getExecutor | ✅ Mesma garantia (mesmo ctx/tx) |

### Pré-requisito

Para esta abordagem funcionar limpa, o `PullEvent()` precisa ser chamado **apenas no decorator**, não no `inner.Save()`. Isso significa que o repositório puro **não** drena os eventos — essa responsabilidade migra para o decorator.

### Quando Implementar

Considerar esta refatoração quando:
- Houver 3+ aggregate repositories com lógica de outbox
- A estrutura do outbox event mudar (ex: adicionar campos como `correlation_id`, `saga_id`)
- O projeto migrar para múltiplos bounded contexts com outbox independente
