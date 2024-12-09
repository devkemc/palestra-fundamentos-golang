package orders

import (
	"context"
	"github.com/devkemc/fundamentos-golang/emails"
	"github.com/devkemc/fundamentos-golang/payments"
)

type orderServiceV1 struct {
	orderRepository OrderRepository
	emailService    emails.EmailService
	paymentService  payments.PaymentService
}

func (o orderServiceV1) GetOrderDetails(ctx context.Context, orderId int64) (*Order, error) {
	return o.orderRepository.FindOrderById(ctx, orderId)
}

func (o orderServiceV1) GetOrders(ctx context.Context) ([]Order, error) {
	return o.orderRepository.GetAllOrders(ctx)
}

func (o orderServiceV1) Sell(ctx context.Context, order *Order) error {
	if err := order.ValidateToSell(ctx); err != nil {
		return err
	}

	o.orderRepository.InitTransaction(ctx)

	order.Status = orderPending
	orderId, err := o.orderRepository.SaveOrder(ctx, *order)
	if err != nil {
		o.orderRepository.Rollback()
		return err
	}

	err = o.orderRepository.Commit()
	if err != nil {
		return err
	}

	order.Id = orderId
	if err := o.paymentService.ProcessPayments(ctx, order.Payments, orderId); err != nil {
		return err
	}

	if err := o.orderRepository.ConfirmOrder(ctx, order.Id); err != nil {
		return err
	}

	if err := o.emailService.SendEmail(ctx, emails.Email{}); err != nil {
		return err
	}

	return nil
}

func NewOrderServiceV1(orderRepository OrderRepository, emailService emails.EmailService, paymentService payments.PaymentService) OrderService {
	return &orderServiceV1{orderRepository, emailService, paymentService}
}
