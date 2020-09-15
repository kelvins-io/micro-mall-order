package server

import (
	"context"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_order_proto/order_business"
)

type OrderServer struct {
}

func NewOrderServer() order_business.OrderBusinessServiceServer {
	return new(OrderServer)
}

func (o *OrderServer) CreateOrder(ctx context.Context, req *order_business.CreateOrderRequest) (*order_business.CreateOrderResponse, error) {
	return &order_business.CreateOrderResponse{}, nil
}
