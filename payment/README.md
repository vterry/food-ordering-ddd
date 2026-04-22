# Payment Service

Microserviço responsável pela orquestração financeira dos pedidos, incluindo autorização, captura, estorno e liberação de pagamentos de forma resiliente e idempotente.

## Tecnologias
- **Linguagem:** Go 1.25+
- **Framework Web:** Echo
- **Banco de Dados:** MySQL 8.0
- **Ferramenta de SQL:** SQLC (Type-safe queries)
- **Mensageria:** RabbitMQ (AMQP 0.9.1) - Consome comandos de pagamento do orquestrador de Sagas.
- **Resiliência:** Circuit Breaker (sony/gobreaker) para proteger integrações com gateways externos.

## Como Executar

### 1. Iniciar Infraestrutura
Certifique-se de ter o Docker instalado e execute na raiz do projeto:
```bash
docker compose up -d
```

### 2. Rodar Migrações
Execute o comando para criar as tabelas no banco de dados (ajuste o DSN se necessário):
```bash
# Exemplo usando golang-migrate ou similar, ou manualmente no MySQL:
# As migrations estão em: payment/internal/adapters/db/migrations/
```

### 3. Iniciar o Servidor
```bash
go run payment/cmd/server/main.go
```
O servidor HTTP estará disponível em `http://localhost:8082`.

## API HTTP (Endpoints)

### 1. Consultar Pagamento
```bash
curl -i http://localhost:8082/api/v1/payments/{payment_id}
```

### 2. Health Checks
```bash
curl -i http://localhost:8082/health/live
curl -i http://localhost:8082/health/ready
```

## Mensageria (RabbitMQ)

O serviço atua principalmente de forma reativa a comandos enviados pelo `Ordering Service` (Saga Orchestrator) na exchange `payment.commands`.

### Comandos Suportados:
- `payment.command.authorize`: Inicia um novo pagamento e tenta autorizar no gateway.
- `payment.command.capture`: Efetiva a cobrança de um pagamento previamente autorizado.
- `payment.command.refund`: Realiza o estorno de um pagamento capturado.
- `payment.command.release`: Libera a reserva de um pagamento autorizado (void).

### Payload de Exemplo (Authorize):
```json
{
  "order_id": "uuid-do-pedido",
  "amount": 5000,
  "card_token": "token-seguro-do-cartao"
}
```

## Resiliência e Idempotência
- **Circuit Breaker:** Caso o gateway de pagamento apresente falhas consecutivas ou alta latência, o circuito abre para proteger o sistema.
- **Idempotência:** As operações de domínio são protegidas contra execuções duplicadas (ex: chamar `Capture` em um pagamento já capturado retornará um erro sem duplicar a transação).
- **Transactional Outbox:** O estado do pagamento e seus eventos de domínio são salvos na mesma transação SQL, garantindo que o orquestrador sempre seja notificado do resultado.

## Rastreabilidade (Traceability)
O serviço propaga o `X-Correlation-ID` em todas as requisições HTTP e mensagens RabbitMQ para permitir o rastreio completo do fluxo de pagamento através dos microserviços.
