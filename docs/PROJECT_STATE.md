# Estado do Projeto (Handoff)

> **Última Atualização:** 2026-04-21
> **Fase Atual:** Fase 2: Restaurant Service (Concluída) -> Fase 3: Payment Service (Início)

## 1. O que foi feito nesta sessão
* **Melhorias Arquiteturais e Padronização:**
    * **Logging Estruturado:** Implementado `slog` em toda a stack, com adapter customizado para o framework Echo, garantindo logs em JSON para observabilidade.
    * **Tratamento de Erros Polimórfico:** Refatorado `AppError` para incluir o status HTTP e slug de erro. Removido o `switch` centralizado no handler, seguindo um padrão mais limpo e extensível.
    * **Desacoplamento de Framework:** O pacote `@common` foi limpo de referências ao Echo. Configurações de servidor e middlewares foram movidas para os adaptadores locais de cada microserviço.
    * **Traceability:** Garantida a propagação do `X-Correlation-ID` entre chamadas HTTP, gRPC e eventos RabbitMQ.
* **Revisão e Refino (Customer & Restaurant):**
    * **Resiliência:** Implementado `Circuit Breaker` (sony/gobreaker) no cliente gRPC do `Customer Service` para chamadas ao catálogo de restaurantes.
    * **Geração de IDs:** Corrigida a geração de IDs no `Restaurant Service` para utilizar UUIDs reais em vez de ponteiros de memória.
    * **Encapsulamento DDD:** Refatorado agregado `Ticket` para proteger o status, utilizando factory `MapFromPersistence` para reconstrução via repositório.
* **Fase 3 (Payment Service) Finalizada:**
    * **API, Persistência & Messaging Concluídas:** Pagamento possui domínio rico, repositório transacional com Outbox e consumidor de comandos RabbitMQ.

## 2. Contexto Atual (Onde paramos)
Avance para a Fase 3 (Payment Service):
- [x] Definir o Modelo de Domínio do serviço de Pagamento (Agregado `Payment`).
- [x] Implementar a Máquina de Estados de Pagamento (`CREATED`, `AUTHORIZED`, `CAPTURED`, `REFUNDED`).
- [x] Implementar o `Mock Payment Gateway` para simular integrações externas com Circuit Breaker.
- [x] Configurar persistência e mensageria para o serviço de Pagamento.

### Notas Técnicas:
* **Customer Service:** HTTP `:8080`, gRPC `:50051`.
* **Restaurant Service:** HTTP `:8081`, gRPC `:50052`.
* O padrão Outbox está implementado nos repositórios, mas o worker (poller) global será consolidado na Fase 6.
