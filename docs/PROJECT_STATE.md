# Estado do Projeto (Handoff)

> **Última Atualização:** 2026-04-23
> **Fase Atual:** Fase 5: Estabilização de Infraestrutura e E2E Concluída

## 1. O que foi feito nesta sessão
* **Estabilização de Testes E2E e Correções de Infraestrutura:**
    * **Padronização de Banco de Dados:** Todos os campos de identificadores (`id`, `aggregate_id`, `correlation_id`) foram atualizados para `VARCHAR(255)` em todas as migrations e no `init.sql`. Isso resolveu erros de truncamento em UUIDs que levavam prefixos (ex: `tkt-`, `ev-`).
    * **Correção no Outbox Relay (Restaurant):** Implementado o worker de Outbox no `Restaurant Service` e corrigido erro de conexão prematuramente fechada pelo uso incorreto de `defer publisher.Close()`.
    * **Robustez na Mensageria:** 
        * O `Ordering Service` agora declara as exchanges de comandos (`payment.commands`, etc.) ao iniciar, garantindo que mensagens não sejam perdidas se o serviço de destino ainda não subiu.
        * O consumidor de eventos da Saga foi atualizado para suportar tanto payloads "flat" (enviados pelo Restaurant) quanto "envelopados" (enviados pelo Payment), tornando a desserialização resiliente a diferentes estilos de publicação.
    * **Correção de Roteamento (Delivery):** Ajustadas as `routing_keys` no consumidor do `Delivery Service` para incluir o prefixo `.command.`, permitindo o processamento correto das solicitações de agendamento.
    * **Estabilização de Mocks:** Removida falha aleatória de 10% no `MockGateway` de pagamento para garantir determinismo nos testes automatizados.
* **Ajuste de Testes:**
    * Atualizados os status esperados no `saga_test.go` para refletir a realidade do domínio: `PREPARING` para o fim do fluxo feliz e `CANCELLED` para compensação de rejeição pelo restaurante.

## 2. Contexto Atual (Onde paramos)
Todos os serviços estão operacionais e integrados via Saga:
- [x] Correção de schema (IDs VARCHAR(255)).
- [x] Validação de cada módulo isoladamente via cURL.
- [x] Sucesso total nos testes E2E (`./scripts/run-e2e.sh`).
- [x] Outbox Relay funcional nos serviços críticos.

### Notas Técnicas:
* **Terminal Status (Happy Path):** O pedido agora atinge o status `PREPARING` (após agendamento da entrega).
* **Terminal Status (Compensation):** 
    * Falha no Pagamento -> `REJECTED`.
    * Rejeição pelo Restaurante -> `CANCELLED` (após estorno do pagamento).
* **Infra:** RabbitMQ Management UI exposta em `:15673` no ambiente de teste para depuração.
