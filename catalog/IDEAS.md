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
