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
		orderState.IsExist = true
		payState := order_business.OrderPayStateType_PAY_READY
		switch orderCodeToOrder[req.OrderCodes[i]].PayState {
		case 0:
			payState = order_business.OrderPayStateType_PAY_READY
		case 1:
			payState = order_business.OrderPayStateType_PAY_RUN
		case 2:
			payState = order_business.OrderPayStateType_PAY_FAILED
		case 3:
			payState = order_business.OrderPayStateType_PAY_SUCCESS
		case 4:
			payState = order_business.OrderPayStateType_PAY_CANCEL
		}
		state := order_business.OrderStateType_ORDER_EFFECTIVE
		switch orderCodeToOrder[req.OrderCodes[i]].State {
		case 0:
			state = order_business.OrderStateType_ORDER_EFFECTIVE
		case 1:
			state = order_business.OrderStateType_ORDER_LOCKED
		case 2:
			state = order_business.OrderStateType_ORDER_INVALID
		}
		orderState.PayState = payState
		orderState.State = state
		result = append(result, orderState)
	}
	return result, code.Success
}
