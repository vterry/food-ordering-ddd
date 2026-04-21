## 1. Objetivo

Criar o backend de um aplicativo de delivery de comida, de forma que seja possivel explorar conceitos avançados de arquitetura distribuida.

Literatura base:

- _Software Architecture: The Hard Parts_ (Neal Ford) - Trade-offs e Granularidade.
- _Building Microservices_ (Sam Newman) - Acoplamento e Coesão.
- _Building Event-Driven Microservices_ (Adam Bellemare) - Topologias e Schemas.

Ferramentas que devem ser utilizadas:
- Go como linguagem principal para escrita do código
- MySQL como banco de dados
- SQLC como ferramenta para geração de código GO a parti de Queries SQL
- GRPC para interface sincronas quanto necessário
- Para interfaces HTTP deve ser utilizado a biblioteca padrão do GO
- RabbitMQ como ferramenta para troca de mensagens/eventos


## Descrição do Aplicativo de Comida
Queremos construir um sistema de delivery de comida, de forma que seja possivel.

REQUISITOS USUÁRIO:

RQ1. Interação com Restaurantes
RQ1.1. O usuário deverá ser capaz de listar restaurantes disponveis
RQ1.1. Para cada restaurante disponivel, o usuário deverá ser capaz de visualizar os itens disponiveis no cardápio do restaurante
RQ1.2. O usuário deverá ser capaz de adicionar um do restaurante no carrinho de compras.
RQ1.3. O carrinho de compras só deve permitir a inclusão de produtos do mesmo restaurante. Se o usuário tentar adicionar itens de outro restaurante o carrinho deverá informar erro.
RQ1.4. O usuário deverá ser capaz de incluir observações nos produtos adicionados ao carrinho. A observação fica associado ao item adicionado e não ao carrinho todo.

Interação com Paragamento
RQ2. O usuário deverá ter a opção de pagar somente via cartão de crédito

Interação com Pedido
RQ3. O usuário pode ser capaz de cancelar um pedido durante a fase de confirmação, depois que o pedido foi confirmado ele não pode ser cancelado.
RQ3.1 Depois que o pedido foi capturado, ele não pode ser modificado.

Interação com o Delivery
RQ4. O usuário pode recusar a receber o pedido, se isso ocorrer o sistema deverá fazer o cancelamento do pedido e solicitar o reembolso a operadora do cartão.


REQUISITOS DO RESTAURANTE

Gestão de Cardápio
RQ1. O restaurante poderá ter mais um cardápio, mas apenas um poderá ser o principal ativo na plataforma
RQ2. O restaruante deverá ser capaz de editar seu cardápio ou ajustar o cardápio principal.

Gestão de Pedidos
RQ3. O restaurante deverá ser capaz de visualizar a lista de pedidos que chegou para seu estabelecimento
RQ4. O restaurante deverá ser capaz de confirmar um pedido caso seja possivel de ser atentido.
RQ5. O restaraunte deverá ser capaz de recusar um pedido caso não seja possivel atende-lo.
RQ6. O restaurante deverá ser capaz de sinalizar a plataforma que o pedido está PRONTO


REQUISITOS GERAIS DO SISTEMA
RQ1. O sistema deverá enviar um pedido para o restaurante após a solicitação de reserva do pagamento
RQ1.1 Se houver erro na reserva do pagamento o pedido deve ser rejeitado na plataforma.
RQ2. O sistema deverá efetivar o pagamento assim que houver a confirmação do restaurante.
RQ2.1. Se houver recusa do restaurante o valor reservado do limite deve ser liberado e o pedido cancelado.