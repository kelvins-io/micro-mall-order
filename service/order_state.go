package service

import (
	"context"
	"gitee.com/cristiane/micro-mall-order/model/mysql"
	"gitee.com/cristiane/micro-mall-order/pkg/code"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_order_proto/order_business"
	"gitee.com/cristiane/micro-mall-order/repository"
	"gitee.com/kelvins-io/kelvins"
)

const (
	sqlSelectCheckOrderState = "order_code,pay_state,state"
)

func CheckOrderState(ctx context.Context, req *order_business.CheckOrderStateRequest) ([]*order_business.OrderState, int) {
	result := make([]*order_business.OrderState, 0)
	orderList, err := repository.FindOrderList(sqlSelectCheckOrderState, req.GetOrderCodes())
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "FindOrderList err: %v, orderCode: %v", err, req.GetOrderCodes())
		return result, code.ErrorServer
	}
	orderCodeToOrder := map[string]mysql.Order{}
	for i := 0; i < len(orderList); i++ {
		orderCodeToOrder[orderList[i].OrderCode] = orderList[i]
	}
	for i := 0; i < len(req.OrderCodes); i++ {
		orderState := &order_business.OrderState{
			OrderCode: req.OrderCodes[i],
			PayState:  0,
			State:     0,
			IsExist:   false,
		}
		if _, ok := orderCodeToOrder[req.OrderCodes[i]]; !ok {
			orderState.IsExist = false
			result = append(result, orderState)
			continue
		}
		orderState.PayState = order_business.OrderPayStateType(orderCodeToOrder[req.OrderCodes[i]].PayState)
		orderState.State = order_business.OrderStateType(orderCodeToOrder[req.OrderCodes[i]].State)
		result = append(result, orderState)
	}
	return result, code.Success
}
