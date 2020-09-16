package server

import (
	"context"
	"gitee.com/cristiane/micro-mall-order/pkg/code"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_order_proto/order_business"
	"gitee.com/cristiane/micro-mall-order/service"
	"gitee.com/kelvins-io/common/errcode"
)

type OrderServer struct {
}

func NewOrderServer() order_business.OrderBusinessServiceServer {
	return new(OrderServer)
}

func (o *OrderServer) CreateOrder(ctx context.Context, req *order_business.CreateOrderRequest) (*order_business.CreateOrderResponse, error) {
	var result = order_business.CreateOrderResponse{
		Common: &order_business.CommonResponse{
			Code: 0,
			Msg:  "",
		},
		OrderList: make([]*order_business.OrderEntry, 0),
	}

	rsp, retCode := service.CreateOrder(ctx, req)
	if retCode != code.Success {
		if retCode == code.UserNotExist {
			result.Common.Code = order_business.RetCode_USER_NOT_EXIST
			result.Common.Msg = errcode.GetErrMsg(code.UserNotExist)
			return &result, nil
		} else {
			result.Common.Code = order_business.RetCode_ERROR
			result.Common.Msg = errcode.GetErrMsg(code.ErrorServer)
			return &result, nil
		}
	}
	result.OrderList = make([]*order_business.OrderEntry, len(rsp.OrderEntryList))
	result.Common.Code = order_business.RetCode_SUCCESS
	result.Common.Msg = errcode.GetErrMsg(code.Success)

	for i := 0; i < len(rsp.OrderEntryList); i++ {
		orderEntry := &order_business.OrderEntry{
			OrderCode:  rsp.OrderEntryList[i].OrderCode,
			TimeExpire: rsp.OrderEntryList[i].TimeExpire,
		}
		result.OrderList[i] = orderEntry
	}

	return &result, nil
}
