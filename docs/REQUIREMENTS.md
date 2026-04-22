# Documento de Requisitos — Food Ordering Platform

> **Versão:** 1.0  
> **Data:** 2026-04-20  
> **Status:** Em Revisão

---

## 1. Visão Geral do Produto

O **Food Ordering Platform** é um sistema de backend para um aplicativo de delivery de comida. O objetivo principal é servir como plataforma de estudo e aplicação prática de conceitos avançados de arquitetura distribuída, utilizando **Domain-Driven Design (DDD)**, **Event-Driven Architecture (EDA)** e padrões de resiliência para sistemas distribuídos.

### 1.1 Literatura de Referência

| Livro | Autor | Foco |
|-------|-------|------|
| *Software Architecture: The Hard Parts* | Neal Ford et al. | Trade-offs e granularidade de serviços |
| *Building Microservices* | Sam Newman | Acoplamento, coesão e decomposição |
| *Building Event-Driven Microservices* | Adam Bellemare | Topologias de eventos e schemas |

### 1.2 Stack Tecnológica

| Componente | Tecnologia | Justificativa |
|------------|-----------|---------------|
| Linguagem | **Go** | Performance, concorrência nativa, tipagem estática |
| Banco de Dados | **MySQL** | ACID, maturidade, ecossistema consolidado |
| Geração de Código SQL | **SQLC** | Type-safety, geração de código Go a partir de queries SQL |
| Comunicação Síncrona | **gRPC** | Performance, contratos tipados via Protobuf |
| Interface HTTP | **Biblioteca padrão do Go** (`net/http`) | Simplicidade, sem dependências externas |
| Mensageria | **RabbitMQ** | Pub/Sub, roteamento flexível, suporte a filas de dead-letter |

---

## 2. Atores do Sistema

| Ator | Descrição |
|------|-----------|
| **Cliente (Usuário)** | Pessoa que utiliza o aplicativo para buscar restaurantes, montar pedidos e realizar compras |
| **Restaurante** | Estabelecimento que gerencia cardápios, recebe e processa pedidos |
| **Sistema (Plataforma)** | Orquestrador que coordena o fluxo entre pagamento, restaurante e entrega |
| **Entregador** | Pessoa responsável pela entrega do pedido ao cliente (ator implícito nos requisitos de delivery) |

---

## 3. Requisitos Funcionais

### 3.1 Domínio do Cliente (Customer)

#### RF-C01 — Cadastro e Autenticação

| ID | Requisito | Prioridade |
|----|-----------|------------|
| RF-C01.1 | O sistema deve permitir que o cliente se cadastre informando nome, e-mail, telefone e endereço de entrega | Alta |
| RF-C01.2 | O sistema deve autenticar o cliente antes de permitir operações de pedido e pagamento | Alta |
| RF-C01.3 | O cliente deve poder gerenciar seus endereços de entrega (adicionar, editar, remover) | Média |

#### RF-C02 — Interação com Restaurantes

| ID | Requisito | Prioridade |
|----|-----------|------------|
| RF-C02.1 | O cliente deve ser capaz de listar os restaurantes disponíveis na plataforma | Alta |
| RF-C02.2 | Para cada restaurante, o cliente deve ser capaz de visualizar o cardápio ativo com os itens disponíveis (nome, descrição, preço e imagem) | Alta |
| RF-C02.3 | O cliente deve ser capaz de adicionar itens do restaurante ao carrinho de compras | Alta |
| RF-C02.4 | O carrinho de compras deve aceitar apenas itens de um **único restaurante por vez**. Caso o cliente tente adicionar um item de outro restaurante, o sistema deve retornar um erro informativo | Alta |
| RF-C02.5 | O cliente deve ser capaz de incluir **observações por item** adicionado ao carrinho (ex.: "sem cebola", "ponto da carne mal passado"). A observação é vinculada ao item, não ao carrinho | Média |
| RF-C02.6 | O cliente deve ser capaz de alterar a quantidade de itens no carrinho | Média |
| RF-C02.7 | O cliente deve ser capaz de remover itens do carrinho | Média |

#### RF-C03 — Pagamento

| ID | Requisito | Prioridade |
|----|-----------|------------|
| RF-C03.1 | O sistema deve permitir pagamento **exclusivamente via cartão de crédito** | Alta |
| RF-C03.2 | O cliente deve ser capaz de cadastrar e gerenciar seus cartões de crédito na plataforma | Alta |
| RF-C03.3 | O sistema deve utilizar o modelo de **autorização e captura** (two-phase payment): primeiro reservar o valor no limite do cartão, depois capturar após confirmação do restaurante | Alta |

#### RF-C04 — Gestão de Pedidos (Visão do Cliente)

| ID | Requisito | Prioridade |
|----|-----------|------------|
| RF-C04.1 | O cliente deve ser capaz de acompanhar o status do pedido em tempo real | Alta |
| RF-C04.2 | O cliente pode cancelar o pedido **enquanto estiver na fase de confirmação** (antes da confirmação do restaurante). Após a confirmação, o cancelamento não é permitido pelo cliente | Alta |
| RF-C04.3 | Após a captura do pagamento, o pedido **não pode ser modificado** pelo cliente | Alta |
| RF-C04.4 | O cliente deve visualizar o histórico de pedidos realizados | Média |

#### RF-C05 — Entrega (Visão do Cliente)

| ID | Requisito | Prioridade |
|----|-----------|------------|
| RF-C05.1 | O cliente pode **recusar o recebimento** do pedido no momento da entrega | Alta |
| RF-C05.2 | Se o cliente recusar o recebimento, o sistema deve **cancelar o pedido** e **solicitar o reembolso** à operadora do cartão de crédito | Alta |
| RF-C05.3 | O cliente deve ser capaz de visualizar informações do entregador (nome e estimativa de chegada) | Baixa |

---

### 3.2 Domínio do Restaurante (Restaurant)

#### RF-R01 — Gestão de Cardápio

| ID | Requisito | Prioridade |
|----|-----------|------------|
| RF-R01.1 | O restaurante pode possuir **múltiplos cardápios**, porém apenas um pode estar **ativo** na plataforma em dado momento | Alta |
| RF-R01.2 | O restaurante deve ser capaz de **criar, editar e excluir** itens de um cardápio | Alta |
| RF-R01.3 | O restaurante deve ser capaz de **alternar o cardápio ativo** (trocar qual cardápio é exibido na plataforma) | Alta |
| RF-R01.4 | Cada item do cardápio deve conter: nome, descrição, preço, categoria e status de disponibilidade | Alta |
| RF-R01.5 | O restaurante deve ser capaz de marcar itens individuais como **indisponíveis** sem remover do cardápio | Média |

#### RF-R02 — Gestão de Pedidos (Visão do Restaurante)

| ID | Requisito | Prioridade |
|----|-----------|------------|
| RF-R02.1 | O restaurante deve visualizar a lista de pedidos recebidos para seu estabelecimento | Alta |
| RF-R02.2 | O restaurante deve ser capaz de **confirmar** um pedido, indicando que pode atendê-lo | Alta |
| RF-R02.3 | O restaurante deve ser capaz de **recusar** um pedido, indicando que não pode atendê-lo (com motivo opcional) | Alta |
| RF-R02.4 | O restaurante deve ser capaz de sinalizar que o pedido está **PRONTO** para retirada pelo entregador | Alta |
| RF-R02.5 | O restaurante deve visualizar o histórico de pedidos processados | Média |

---

### 3.3 Regras de Negócio do Sistema (Orquestração)

#### RF-S01 — Fluxo de Criação de Pedido

| ID | Requisito | Prioridade |
|----|-----------|------------|
| RF-S01.1 | Ao criar um pedido, o sistema deve primeiro solicitar a **autorização (reserva)** do pagamento no cartão de crédito | Alta |
| RF-S01.2 | Se a autorização do pagamento **falhar**, o pedido deve ser **rejeitado** imediatamente e o cliente notificado | Alta |
| RF-S01.3 | Se a autorização do pagamento for **bem-sucedida**, o sistema deve enviar o pedido ao restaurante para confirmação | Alta |

#### RF-S02 — Fluxo de Confirmação do Restaurante

| ID | Requisito | Prioridade |
|----|-----------|------------|
| RF-S02.1 | Se o restaurante **confirmar** o pedido, o sistema deve efetivar a **captura do pagamento** | Alta |
| RF-S02.2 | Se o restaurante **recusar** o pedido, o sistema deve **liberar o valor reservado** no cartão e **cancelar** o pedido | Alta |
| RF-S02.3 | Após a captura do pagamento, o sistema deve acionar o módulo de delivery para designar um entregador | Alta |

#### RF-S03 — Fluxo de Entrega

| ID | Requisito | Prioridade |
|----|-----------|------------|
| RF-S03.1 | Quando o restaurante sinalizar que o pedido está pronto, o sistema deve notificar o entregador designado | Alta |
| RF-S03.2 | O entregador deve ser capaz de sinalizar à plataforma que **coletou** o pedido | Média |
| RF-S03.3 | O entregador deve ser capaz de sinalizar à plataforma que **entregou** o pedido ao cliente | Alta |
| RF-S03.4 | Se o cliente recusar o recebimento, o sistema deve iniciar o fluxo de **compensação**: cancelar o pedido e solicitar reembolso | Alta |

#### RF-S04 — Fluxo de Cancelamento e Compensação

| ID | Requisito | Prioridade |
|----|-----------|------------|
| RF-S04.1 | Cancelamento na fase de confirmação (pré-confirmação do restaurante): liberar a autorização do pagamento | Alta |
| RF-S04.2 | Cancelamento por recusa do restaurante: liberar a autorização do pagamento | Alta |
| RF-S04.3 | Cancelamento por recusa na entrega: solicitar reembolso do valor capturado | Alta |
| RF-S04.4 | Todas as compensações devem ser **idempotentes** para lidar com retentativas | Alta |

---

## 4. Requisitos Não-Funcionais

### 4.1 Performance

| ID | Requisito |
|----|-----------|
| RNF-01 | Operações síncronas (consulta de cardápio, listagem de restaurantes) devem responder em menos de **200ms** (p95) |
| RNF-02 | Operações assíncronas (criação de pedido, pagamento) devem ser processadas em menos de **5s** end-to-end |

### 4.2 Resiliência e Confiabilidade

| ID | Requisito |
|----|-----------|
| RNF-03 | Comunicação entre serviços via mensageria deve garantir **at-least-once delivery** |
| RNF-04 | Consumidores de mensagens devem ser **idempotentes** |
| RNF-05 | O sistema deve implementar **circuit breaker** para chamadas a serviços externos (gateway de pagamento) |
| RNF-06 | O sistema deve utilizar **dead-letter queues** para mensagens que falharem após retentativas |
| RNF-07 | Transações distribuídas devem ser gerenciadas via **Saga Pattern** com compensação explícita |

### 4.3 Consistência de Dados

| ID | Requisito |
|----|-----------|
| RNF-08 | Cada serviço deve possuir seu **próprio banco de dados** (database-per-service) |
| RNF-09 | Operações que cruzam bounded contexts devem trabalhar com **consistência eventual** |
| RNF-10 | Eventos de domínio devem ser publicados via **Outbox Pattern** para garantir consistência entre estado e eventos |

### 4.4 Observabilidade

| ID | Requisito |
|----|-----------|
| RNF-11 | Toda requisição deve carregar um **correlation ID** propagado entre todos os serviços |
| RNF-12 | Logs devem ser centralizados e estruturados (JSON) |
| RNF-13 | Métricas de negócio devem ser expostas (pedidos/min, taxa de recusa, tempo médio de preparo) |

### 4.5 Segurança

| ID | Requisito |
|----|-----------|
| RNF-14 | Dados de cartão de crédito **não devem** ser armazenados pela plataforma; utilizar tokenização via gateway de pagamento |
| RNF-15 | Comunicação entre serviços internos deve utilizar **mTLS** ou autenticação via token |

### 4.6 Manutenibilidade

| ID | Requisito |
|----|-----------|
| RNF-16 | O código deve seguir os princípios de **Clean Architecture** / **Hexagonal Architecture** |
| RNF-17 | Contratos de API devem ser versionados |
| RNF-18 | Migrações de banco de dados devem ser versionadas e reprodutíveis |
| RNF-19 | O uso de **TDD (Test-Driven Development)** é obrigatório em todos os módulos, garantindo cobertura de testes desde o início do desenvolvimento |


---

## 5. Diagrama de Estados do Pedido

```
┌─────────────┐
│   CREATED   │ ← Pedido criado pelo cliente
└──────┬──────┘
       │
       ▼
┌──────────────────┐    Falha na autorização    ┌────────────┐
│ PAYMENT_PENDING  │ ─────────────────────────► │  REJECTED  │
└──────┬───────────┘                            └────────────┘
       │ Autorização OK
       ▼
┌──────────────────────┐ Cancelado pelo cliente  ┌────────────────┐
│ AWAITING_RESTAURANT  │ ──────────────────────► │   CANCELLED    │
│    _CONFIRMATION     │                         │ (libera reserva│
└──────┬───────────────┘                         └────────────────┘
       │
       ├── Restaurante recusa ──────────────────► CANCELLED (libera reserva)
       │
       ▼ Restaurante confirma
┌─────────────────┐
│   CONFIRMED     │ ← Captura do pagamento efetuada
└──────┬──────────┘
       │
       ▼
┌─────────────────┐
│   PREPARING     │ ← Restaurante está preparando
└──────┬──────────┘
       │
       ▼
┌──────────────┐
│    READY     │ ← Pedido pronto para coleta
└──────┬───────┘
       │
       ▼
┌──────────────────┐
│  OUT_FOR_DELIVERY│ ← Entregador coletou
└──────┬───────────┘
       │
       ├── Cliente recusa ──────────────────────► CANCELLED (reembolso)
       │
       ▼ Cliente aceita
┌──────────────┐
│  DELIVERED   │ ← Pedido entregue com sucesso
└──────────────┘
```

---

## 6. Glossário

| Termo | Definição |
|-------|-----------|
| **Autorização** | Reserva de um valor no limite do cartão de crédito sem efetivar a cobrança |
| **Captura** | Efetivação da cobrança de um valor previamente autorizado |
| **Reembolso (Refund)** | Devolução total ou parcial de um valor previamente capturado |
| **Bounded Context** | Fronteira de um modelo de domínio onde termos e regras possuem significado específico |
| **Saga** | Padrão para gerenciar transações distribuídas por meio de sequências de transações locais com compensação |
| **Outbox Pattern** | Padrão que garante publicação de eventos em conjunto com mudança de estado, usando uma tabela auxiliar e polling |
| **Idempotência** | Propriedade de uma operação que produz o mesmo resultado mesmo quando executada múltiplas vezes |
| **Dead-Letter Queue (DLQ)** | Fila especial para mensagens que falharam após todas as tentativas de processamento |
| **Correlation ID** | Identificador único propagado entre serviços para rastreamento de uma requisição end-to-end |

---

## 7. Rastreabilidade de Requisitos

A tabela abaixo mapeia os requisitos originais do prompt para os requisitos formais deste documento:

| Requisito Original | Requisito Formal | Status |
|--------------------|------------------|--------|
| RQ1.1 (Usuário - Listar restaurantes) | RF-C02.1 | ✅ Mantido |
| RQ1.1 (Usuário - Visualizar cardápio) | RF-C02.2 | ✅ Mantido |
| RQ1.2 (Usuário - Adicionar ao carrinho) | RF-C02.3 | ✅ Mantido |
| RQ1.3 (Usuário - Carrinho mesmo restaurante) | RF-C02.4 | ✅ Mantido |
| RQ1.4 (Usuário - Observações por item) | RF-C02.5 | ✅ Mantido |
| RQ2 (Pagamento via cartão) | RF-C03.1 | ✅ Mantido |
| RQ3 (Cancelar antes de confirmar) | RF-C04.2 | ✅ Mantido |
| RQ3.1 (Não modificar após captura) | RF-C04.3 | ✅ Mantido |
| RQ4 (Recusar recebimento) | RF-C05.1, RF-C05.2 | ✅ Mantido |
| RQ1 (Restaurante - Múltiplos cardápios) | RF-R01.1 | ✅ Mantido |
| RQ2 (Restaurante - Editar cardápio) | RF-R01.2, RF-R01.3 | ✅ Mantido |
| RQ3 (Restaurante - Listar pedidos) | RF-R02.1 | ✅ Mantido |
| RQ4 (Restaurante - Confirmar pedido) | RF-R02.2 | ✅ Mantido |
| RQ5 (Restaurante - Recusar pedido) | RF-R02.3 | ✅ Mantido |
| RQ6 (Restaurante - Pedido pronto) | RF-R02.4 | ✅ Mantido |
| RQ1 (Sistema - Enviar após reserva) | RF-S01.1, RF-S01.3 | ✅ Mantido |
| RQ1.1 (Sistema - Erro na reserva) | RF-S01.2 | ✅ Mantido |
| RQ2 (Sistema - Efetivar pagamento) | RF-S02.1 | ✅ Mantido |
| RQ2.1 (Sistema - Recusa restaurante) | RF-S02.2 | ✅ Mantido |
| — | RF-C01.* (Cadastro/Auth) | ➕ Adicionado |
| — | RF-C02.6, RF-C02.7 (Carrinho) | ➕ Adicionado |
| — | RF-C03.2, RF-C03.3 (Gestão cartão) | ➕ Adicionado |
| — | RF-S03.* (Fluxo entrega) | ➕ Adicionado |
| — | RF-S04.* (Compensação) | ➕ Adicionado |
| — | RNF-01..18 (Não-funcionais) | ➕ Adicionado |
