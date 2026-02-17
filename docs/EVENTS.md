# 📋 ESPECIFICAÇÃO COMPLETA DOS EVENTOS DE DOMÍNIO

## Sistema de Delivery de Comida - DDD + Saga + Outbox + EDA

**Versão**: 1.0  
**Data**: 05/02/2026  
**Total de Eventos**: 25

---

## 🎯 CONVENÇÕES GERAIS (OBRIGATÓRIOS EM TODOS OS EVENTOS)

```json
{
  "eventId": "uuid",
  "eventType": "string",
  "aggregateType": "Order|Payment|RestaurantOrder|Delivery|Restaurant|Menu",
  "aggregateId": "uuid",
  "version": "int",
  "timestamp": "ISO8601",
  "correlationId": "uuid"
}
```

## 🛒 CUSTOMER ORDERING CONTEXT (8 eventos)

**OrderPlaced**

```json
{
  "eventId": "uuid",
  "eventType": "OrderPlaced",
  "aggregateType": "Order",
  "aggregateId": "orderId",
  "version": 1,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "customerId": "uuid",
  "restaurantId": "uuid",
  "totalAmount": { "amount": "decimal", "currency": "string" },
  "deliveryAddress": {
    "street": "string",
    "number": "string",
    "neighborhood": "string",
    "city": "string",
    "state": "string",
    "zipCode": "string"
  },
  "itemsSummary": [
    {
      "menuItemId": "uuid",
      "name": "string",
      "quantity": "int",
      "unitPrice": { "amount": "decimal", "currency": "string" }
    }
  ]
}
```

**OrderPaid**

```json
{
  "eventId": "uuid",
  "eventType": "OrderPaid",
  "aggregateType": "Order",
  "aggregateId": "orderId",
  "version": 2,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "paymentId": "uuid",
  "totalAmount": { "amount": "decimal", "currency": "string" }
}
```

**OrderCancelled**

```json
{
  "eventId": "uuid",
  "eventType": "OrderConfirmed",
  "aggregateType": "Order",
  "aggregateId": "orderId",
  "version": 4,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "restaurantOrderId": "uuid",
  "estimatedPreparationTime": "ISO8601|null"
}
```

**OrderConfirmed**

```json
{
  "eventId": "uuid",
  "eventType": "OrderPaid",
  "aggregateType": "Order",
  "aggregateId": "orderId",
  "version": 2,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "paymentId": "uuid",
  "totalAmount": { "amount": "decimal", "currency": "string" }
}
```

**OrderInDelivery**

```json
{
  "eventId": "uuid",
  "eventType": "OrderInDelivery",
  "aggregateType": "Order",
  "aggregateId": "orderId",
  "version": 5,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "deliveryId": "uuid",
  "courierId": "uuid",
  "estimatedDeliveryTime": "ISO8601|null"
}
```

**OrderDelivered**

```json
{
  "eventId": "uuid",
  "eventType": "OrderDelivered",
  "aggregateType": "Order",
  "aggregateId": "orderId",
  "version": 6,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "deliveryId": "uuid",
  "deliveredAt": "ISO8601",
  "courierId": "uuid"
}
```

**OrderFailed**

```json
{
  "eventId": "uuid",
  "eventType": "OrderFailed",
  "aggregateType": "Order",
  "aggregateId": "orderId",
  "version": 7,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "reason": "DELIVERY_FAILED|TECHNICAL_ERROR|SUPPORT_NEEDED",
  "refundedAmount": { "amount": "decimal|null", "currency": "string|null" }
}
```

## 💳 PAYMENT CONTEXT (5 eventos)

**PaymentInitiated**

```json
{
  "eventId": "uuid",
  "eventType": "PaymentInitiated",
  "aggregateType": "Payment",
  "aggregateId": "paymentId",
  "version": 1,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "orderId": "uuid",
  "customerId": "uuid",
  "amount": { "amount": "decimal", "currency": "string" },
  "method": "CARD|PIX|WALLET"
}
```

**PaymentAuthorized**

```json
{
  "eventId": "uuid",
  "eventType": "PaymentAuthorized",
  "aggregateType": "Payment",
  "aggregateId": "paymentId",
  "version": 2,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "orderId": "uuid",
  "customerId": "uuid",
  "amount": { "amount": "decimal", "currency": "string" },
  "externalTransactionId": "string"
}
```

**PaymentCaptured**

```json
{
  "eventId": "uuid",
  "eventType": "PaymentCaptured",
  "aggregateType": "Payment",
  "aggregateId": "paymentId",
  "version": 3,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "orderId": "uuid",
  "amount": { "amount": "decimal", "currency": "string" },
  "externalCaptureId": "string"
}
```

**PaymentFailed**

```json
{
  "eventId": "uuid",
  "eventType": "PaymentFailed",
  "aggregateType": "Payment",
  "aggregateId": "paymentId",
  "version": 2,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "orderId": "uuid",
  "reason": "INSUFFICIENT_FUNDS|INVALID_CARD|GATEWAY_ERROR|DECLINED"
}
```

**PaymentRefunded**

```json
{
  "eventId": "uuid",
  "eventType": "PaymentRefunded",
  "aggregateType": "Payment",
  "aggregateId": "paymentId",
  "version": 4,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "orderId": "uuid",
  "refundedAmount": { "amount": "decimal", "currency": "string" },
  "externalRefundId": "string",
  "partial": "boolean"
}
```

## 🍽️ RESTAURANT MANAGEMENT CONTEXT (4 eventos)

**RestaurantOrderCreated**

```json
{
  "eventId": "uuid",
  "eventType": "RestaurantOrderCreated",
  "aggregateType": "RestaurantOrder",
  "aggregateId": "restaurantOrderId",
  "version": 1,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "restaurantId": "uuid",
  "customerOrderId": "uuid",
  "totalAmount": { "amount": "decimal", "currency": "string" },
  "itemsCount": "int"
}
```

**RestaurantOrderAccepted**

```json
{
  "eventId": "uuid",
  "eventType": "RestaurantOrderAccepted",
  "aggregateType": "RestaurantOrder",
  "aggregateId": "restaurantOrderId",
  "version": 2,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "customerOrderId": "uuid",
  "restaurantId": "uuid",
  "estimatedPreparationTime": "ISO8601"
}
```

**RestaurantOrderRejected**

```json
{
  "eventId": "uuid",
  "eventType": "RestaurantOrderRejected",
  "aggregateType": "RestaurantOrder",
  "aggregateId": "restaurantOrderId",
  "version": 2,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "customerOrderId": "uuid",
  "reason": "OUT_OF_STOCK|CLOSED|BUSY|MAX_ORDERS"
}
```

**RestaurantOrderStatusChanged**

```json
{
  "eventId": "uuid",
  "eventType": "RestaurantOrderStatusChanged",
  "aggregateType": "RestaurantOrder",
  "aggregateId": "restaurantOrderId",
  "version": 3,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "customerOrderId": "uuid",
  "newStatus": "PREPARING|READY_FOR_PICKUP|HANDED_TO_COURIER",
  "previousStatus": "string"
}
```

## 🚚 DELIVERY LOGISTICS CONTEXT (3 eventos)

**DeliveryCreated**

```json
{
  "eventId": "uuid",
  "eventType": "DeliveryCreated",
  "aggregateType": "Delivery",
  "aggregateId": "deliveryId",
  "version": 1,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "orderId": "uuid",
  "restaurantOrderId": "uuid",
  "pickupAddress": {
    "street": "string",
    "number": "string",
    "neighborhood": "string",
    "city": "string",
    "state": "string",
    "zipCode": "string"
  },
  "dropoffAddress": {
    "street": "string",
    "number": "string",
    "neighborhood": "string",
    "city": "string",
    "state": "string",
    "zipCode": "string"
  },
  "status": "PENDING_ASSIGNMENT"
}
```

**DeliveryAssigned**

```json
{
  "eventId": "uuid",
  "eventType": "DeliveryAssigned",
  "aggregateType": "Delivery",
  "aggregateId": "deliveryId",
  "version": 2,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "orderId": "uuid",
  "courierId": "uuid",
  "estimatedDeliveryTime": "ISO8601|null"
}
```

**DeliveryStatusChanged**

```json
{
  "eventId": "uuid",
  "eventType": "DeliveryStatusChanged",
  "aggregateType": "Delivery",
  "aggregateId": "deliveryId",
  "version": 3,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "orderId": "uuid",
  "newStatus": "PICKED_UP|ON_ROUTE|DELIVERED|FAILED",
  "previousStatus": "string",
  "details": {
    "pickedUpAt": "ISO8601|null",
    "deliveredAt": "ISO8601|null",
    "failureReason": "string|null"
  }
}
```

## 🏪 CATALOG CONTEXT (5 eventos)

**RestaurantCreated**

```json
{
  "eventId": "uuid",
  "eventType": "RestaurantCreated",
  "aggregateType": "Restaurant",
  "aggregateId": "restaurantId",
  "version": 1,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "name": "string",
  "legalName": "string",
  "document": "string",
  "status": "INACTIVE"
}
```

**RestaurantActivated**

```json
{
  "eventId": "uuid",
  "eventType": "RestaurantActivated",
  "aggregateType": "Restaurant",
  "aggregateId": "restaurantId",
  "version": 2,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "status": "ACTIVE"
}
```

**MenuCreated**

```json
{
  "eventId": "uuid",
  "eventType": "MenuCreated",
  "aggregateType": "Menu",
  "aggregateId": "menuId",
  "version": 1,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "restaurantId": "uuid",
  "name": "string",
  "status": "DRAFT"
}
```

**MenuActivated**

```json
{
  "eventId": "uuid",
  "eventType": "MenuActivated",
  "aggregateType": "Menu",
  "aggregateId": "menuId",
  "version": 2,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "restaurantId": "uuid",
  "name": "string",
  "previousMenuId": "uuid|null"
}
```

**MenuItemAvailabilityChanged**

```json
{
  "eventId": "uuid",
  "eventType": "MenuItemAvailabilityChanged",
  "aggregateType": "MenuItem",
  "aggregateId": "menuItemId",
  "version": 2,
  "timestamp": "ISO8601",
  "correlationId": "uuid",
  "menuId": "uuid",
  "restaurantId": "uuid",
  "newStatus": "AVAILABLE|TEMP_UNAVAILABLE|UNAVAILABLE",
  "previousStatus": "string"
}
```
