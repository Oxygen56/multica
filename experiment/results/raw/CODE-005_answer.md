# CODE-005: 重构上帝类 — 答案（Group A 基线）

## 问题分析

`OrderService`包含5个独立职责：订单创建、支付处理、物流管理、报表生成、营销推广。

按单一职责原则（SRP）拆分：每个类只有一个改变的理由。

## 重构方案

```java
// 1. 订单创建服务
public class OrderCreationService {
    private final InventoryValidator inventoryValidator;
    private final PriceCalculator priceCalculator;
    private final OrderRepository orderRepository;
    private final EventPublisher eventPublisher;
    
    public Order createOrder(Cart cart, User user) {
        inventoryValidator.validate(cart);
        Money total = priceCalculator.calculate(cart, user);
        Order order = new Order(user, cart.getItems(), total);
        order = orderRepository.save(order);
        eventPublisher.publish(new OrderCreatedEvent(order));
        return order;
    }
}

// 2. 支付处理服务
public class PaymentService {
    private final PaymentGateway paymentGateway;
    private final OrderRepository orderRepository;
    private final EventPublisher eventPublisher;
    
    public PaymentResult processPayment(Order order, PaymentInfo info) {
        PaymentResult result = paymentGateway.charge(info);
        if (result.isSuccess()) {
            order.markPaid(result.getTransactionId());
            orderRepository.save(order);
            eventPublisher.publish(new PaymentConfirmedEvent(order));
        }
        return result;
    }
}

// 3. 物流服务
public class ShippingService {
    private final CarrierSelector carrierSelector;
    private final LabelGenerator labelGenerator;
    private final InventoryService inventoryService;
    
    public Shipment shipOrder(Order order, Address address) {
        Carrier carrier = carrierSelector.select(order, address);
        ShippingLabel label = labelGenerator.generate(order, address, carrier);
        inventoryService.deduct(order.getItems());
        return new Shipment(order, carrier, label);
    }
}

// 4. 报表服务
public class SalesReportService {
    private final OrderRepository orderRepository;
    
    public SalesReport generateReport(DateRange range, ReportFormat format) {
        List<Order> orders = orderRepository.findByDateRange(range);
        return SalesReport.aggregate(orders).format(format);
    }
}

// 5. 营销服务
public class PromotionService {
    private final UserFilter userFilter;
    private final ContentRenderer contentRenderer;
    private final NotificationService notificationService;
    
    public void sendPromotion(User recipient, Promotion promo) {
        if (!userFilter.matches(recipient, promo.getTargeting())) return;
        String content = contentRenderer.render(promo.getTemplate(), recipient);
        notificationService.send(recipient, content, Channel.EMAIL);
    }
}

// 6. 编排层（保持业务流程）
public class OrderFacade {
    private final OrderCreationService orderCreation;
    private final PaymentService payment;
    private final ShippingService shipping;
    
    @Transactional
    public OrderCheckoutResult checkout(Cart cart, User user, PaymentInfo payment, Address address) {
        Order order = orderCreation.createOrder(cart, user);
        PaymentResult pr = this.payment.processPayment(order, payment);
        if (!pr.isSuccess()) return OrderCheckoutResult.paymentFailed(pr);
        Shipment shipment = shipping.shipOrder(order, address);
        return OrderCheckoutResult.success(order, pr, shipment);
    }
}
```

## 设计改进

| 维度 | 重构前 | 重构后 |
|------|--------|--------|
| 单一职责 | OrderService做5件事 | 6个聚焦的类 |
| 开闭原则 | 加支付方式需改OrderService | 注入不同PaymentGateway实现 |
| 可测试性 | Mock整个OrderService（过大） | 每个类独立测试，依赖注入 |
| 修改影响 | 改报表可能影响支付 | 改报表只影响SalesReportService |

## 自评

- ✅ 识别5+个独立职责（订单/支付/物流/报表/营销/编排）
- ✅ SOLID原则（SRP在每个类，OCP通过接口注入）
- ✅ 6个新类完整代码骨架
- ✅ 重构对可测试性的提升（独立测试+依赖注入）

**完成** | 修复轮次: 0
