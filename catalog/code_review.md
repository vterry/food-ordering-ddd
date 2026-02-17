# Code Review — Catalog Module (Revisão 2)

**Data:** 2026-02-17
**Nota geral:** 9/10

## Status das correções anteriores

- [x] Violação de Demeter em `Menu` -> `Category` (`AddItemToCategory` agora delega para `Category.AddItem`)
- [x] Vazamento de Transação (panic safety) em `UnitOfWork` (`defer tx.Rollback()` adicionado)
- [x] Falta de Atomicidade nos AppServices (todos os `Save` agora usam `uow.Run`)
- [x] Status Codes HTTP corrigidos (GET=200, POST=201, PATCH/PUT=200)
- [x] Mapeamento de erros de domínio para 422 adicionado
- [x] Typo na rota `categorie` -> `categories`
- [x] `hasActiveItems` renomeado para `hasItems`
- [x] `CanAcceptOrder` simplificado

---

## Problemas remanescentes

### 1. Violações da Lei de Demeter

#### 1.1 [MÉDIA] Mappers com cadeias de 3 niveis — `mappers.go`

**Problema:**

As cadeias `.ID().String()` atravessam a fronteira do Value Object embeddado, expondo a estrutura interna de `BaseID`:

| Linha | Expressao | Niveis |
|-------|-----------|--------|
| 14 | `restaurant.RestaurantID.ID().String()` | 3 |
| 18 | `restaurant.ActiveMenuID().ID().String()` | 3 |
| 32 | `menu.MenuID.ID().String()` | 3 |
| 34 | `menu.RestaurantID().ID().String()` | 3 |
| 62 | `category.ID().String()` | 2 |
| 70 | `item.ItemID.ID().String()` | 3 |

Os Value Objects de ID ja possuem `String()` definido (ex: `RestaurantID.String()` em `restaurant_id.go:18` retorna `r.ID().String()`). Como os agregados usam embedding, o Go promove esse metodo automaticamente.

**Recomendacao:**

Substituir `.ID().String()` por `.String()` em todos os casos onde o embedding ja promove o metodo:

```go
// antes (3 niveis)
restaurant.RestaurantID.ID().String()

// depois (1 nivel — promotion via embedding)
restaurant.String()
```

Para `ActiveMenuID` e `RestaurantID()` (retornados por getter, sem embedding):

```go
// antes
restaurant.ActiveMenuID().ID().String()

// depois — o getter retorna MenuID que ja tem String()
restaurant.ActiveMenuID().String()
```

---

#### 1.2 [BAIXA] Events acessam campos embeddados explicitamente — `events.go` (menu)

**Problema:**

Em `NewMenuActivated` (`events.go:34`), `NewItemAddedToCategory` (`events.go:165`), `NewMenuCategoryAdded` (`events.go:182-192`):

```go
item.ItemID.String()       // acessa campo embeddado explicitamente
cat.CategoryID.String()    // idem
```

Como `ItemMenu` embute `ItemID` e `Category` embute `CategoryID`, a promotion de Go ja disponibiliza `String()` diretamente.

**Recomendacao:**

Padronizar para `item.String()` e `cat.String()` via promotion. Isso elimina a dependencia do nome do campo embeddado.

---

#### 1.3 [BAIXA] Repositorios acessam campos embeddados — `menu_repository.go` e `restaurant_repository.go`

**Problema:**

Em `menu_repository.go:60`:
```go
cat.CategoryID.String()
```

Em `menu_repository.go:74`:
```go
item.ItemID.String()
```

Em `restaurant_repository.go:39`:
```go
agg.ID().String()  // atravessa BaseID
```

**Recomendacao:**

Mesma correcao: usar `cat.String()`, `item.String()`, `agg.String()` via promotion.

---

### 2. Rollback redundante no UnitOfWork

#### 2.1 [BAIXA] Duplo Rollback — `unit_of_work.go:31-36`

**Problema:**

```go
func (u *UnitOfWork) Run(ctx context.Context, fn func(ctx context.Context) error) error {
    tx, err := u.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()          // <- rollback via defer (correto)

    ctxWithTx := context.WithValue(ctx, txKey{}, tx)

    if err := fn(ctxWithTx); err != nil {
        tx.Rollback()            // <- rollback explicito (redundante)
        return err
    }

    return tx.Commit()
}
```

O `defer tx.Rollback()` ja garante rollback em caso de erro ou panic. O `tx.Rollback()` explicito na linha 36 e redundante — ele funciona, mas adiciona ruido. O `defer` sozinho ja cobre todos os cenarios.

**Recomendacao:**

Remover o rollback explicito e confiar apenas no `defer`:

```go
// pseudocodigo
defer tx.Rollback()

if err := fn(ctxWithTx); err != nil {
    return err  // defer faz o rollback
}
return tx.Commit()  // apos commit, defer rollback e no-op
```

---

### 3. Adapter importa dominio diretamente — `utils.go`

#### 3.1 [MEDIA] Handler HTTP importa pacotes de dominio — `utils.go:11-13`

**Problema:**

```go
import (
    "github.com/vterry/food-ordering/catalog/internal/core/domain/menu"
    "github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
    "github.com/vterry/food-ordering/catalog/internal/core/ports/output"
)
```

O adapter de entrada (`adapters/input/rest/utils.go`) importa diretamente os pacotes de dominio `menu` e `restaurant` para referenciar sentinels como `menu.ErrMenuNotEditable`, `restaurant.ErrAlreadyOpened`, etc. Isso cria um acoplamento direto entre a camada HTTP e o dominio, quebrando a indireção que a Arquitetura Hexagonal propoe.

Na arquitetura atual, os handlers deveriam depender apenas dos ports (`ports/input` e `ports/output`) e nunca do dominio diretamente.

**Recomendacao:**

Duas abordagens possiveis:

1. **Erros tipados no dominio**: Criar um tipo `DomainError` com codigo semantico. A camada de app propaga esses erros, e o handler faz switch no tipo sem conhecer cada sentinel.

2. **Re-exportar via ports**: Definir os erros de dominio relevantes como parte do contrato de `ports/output` (ou criar um `ports/errors.go`), para que o adapter dependa apenas dos ports.

Pergunta guia: Se amanha voce criar um novo bounded context que reutiliza os mesmos handlers HTTP, que imports ele precisaria? Quanto mais ele depender de ports (contratos) em vez de dominio (implementacao), mais desacoplado sera.

---

### 4. Validacao dos DTOs — tags `binding` vs `validate`

#### 4.1 [MEDIA] Tags de validacao incompativeis com o validator — `dtos.go`

**Problema:**

Os DTOs em `ports/input/dtos.go` usam a tag `binding`:

```go
type AddItemRequest struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description" binding:"required"`
    PriceCents  int64  `json:"price_cents" binding:"required,gt=0"`
}
```

Porem, o validator usado em `utils.go:15` e o `go-playground/validator`, que reconhece a tag `validate`, nao `binding`. A tag `binding` e especifica do Gin framework. Como o projeto usa `net/http` puro (sem Gin), as validacoes **nao estao sendo executadas**.

`Validate.Struct(payload)` nao vai encontrar regras na tag `binding` — vai sempre retornar `nil`, aceitando payloads invalidos silenciosamente.

**Recomendacao:**

Trocar todas as tags `binding` por `validate`:

```go
// antes
Name string `json:"name" binding:"required"`

// depois
Name string `json:"name" validate:"required"`
```

Pergunta guia: Voce ja testou enviar um request sem o campo `name`? O que acontece? Se o validator nao reclama, a validacao nao esta funcionando.

---

### 5. `handleInputValidation` — branch morto

#### 5.1 [BAIXA] Logica redundante — `utils.go:39-44`

**Problema:**

```go
if err := Validate.Struct(payload); err != nil {
    var validationErrors validator.ValidationErrors
    if errors.As(err, &validationErrors) {
        return err           // retorna err
    }
    return err               // retorna err tambem
}
```

Ambos os branches retornam `err`. O `errors.As` nao altera o comportamento. O check interno e dead code.

**Recomendacao:**

Simplificar:

```go
if err := Validate.Struct(payload); err != nil {
    return err
}
```

---

### 6. Erros silenciados nos repositorios

#### 6.1 [MEDIA] Parse errors ignorados — `menu_repository.go`

**Problema:**

Em `scanMenuRows` (`menu_repository.go:174, 195-196`) e `processChildren` (`menu_repository.go:213, 219, 221`):

```go
menuStatus, _ := enums.ParseMenuStatus(row.MenuStatus)    // erro ignorado
mID, _ := valueobjects.ParseMenuId(uuid)                   // erro ignorado
rID, _ := valueobjects.ParseRestaurantId(b.RestaurantID)   // erro ignorado
cid, _ := valueobjects.ParseCategoryId(catUUID)            // erro ignorado
iid, _ := valueobjects.ParseItemId(*row.ItemUUID)          // erro ignorado
status, _ := enums.ParseItemStatus(*row.ItemStatus)         // erro ignorado
```

Se o banco tiver dados corrompidos ou invalidos, esses erros sao engolidos silenciosamente, resultando em agregados com valores zero. Isso pode causar comportamento inesperado downstream sem nenhum log ou indicacao do problema.

**Recomendacao:**

Propagar os erros. Se o dado veio do banco e nao parseia, e um problema serio que deve ser visivel:

```go
// pseudocodigo
menuStatus, err := enums.ParseMenuStatus(row.MenuStatus)
if err != nil {
    return nil, fmt.Errorf("corrupted menu status in DB: %w", err)
}
```

---

### 7. `Category.AddItem` retorna eventos do item e altera assinatura

#### 7.1 [BAIXA] Efeito colateral na refatoracao — `category.go:39-53`

**Problema:**

```go
func (c *Category) AddItem(item ItemMenu) ([]common.DomainEvent, error) {
    // ...
    c.items = append(c.items, item)
    return item.PullEvent(), nil
}
```

A refatoracao (correta) de retornar eventos coletados tem um efeito colateral: `PullEvent()` e chamado no parametro `item` (copia por valor), nao no item que foi adicionado ao slice. Como `ItemMenu` e passado por valor, o `PullEvent()` drena os eventos da **copia**, enquanto o item dentro de `c.items` ainda mantem seus eventos.

Isso funciona corretamente **porque a copia acontece antes do append** — mas e um detalhe sutil. Se um dia `ItemMenu` passar a ser por ponteiro, o comportamento muda.

**Recomendacao:**

Documentar essa decisao com um comentario breve ou considerar chamar `PullEvent()` no item ja armazenado:

```go
// pseudocodigo
c.items = append(c.items, item)
events := c.items[len(c.items)-1].PullEvent()  // drena do item armazenado
return events, nil
```

---

### 8. Observacoes menores

| Severidade | Arquivo | Problema | Recomendacao |
|:----------:|---------|----------|--------------|
| BAIXA | `utils.go:17` | Variavel global `Validate` dificulta testes isolados | Injetar como dependencia ou criar via factory nos handlers |
| BAIXA | `utils.go:76` | Log de erro interno sem contexto de request | Incluir metodo HTTP, path e remote addr: `slog.Error("internal error", "err", err, "method", req.Method, "path", req.URL.Path)` |
| BAIXA | `menu_handler.go:102` | Nome do handler `handleAddMenuCategorie` manteve o typo (singular frances) | Renomear para `handleAddCategory` |
| BAIXA | `menu_handler.go:41,61` | Variavel `restauntIdStr` com typo (falta o 'r' de restaurant) | Renomear para `restaurantIdStr` |
| BAIXA | `restaurant/events.go:23-26` | `NewRestaurantCreated` acessa `restaurant.address` (campo privado) diretamente em vez de usar `restaurant.Address()` | Usar o getter `Address()` por consistencia com o resto do codigo |
| BAIXA | `menu_repository.go:70` | `LastInsertId()` ignora erro com `_` | Verificar o erro: `catDbID, err := res.LastInsertId()` |

---

## Pontos fortes

- **Outbox Pattern** corretamente implementado — agregado + eventos na mesma transacao atomica
- **Specification Pattern** com composicao via `And`
- **Value Objects** com encapsulamento (campos privados, factory com validacao)
- **Compile-time assertions** em todos os App Services
- **Domain Events fat** com snapshots — suficientes para consumidores downstream
- **Graceful shutdown** com signal handling e timeout no `main.go`
- **Segregacao de interfaces** nos output ports — interfaces pequenas e focadas
- **Padrao Restore** para hidratacao sem passar pela validacao de criacao
- **Domain Service** (`MenuAssignmentService`) corretamente stateless e validando ownership
- **UnitOfWork** com panic safety via `defer tx.Rollback()`
- **Mapeamento de erros HTTP** cobrindo erros de dominio com 422
- **Middleware chain** limpa e composavel
