package service

import (
	"context"
	"fmt"
	"gitee.com/cristiane/micro-mall-order/model/args"
	"gitee.com/cristiane/micro-mall-order/model/mysql"
	"gitee.com/cristiane/micro-mall-order/pkg/code"
	"gitee.com/cristiane/micro-mall-order/pkg/util"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_order_proto/order_business"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_search_proto/search_business"
	"gitee.com/cristiane/micro-mall-order/repository"
	"gitee.com/kelvins-io/common/json"
	"gitee.com/kelvins-io/kelvins"
	"time"
)

func SearchTradeOrder(ctx context.Context, query string) (result []*order_business.SearchTradeOrderInfo, retCode int) {
	result = make([]*order_business.SearchTradeOrderInfo, 0)
	retCode = code.Success
	searchKey := "micro-mall-order:search-order:" + query
	err := kelvins.G2CacheEngine.Get(searchKey, 120, &result, func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()
		list, ret := searchTradeOrder(ctx, query)
		if ret != code.Success {
			return &list, fmt.Errorf("searchTradeOrder ret %v", ret)
		}
		return &list, nil
	})
	if err != nil {
		retCode = code.ErrorServer
		return
	}
	return
}

func searchTradeOrder(ctx context.Context, query string) (result []*order_business.SearchTradeOrderInfo, retCode int) {
	serverName := args.RpcServiceMicroMallSearch
	retCode = code.Success
	conn, err := util.GetGrpcClient(ctx, serverName)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetGrpcClient %v,err: %v", serverName, err)
		retCode = code.ErrorServer
		return
	}
	client := search_business.NewSearchBusinessServiceClient(conn)
	rsp, err := client.TradeOrderSearch(ctx, &search_business.TradeOrderSearchRequest{Query: query})
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "TradeOrderSearch err: %v query: %v", err, query)
		retCode = code.ErrorServer
		return
	}
	if rsp.Common.Code != search_business.RetCode_SUCCESS {
		kelvins.ErrLogger.Errorf(ctx, "TradeOrderSearch err: %v query: %v, rsp: %v", err, query, json.MarshalToStringNoError(rsp))
		retCode = code.ErrorServer
		return
	}
	if len(rsp.GetList()) == 0 {
		return
	}
	orderCodeList := make([]string, len(rsp.GetList()))
	for i := 0; i < len(rsp.GetList()); i++ {
		searchInfo := rsp.List[i]
		orderCodeList[i] = searchInfo.GetOrderCode()
	}
	orderList, err := repository.FindOrderListByOrderCode("order_code,shop_id,description,amount,money,pay_state,create_time,update_time", orderCodeList)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "FindOrderListByOrderCode err: %v orderCodeList: %v", err, json.MarshalToStringNoError(orderCodeList))
		retCode = code.ErrorServer
		return
	}
	if len(orderList) == 0 {
		return
	}
	orderCodeToOrder := map[string]mysql.Order{}
	for i := 0; i < len(orderList); i++ {
		orderCodeToOrder[orderList[i].OrderCode] = orderList[i]
	}
	orderSkuList, err := repository.FindOrderSkuByOrderCode("order_code,sku_code,name,price,amount", orderCodeList)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "FindOrderSkuByOrderCode err: %v orderCodeList: %v", err, json.MarshalToStringNoError(orderCodeList))
		retCode = code.ErrorServer
		return
	}
	if len(orderSkuList) == 0 {
		return
	}
	orderCodeToSku := map[string][]mysql.OrderSku{}
	for i := 0; i < len(orderSkuList); i++ {
		orderCodeToSku[orderSkuList[i].OrderCode] = append(orderCodeToSku[orderSkuList[i].OrderCode], orderSkuList[i])
	}
	orderSceneShopList, err := repository.FindOrderSceneShop("order_code,shop_name,shop_address", orderCodeList)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "FindOrderSceneShop err: %v orderCodeList: %v", err, json.MarshalToStringNoError(orderCodeList))
		retCode = code.ErrorServer
		return
	}
	orderCodeToSceneShop := map[string]mysql.OrderSceneShop{}
	for i := 0; i < len(orderSceneShopList); i++ {
		orderCodeToSceneShop[orderSceneShopList[i].OrderCode] = orderSceneShopList[i]
	}

	// 聚合数据
	for i := 0; i < len(rsp.List); i++ {
		if rsp.List[i].OrderCode == "" {
			continue
		}
		orderInfo := orderCodeToOrder[rsp.List[i].OrderCode]
		var payState string
		switch orderInfo.PayState {
		case 0:
			payState = "未支付"
		case 1:
			payState = "支付中"
		case 2:
			payState = "支付失败"
		case 3:
			payState = "支付成功"
		case 4:
			payState = "支付取消"
		}
		info := &order_business.SearchTradeOrderInfo{
			OrderCode:   orderInfo.OrderCode,
			ShopId:      orderInfo.ShopId,
			Money:       fmt.Sprintf("共计：%v", orderInfo.Money),
			Description: orderInfo.Description,
			CreateTime:  orderInfo.CreateTime.Format("2006-01-02 15:04:05"),
			PayState:    payState,
			PayTime:     orderInfo.UpdateTime.Format("2006-01-02 15:04:05"),
			ShopAddress: orderCodeToSceneShop[orderInfo.OrderCode].ShopAddress,
			ShopName:    orderCodeToSceneShop[orderInfo.OrderCode].ShopName,
			Score:       rsp.List[i].Score,
		}
		for _, v := range orderCodeToSku[rsp.List[i].OrderCode] {
			goods := &order_business.SearchTradeOrderGoods{
				GoodsName: v.Name,
				Price:     v.Price,
				SkuCode:   v.SkuCode,
				Amount:    int32(v.Amount),
			}
			info.Goods = append(info.Goods, goods)
		}
		result = append(result, info)
	}

	return
}
