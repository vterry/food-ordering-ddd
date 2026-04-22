# Estado do Projeto (Handoff)

> **Última Atualização:** 2026-04-21
> **Fase Atual:** Fase 2: Restaurant Service (Concluída) -> Fase 3: Payment Service (Início)

## 1. O que foi feito nesta sessão
* **Melhorias Arquiteturais e Padronização:**
    * **Logging Estruturado:** Implementado `slog` em toda a stack, com adapter customizado para o framework Echo, garantindo logs em JSON para observabilidade.
    * **Tratamento de Erros Polimórfico:** Refatorado `AppError` para incluir o status HTTP e slug de erro. Removido o `switch` centralizado no handler, seguindo um padrão mais limpo e extensível.
    * **Desacoplamento de Framework:** O pacote `@common` foi limpo de referências ao Echo. Configurações de servidor e middlewares foram movidas para os adaptadores locais de cada microserviço.
    * **Traceability:** Garantida a propagação do `X-Correlation-ID` entre chamadas HTTP, gRPC e eventos RabbitMQ.
* **Fase 2 (Restaurant Service) Finalizada:**
    * **API & Messaging Concluídas:** Operações administrativas, cardápio público e gestão de tickets totalmente operacionais.
    * **Comunicação Cross-Service:** Validada integração gRPC onde o `Customer Service` consome o catálogo do `Restaurant Service`.
    * **Documentação:** Criado `README.md` detalhado para o serviço de Restaurante com exemplos de uso.

## 2. Contexto Atual (Onde paramos)
Avance para a Fase 3 (Payment Service):
- [ ] Definir o Modelo de Domínio do serviço de Pagamento (Agregado `Payment`).
- [ ] Implementar a Máquina de Estados de Pagamento (`CREATED`, `AUTHORIZED`, `CAPTURED`, `REFUNDED`).
- [ ] Implementar o `Mock Payment Gateway` para simular integrações externas.
- [ ] Configurar persistência e mensageria para o serviço de Pagamento.

### Notas Técnicas:
* **Customer Service:** HTTP `:8080`, gRPC `:50051`.
* **Restaurant Service:** HTTP `:8081`, gRPC `:50052`.
* O padrão Outbox está implementado nos repositórios, mas o worker (poller) global será consolidado na Fase 6.
