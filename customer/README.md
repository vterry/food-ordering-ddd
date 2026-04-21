# Customer Service

Microserviço responsável pela gestão de clientes, endereços e carrinhos de compras da plataforma de pedidos de comida.

## Tecnologias
- **Linguagem:** Go 1.25+
- **Framework Web:** Echo
- **Banco de Dados:** MySQL 8.0
- **Ferramenta de SQL:** SQLC (Type-safe queries)
- **Mensageria:** RabbitMQ (AMQP 0.9.1)
- **Contrato de API:** OpenAPI 3.0 (oapi-codegen)

## Como Executar

### 1. Iniciar Infraestrutura
Certifique-se de ter o Docker instalado e execute:
```bash
docker compose up -d
```

### 2. Rodar Migrações
Execute o comando para criar as tabelas no banco de dados:
```bash
task migrate-up
```

### 3. Iniciar o Servidor
```bash
go run customer/cmd/server/main.go
```
O servidor estará disponível em `http://localhost:8080/api/v1`.

## Exemplos de Teste (CURL)

### 1. Registrar Cliente
```bash
curl -i -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com", "phone": "123456789"}'
```
*Anote o ID retornado na resposta para usar nos comandos abaixo.*

### 2. Adicionar Endereço
```bash
# Substitua {id} pelo ID do cliente
curl -i -X POST http://localhost:8080/api/v1/customers/{id}/addresses \
  -H "Content-Type: application/json" \
  -d '{"street": "Rua das Flores, 123", "city": "São Paulo", "zip_code": "01234-567", "is_default": true}'
```

### 3. Consultar Cliente
```bash
curl -i http://localhost:8080/api/v1/customers/{id}
```

### 4. Gestão de Carrinho

#### Adicionar Item
```bash
curl -i -X POST http://localhost:8080/api/v1/customers/{id}/cart/items \
  -H "Content-Type: application/json" \
  -d '{
    "restaurant_id": "rest-001",
    "product_id": "prod-123",
    "name": "Hambúrguer Gourmet",
    "price": 35.50,
    "currency": "BRL",
    "quantity": 2,
    "observation": "Sem cebola"
  }'
```

#### Consultar Carrinho
```bash
curl -i http://localhost:8080/api/v1/customers/{id}/cart
```

#### Atualizar Quantidade de Item
```bash
curl -i -X PUT http://localhost:8080/api/v1/customers/{id}/cart/items/prod-123 \
  -H "Content-Type: application/json" \
  -d '{"quantity": 1}'
```

#### Remover Item do Carrinho
```bash
curl -i -X DELETE http://localhost:8080/api/v1/customers/{id}/cart/items/prod-123
```

### 5. Checkout
```bash
curl -i -X POST http://localhost:8080/api/v1/customers/{id}/cart/checkout
```
*Este comando gera um evento `CheckoutRequested` no RabbitMQ.*

## Rastreabilidade (Traceability)
Todas as requisições aceitam o header `X-Correlation-ID`. Caso não seja enviado, o sistema gera um automaticamente.
Exemplo:
```bash
curl -i -H "X-Correlation-ID: meu-rastreio-unico" http://localhost:8080/api/v1/customers/{id}
```
