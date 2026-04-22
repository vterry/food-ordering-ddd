# Restaurant Service

Microserviço responsável pela gestão de restaurantes, menus e processamento de pedidos (tickets) na cozinha.

## Tecnologias
- **Linguagem:** Go 1.25+
- **Framework Web:** Echo
- **Banco de Dados:** MySQL 8.0
- **Ferramenta de SQL:** SQLC (Type-safe queries)
- **Mensageria:** RabbitMQ (AMQP 0.9.1) - Consome `CheckoutRequested` e publica eventos de ticket.
- **Contrato de API:** OpenAPI 3.0 (oapi-codegen)
- **Comunicação Interna:** gRPC (Catalog Service para outros microserviços)

## Como Executar

### 1. Iniciar Infraestrutura
Certifique-se de ter o Docker instalado e execute:
```bash
docker compose up -d
```

### 2. Rodar Migrações
Execute o comando para criar as tabelas no banco de dados (ajuste o caminho se necessário):
```bash
migrate -path restaurant/internal/adapters/db/migrations -database "mysql://root:root@tcp(localhost:3306)/food_project" up
```

### 3. Iniciar o Servidor
```bash
go run restaurant/cmd/server/main.go
```
O servidor HTTP estará disponível em `http://localhost:8081/api/v1` e o gRPC em `:50052`.

## Exemplos de Teste (CURL)

### 1. Criar Restaurante
```bash
curl -i -X POST http://localhost:8081/api/v1/restaurants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Pizza Palace",
    "address": {
      "street": "Rua das Pizzas, 123",
      "city": "São Paulo",
      "zip_code": "01234-567"
    }
  }'
```
*Anote o ID retornado (ex: `rest-0x...`).*

### 2. Criar Menu
```bash
# Substitua {restaurant_id} pelo ID obtido acima
curl -i -X POST http://localhost:8081/api/v1/restaurants/{restaurant_id}/menus \
  -H "Content-Type: application/json" \
  -d '{"name": "Menu Principal"}'
```
*Anote o ID do menu (ex: `menu-Menu Principal`).*

### 3. Adicionar Item ao Menu
```bash
curl -i -X POST http://localhost:8081/api/v1/restaurants/{restaurant_id}/menus/{menu_id}/items \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Pizza Margherita",
    "description": "Molho de tomate, mussarela e manjericão",
    "price": 45.90,
    "currency": "BRL",
    "category": "Pizzas"
  }'
```

### 4. Ativar Menu
```bash
curl -i -X PUT http://localhost:8081/api/v1/restaurants/{restaurant_id}/menus/{menu_id}/activate
```

### 5. Consultar Cardápio (Público)
```bash
curl -i http://localhost:8081/api/v1/restaurants/{restaurant_id}/menu
```

### 6. Gestão de Tickets (Cozinha)

#### Consultar Ticket
```bash
curl -i http://localhost:8081/api/v1/tickets/{id}
```

#### Confirmar Pedido
```bash
curl -i -X POST http://localhost:8081/api/v1/tickets/{id}/confirm
```

#### Iniciar Preparo
```bash
curl -i -X POST http://localhost:8081/api/v1/tickets/{id}/prepare
```

#### Marcar como Pronto
```bash
curl -i -X POST http://localhost:8081/api/v1/tickets/{id}/ready
```

## Traceability
O serviço propaga o `X-Correlation-ID` em todos os logs estruturados e respostas HTTP, garantindo rastreabilidade ponta-a-ponta entre chamadas gRPC e eventos RabbitMQ.
