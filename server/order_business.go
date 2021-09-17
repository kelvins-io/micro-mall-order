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
			Code: order_business.RetCode_SUCCESS,
		},
	}
	rsp, retCode := service.CreateOrder(ctx, req)
	result.TxCode = rsp.TxCode
	if retCode != code.Success {
		if retCode == code.SkuPriceVersionNotExist {
			result.Common.Code = order_business.RetCode_SKU_PRICE_VERSION_NOT_EXIST
		} else if retCode == code.OrderDeliveryNotExist {
			result.Common.Code = order_business.RetCode_ORDER_DELIVERY_NOT_EXIST
		} else if retCode == code.OrderTxCodeEmpty {
			result.Common.Code = order_business.RetCode_ORDER_TX_CODE_EMPTY
		} else if retCode == code.OrderExist {
			result.Common.Code = order_business.RetCode_ORDER_EXIST
		} else if retCode == code.UserNotExist {
			result.Common.Code = order_business.RetCode_USER_NOT_EXIST
		} else if retCode == code.SkuAmountNotEnough {
			result.Common.Code = order_business.RetCode_SKU_AMOUNT_NOT_ENOUGH
		} else if retCode == code.TransactionFailed {
			result.Common.Code = order_business.RetCode_TRANSACTION_FAILED
		} else if retCode == code.UserStateNotVerify {
			result.Common.Code = order_business.RetCode_USER_STATE_NOT_VERIFY
		} else {
			result.Common.Code = order_business.RetCode_ERROR
		}
		result.Common.Msg = errcode.GetErrMsg(retCode)
		return &result, nil
	}

	return &result, nil
}

func (o *OrderServer) GetOrderDetail(ctx context.Context, req *order_business.GetOrderDetailRequest) (*order_business.GetOrderDetailResponse, error) {
	var result order_business.GetOrderDetailResponse
	result.Common = &order_business.CommonResponse{
		Code: order_business.RetCode_SUCCESS,
	}
	rsp, retCode := service.GetOrderDetail(ctx, req)
	if retCode != code.Success {
		switch retCode {
		case code.OrderTxCodeNotExist:
			result.Common.Code = order_business.RetCode_ORDER_TX_CODE_NOT_EXIST
		case code.OrderExpire:
			result.Common.Code = order_business.RetCode_ORDER_EXPIRE
		case code.OrderPayIng:
			result.Common.Code = order_business.RetCode_ORDER_PAY_ING
		case code.OrderPayCompleted:
			result.Common.Code = order_business.RetCode_ORDER_PAY_COMPLETED
		case code.OrderStateInvalid:
			result.Common.Code = order_business.RetCode_ORDER_STATE_INVALID
		case code.OrderStateLocked:
			result.Common.Code = order_business.RetCode_ORDER_STATE_LOCKED
		default:
			result.Common.Code = order_business.RetCode_ERROR
		}
		result.Common.Msg = errcode.GetErrMsg(retCode)
		return &result, nil
	}
	result.CoinType = order_business.CoinType(rsp.CoinType)
	result.List = make([]*order_business.ShopOrderDetail, len(rsp.List))
	for i := 0; i < len(rsp.List); i++ {
		shopOrderDe := &order_business.ShopOrderDetail{
			ShopId:      rsp.List[i].ShopId,
			OrderCode:   rsp.List[i].OrderCode,
			Description: rsp.List[i].Description,
			Money:       rsp.List[i].Amount,
		}
		result.List[i] = shopOrderDe
	}

	return &result, nil
}

func (o *OrderServer) GetOrderSku(ctx context.Context, req *order_business.GetOrderSkuRequest) (*order_business.GetOrderSkuResponse, error) {
	result := &order_business.GetOrderSkuResponse{
		Common: &order_business.CommonResponse{
			Code: order_business.RetCode_SUCCESS,
		},
	}
	orderSku, retCode := service.GetOrderSku(ctx, req)
	if retCode != code.Success {
		result.Common.Code = order_business.RetCode_ERROR
		result.Common.Msg = errcode.GetErrMsg(retCode)
		return result, nil
	}
	result.OrderList = make([]*order_business.OrderSku, len(orderSku.SkuList))
	for i := 0; i < len(orderSku.SkuList); i++ {
		row := orderSku.SkuList[i]
		orderSku := &order_business.OrderSku{
			OrderCode: row.OrderCode,
			Goods:     nil,
		}
		goods := make([]*order_business.OrderGoods, len(row.SkuList))
		for j := 0; j < len(row.SkuList); j++ {
			orderGoods := &order_business.OrderGoods{
				SkuCode: row.SkuList[j].SkuCode,
				Price:   row.SkuList[j].Price,
				Amount:  int64(row.SkuList[j].Amount),
				Name:    row.SkuList[j].Name,
			}
			goods[j] = orderGoods
		}
		orderSku.Goods = goods
		result.OrderList[i] = orderSku
	}
	return result, nil
}

func (o *OrderServer) UpdateOrderState(ctx context.Context, req *order_business.UpdateOrderStateRequest) (*order_business.UpdateOrderStateResponse, error) {
	var result = order_business.UpdateOrderStateResponse{
		Common: &order_business.CommonResponse{
			Code: order_business.RetCode_SUCCESS,
		},
	}
	retCode := service.UpdateOrderState(ctx, req)
	if retCode != code.Success {
		if retCode == code.UserNotExist {
			result.Common.Code = order_business.RetCode_USER_NOT_EXIST
		} else if retCode == code.OperationNotEffect {
			result.Common.Code = order_business.RetCode_OPERATION_NOT_EFFECT
		} else if retCode == code.OrderNotExist {
			result.Common.Code = order_business.RetCode_ORDER_NOT_EXIST
		} else if retCode == code.OrderStateLocked {
			result.Common.Code = order_business.RetCode_ORDER_STATE_LOCKED
		} else if retCode == code.OrderStateProhibit {
			result.Common.Code = order_business.RetCode_ORDER_STATE_PROHIBIT
		} else {
			result.Common.Code = order_business.RetCode_ERROR
		}
		result.Common.Msg = errcode.GetErrMsg(retCode)
		return &result, nil
	}
	return &result, nil
}

func (o *OrderServer) GenOrderTxCode(ctx context.Context, req *order_business.GenOrderTxCodeRequest) (*order_business.GenOrderTxCodeResponse, error) {
	result := &order_business.GenOrderTxCodeResponse{
		Common: &order_business.CommonResponse{
			Code: order_business.RetCode_SUCCESS,
		},
	}
	txCode, retCode := service.GenOrderTxCode(ctx, req)
	if retCode != code.Success {
		result.Common.Code = order_business.RetCode_ERROR
		return result, nil
	}
	result.OrderTxCode = txCode
	return result, nil
}

func (o *OrderServer) CheckOrderExist(ctx context.Context, req *order_business.CheckOrderExistRequest) (*order_business.CheckOrderExistResponse, error) {
	result := &order_business.CheckOrderExistResponse{
		Common: &order_business.CommonResponse{
			Code: order_business.RetCode_SUCCESS,
		},
	}
	isExist, retCode := service.CheckOrderExist(ctx, req)
	switch retCode {
	case code.Success:
		result.Common.Code = order_business.RetCode_SUCCESS
	case code.OrderExist:
		result.Common.Code = order_business.RetCode_ORDER_EXIST
	case code.OrderNotExist:
		result.Common.Code = order_business.RetCode_ORDER_NOT_EXIST
	case code.ErrorServer:
		result.Common.Code = order_business.RetCode_ERROR
	}
	result.IsExist = isExist
	return result, nil
}

func (o *OrderServer) OrderTradeNotice(ctx context.Context, req *order_business.OrderTradeNoticeRequest) (*order_business.OrderTradeNoticeResponse, error) {
	result := &order_business.OrderTradeNoticeResponse{Common: &order_business.CommonResponse{
		Code: order_business.RetCode_SUCCESS,
	}}
	retCode := service.OrderTradeNotice(ctx, req)
	switch retCode {
	case code.Success:
		result.Common.Code = order_business.RetCode_SUCCESS
	case code.UserNotExist:
		result.Common.Code = order_business.RetCode_USER_NOT_EXIST
	case code.OrderTxCodeNotExist:
		result.Common.Code = order_business.RetCode_ORDER_TX_CODE_NOT_EXIST
	case code.ErrorServer:
		result.Common.Code = order_business.RetCode_ERROR
	}
	return result, nil
}

func (o *OrderServer) CheckOrderState(ctx context.Context, req *order_business.CheckOrderStateRequest) (*order_business.CheckOrderStateResponse, error) {
	result := &order_business.CheckOrderStateResponse{
		Common: &order_business.CommonResponse{
			Code: order_business.RetCode_SUCCESS,
		},
		List: nil,
	}
	if len(req.GetOrderCodes()) > 500 {
		result.Common.Code = order_business.RetCode_REQUEST_DATA_TOO_MUCH
		return result, nil
	}
	stateList, retCode := service.CheckOrderState(ctx, req)
	if retCode != code.Success {
		result.Common.Code = order_business.RetCode_ERROR
		return result, nil
	}
	result.List = stateList
	return result, nil
}

func (o *OrderServer) FindOrderList(ctx context.Context, req *order_business.FindOrderListRequest) (*order_business.FindOrderListResponse, error) {
	result := &order_business.FindOrderListResponse{
		Common: &order_business.CommonResponse{
			Code: order_business.RetCode_SUCCESS,
		},
		List:  nil,
		Total: 0,
	}
	list, total, retCode := service.FindOrderList(ctx, req)
	if retCode != code.Success {
		switch retCode {
		case code.ErrRequestDataFormat:
			result.Common.Code = order_business.RetCode_ERR_REQUEST_DATA_FORMAT
		default:
			result.Common.Code = order_business.RetCode_ERROR
		}
		return result, nil
	}
	result.Total = total
	result.List = list
	return result, nil
}

func (o *OrderServer) InspectShopOrder(ctx context.Context, req *order_business.InspectShopOrderRequest) (*order_business.InspectShopOrderResponse, error) {
	result := &order_business.InspectShopOrderResponse{Common: &order_business.CommonResponse{
		Code: order_business.RetCode_SUCCESS,
		Msg:  "",
	}}
	retCode := service.InspectShopOrder(ctx, req)
	if retCode != code.Success {
		switch retCode {
		case code.OrderNotExist:
			result.Common.Code = order_business.RetCode_ORDER_NOT_EXIST
		case code.OrderStateInvalid:
			result.Common.Code = order_business.RetCode_ORDER_STATE_INVALID
		default:
			result.Common.Code = order_business.RetCode_ERROR
		}
		result.Common.Msg = errcode.GetErrMsg(retCode)
		return result, nil
	}
	return result, nil
}
