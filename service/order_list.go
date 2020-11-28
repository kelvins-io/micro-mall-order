package service

import (
	"context"
	"gitee.com/cristiane/micro-mall-order/pkg/code"
	"gitee.com/cristiane/micro-mall-order/pkg/util"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_order_proto/order_business"
	"gitee.com/cristiane/micro-mall-order/repository"
	"gitee.com/kelvins-io/kelvins"
)

const (
	sqlSelectFindOrderList = "order_code,uid,shop_id,description,client_ip,device_code,state,pay_state,money,create_time"
)

func FindOrderList(ctx context.Context, req *order_business.FindOrderListRequest) (result []*order_business.OrderListEntry, total int64, retCode int) {
	result = make([]*order_business.OrderListEntry, 0)
	total = int64(0)
	retCode = code.Success
	startTime, err := util.GenTimeOfStr(req.GetTimeMeta().GetStartTime())
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "FindOrderList startTime parse err: %v, req: %v", err, req.GetTimeMeta().GetStartTime())
		retCode = code.ErrRequestDataFormat
		return
	}
	startTime = startTime.UTC()
	endTime, err := util.GenTimeOfStr(req.GetTimeMeta().GetEndTime())
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "FindOrderList endTime parse err: %v, req: %v", err, req.GetTimeMeta().GetEndTime())
		retCode = code.ErrRequestDataFormat
		return
	}
	endTime = endTime.UTC()
	where := map[string]interface{}{
		"shop_id": req.GetShopIdList(),
		"uid":     req.GetUidList(),
	}
	list, total, err := repository.FindOrderListByTime(sqlSelectFindOrderList, where, startTime, endTime, req.GetPageMeta().GetPageSize(), req.GetPageMeta().GetPageNum())
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "FindOrderListByShopId err: %v, shopId: %v", err, req.GetShopIdList())
		retCode = code.ErrorServer
		return
	}
	result = make([]*order_business.OrderListEntry, len(list))
	for i := 0; i < len(list); i++ {
		entry := &order_business.OrderListEntry{
			OrderCode:   list[i].OrderCode,
			Uid:         list[i].Uid,
			ShopId:      list[i].ShopId,
			Description: list[i].Description,
			ClientIp:    list[i].ClientIp,
			DeviceCode:  list[i].DeviceCode,
			State:       order_business.OrderStateType(list[i].State),
			PayState:    order_business.OrderPayStateType(list[i].PayState),
			Money:       list[i].Money,
			CreateTime:  util.ParseTimeOfStr(list[i].CreateTime.Unix()),
		}
		result[i] = entry
	}

	return
}
