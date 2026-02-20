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

---

## 3. Garantia de Ordem e Idempotência (At-Least-Once Delivery)

> Registrado em: 2026-02-19
> Ref: Discussão de Design Tático (Mensageria e Sagas)

### Problema Atual: Duplicidade e Concorrência de Eventos

Na arquitetura atual com o **Outbox Pattern**, utilizamos uma tabela no MySQL e um *Worker* (com *polling* e *Pessimistic Locking* via `SELECT ... FOR UPDATE SKIP LOCKED`).
Isso gera uma garantia de entrega **At-Least-Once** para o RabbitMQ.

O cenário de conflito/duplicidade acontece da seguinte forma:
1. O Worker lê um evento da tabela `outbox_events`.
2. O evento é publicado com sucesso no RabbitMQ.
3. **Antes** que o Worker consiga atualizar a coluna `processed` para `true` no banco de dados e dar o COMMIT, ocorre uma falha brusca (ex: Out-Of-Memory, crash do pod).
4. A transação do MySQL sofre Rollback, o lock é liberado e o evento volta a pendência.
5. Quando o Worker (ou outra instância) rodar novamente, ele republicará a mesma mensagem.

Adicionalmente, utilizando RabbitMQ e sendo necessário um roteamento avançado, a garantia universal de **ordem dos eventos** distribuídos sem gargalos (ex: um Menu alterado múltiplas vezes numa janela curta) fica sob mais propensão a desafios.

### Solução Adotada Atualmente (RabbitMQ + Inbox / Consumer Track)

1.  **RabbitMQ for Message Routing:** Simplifica a adoção inicial focada na roteirização de eventos entre os contextos de negócio (*Smart Broker*).
2.  **Idempotency Key & Inbox Pattern:** Os serviços consumidores downstream (Ex: Ordering, Delivery) utilizarão uma tabela `inbox_events` local no seu próprio banco de dados, protegida por chave primária baseada no `EventID`. O processamento será agrupado em uma transação local do consumidor. Se houver falha de chave duplicada (violação da PK), o consumidor saberá que o evento é uma duplicata, fará rollback do processamento local daquela request, mas retornará "ACK" pro RabbitMQ limpando a fila silenciosamente.

### Evolução Futura: Kafka & Event Sourced Read Models

Para explorar cenários de escala maciça e garantir de forma sistêmica a consistência eventual e a ordem, as seguintes evoluções arquiteturais estão mapeadas:

1.  **Migração para o Apache Kafka (Dumb Broker / Smart Consumer):**
    *   **Garantia de Ordem Absoluta:** Ao usar a chave do agregado (ex: `MenuID`, `RestaurantID`) como *Partition Key* no Kafka, todos os eventos do mesmo agregado cairão na mesma partição, garantindo um processamento estritamente FIFO por domínio.
    *   **Replay Nativo:** O Log Imutável do Kafka permite que os serviços "re-leiam" mensagens ativamente. Isso é essencial num padrão de "Fat Events" se quisermos recriar bancos de dados de leitura (Read Models) do zero.

2.  **Transição para Event Sourced Read Models & Version Tracking:**
    *   Em contraste ao padrão *Inbox/Tabela de Idempotência*, os consumidores downstream poderão adotar *Natural Idempotence* via histórico de Eventos Consumidos (Versionamento/OccurredAt).
    *   **Padrão CQRS:** As projeções (para consultas/read-models da UI) são reconstruídas e alimentadas pelas "versões" lidas do Kafka. Se um consumidor de visão de cardápio ler um evento `MenuUpdated(Version=3)` quando seu estado já processou a versão `4`, ele rejeita / ignora naturalmente esse evento por ser defasado temporalmente, substituindo assim o custoso check transacional.

**Gatilhos práticos para quando focar na mudança:**
*   A gestão de filas/bindings se tornar complexa diante da malha de Sagas, ou gargalo no roteamento do RabbitMQ.
*   Necessidade da recriação e projeção dos bancos de dados visuais dos aplicativos (CQRS real) e dos relatórios de forma determinística ("Como nosso cardápio estava configurado no dia X e por quê?").
*   Queda ou impacto na performance em consumidores devido unicamente à excessiva competição por exclusão em tabelas `inbox` contra o MySQL nas validações de Idempotência.
