package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"gitee.com/cristiane/micro-mall-order/model/args"
	"gitee.com/cristiane/micro-mall-order/model/mysql"
	"gitee.com/cristiane/micro-mall-order/pkg/code"
	"gitee.com/cristiane/micro-mall-order/pkg/util"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_order_proto/order_business"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_search_proto/search_business"
	"gitee.com/cristiane/micro-mall-order/repository"
	"gitee.com/kelvins-io/common/hash"
	"gitee.com/kelvins-io/common/json"
	"gitee.com/kelvins-io/kelvins"
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

type Int64Slice []int64

func (x Int64Slice) Len() int           { return len(x) }
func (x Int64Slice) Less(i, j int) bool { return x[i] < x[j] }
func (x Int64Slice) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x Int64Slice) Sort()              { sort.Sort(x) }

func OrderShopRank(ctx context.Context, req *order_business.OrderShopRankRequest) (list []*order_business.OrderShopRankEntry, retCode int) {
	retCode = code.Success
	list = make([]*order_business.OrderShopRankEntry, 0)
	sort.Sort((Int64Slice)(req.Option.ShopId))
	sort.Sort((Int64Slice)(req.Option.Uid))
	shopIdKey := strings.Builder{}
	for _, v := range req.Option.ShopId {
		if v != 0 {
			shopIdKey.WriteString(fmt.Sprintf("%v", v))
		}
	}
	uidKey := strings.Builder{}
	for _, v := range req.Option.Uid {
		if v != 0 {
			uidKey.WriteString(fmt.Sprintf("%v", v))
		}
	}
	cacheKey := fmt.Sprintf("micro-mall-order:order-shop-rank:%v-%v-%v-%v-%v-%v",
		shopIdKey.String(), uidKey.String(), req.Option.StartTime, req.Option.EndTime, req.PageMeta.PageSize, req.PageMeta.PageNum)

	if len(cacheKey) > 512 {
		cacheKey = hash.MD5EncodeToString(cacheKey)
	}
	err := kelvins.G2CacheEngine.Get(cacheKey, 5, &list, func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()
		result, ret := orderShopRank(ctx, req)
		if ret != code.Success {
			return result, fmt.Errorf("%v", ret)
		}
		return result, nil
	})
	if err != nil {
		retCode = code.ErrorServer
		return nil, 0
	}
	return
}

const orderShopRankSql = "uid, shop_id, sum(amount) as s_amount,sum(money) as s_money"

func orderShopRank(ctx context.Context, req *order_business.OrderShopRankRequest) (list []*order_business.OrderShopRankEntry, retCode int) {
	retCode = code.Success
	orderRankWhere := map[string]interface{}{
		"state":            0, // 有效
		"pay_state":        3, // 已支付
		"inventory_verify": 1, // 库存核实
	}
	if len(req.GetOption().GetShopId()) > 0 {
		orderRankWhere["shop_id"] = req.GetOption().GetShopId()
	}
	if len(req.GetOption().GetUid()) > 0 {
		orderRankWhere["uid"] = req.GetOption().GetUid()
	}
	inOrder := []string{"s_money", "s_amount"}
	groupBy := "shop_id,uid"
	var err error
	var startTime time.Time
	var endTime time.Time
	if req.GetOption() != nil && req.GetOption().GetStartTime() != "" {
		startTime, err = util.GenTimeOfStr(req.GetOption().GetStartTime())
		if err != nil {
			retCode = code.InvalidParamTimeFormat
			return
		}
		startTime = startTime.UTC()
	}
	if req.GetOption() != nil && req.GetOption().GetEndTime() != "" {
		endTime, err = util.GenTimeOfStr(req.GetOption().GetEndTime())
		if err != nil {
			retCode = code.InvalidParamTimeFormat
			return
		}
		endTime = endTime.UTC()
	}
	ranks, err := repository.OrderShopRank(orderShopRankSql, orderRankWhere, groupBy, inOrder, startTime, endTime, req.GetPageMeta().GetPageSize(), req.GetPageMeta().GetPageNum())
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "OrderShopRank err: %v where: %+v", err, orderRankWhere)
		retCode = code.ErrorServer
		return
	}
	if len(ranks) == 0 {
		return
	}
	for _, v := range ranks {
		e := &order_business.OrderShopRankEntry{
			ShopId: v.ShopId,
			Uid:    v.Uid,
			Money:  v.SMoney,
			Amount: int64(v.SAmount),
		}
		list = append(list, e)
	}
	return
}

func OrderSkuRank(ctx context.Context, req *order_business.OrderSkuRankRequest) (list []*order_business.OrderSkuRankEntry, retCode int) {
	retCode = code.Success
	list = make([]*order_business.OrderSkuRankEntry, 0)
	sort.Sort((Int64Slice)(req.Option.ShopId))
	sort.Strings(req.Option.SkuCode)
	sort.Strings(req.Option.GoodsName)
	shopIdKey := strings.Builder{}
	for _, v := range req.Option.ShopId {
		if v != 0 {
			shopIdKey.WriteString(fmt.Sprintf("%v", v))
		}
	}
	skuCodeKey := strings.Builder{}
	for _, v := range req.Option.SkuCode {
		if v != "" {
			skuCodeKey.WriteString(fmt.Sprintf("%v", v))
		}
	}
	goodsNameKey := strings.Builder{}
	for _, v := range req.Option.GoodsName {
		if v != "" {
			goodsNameKey.WriteString(fmt.Sprintf("%v", v))
		}
	}
	cacheKey := fmt.Sprintf("micro-mall-order:order-sku-rank:%v-%v-%v-%v-%v-%v-%v",
		shopIdKey.String(), skuCodeKey.String(), goodsNameKey.String(), req.Option.StartTime, req.Option.EndTime, req.PageMeta.PageSize, req.PageMeta.PageNum)

	if len(cacheKey) > 512 {
		cacheKey = hash.MD5EncodeToString(cacheKey)
	}
	err := kelvins.G2CacheEngine.Get(cacheKey, 5, &list, func() (interface{}, error) {
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()
		result, ret := orderSkuRank(ctx, req)
		if ret != code.Success {
			return result, fmt.Errorf("%v", ret)
		}
		return result, nil
	})
	if err != nil {
		retCode = code.ErrorServer
		return nil, 0
	}
	return
}

const orderSkuRankSql = "sku_code,name,shop_id ,sum(amount) as s_amount"

func orderSkuRank(ctx context.Context, req *order_business.OrderSkuRankRequest) (list []*order_business.OrderSkuRankEntry, retCode int) {
	retCode = code.Success
	orderRankWhere := map[string]interface{}{}
	if len(req.GetOption().GetShopId()) > 0 {
		orderRankWhere["shop_id"] = req.GetOption().GetShopId()
	}
	if len(req.GetOption().GetSkuCode()) > 0 {
		orderRankWhere["sku_code"] = req.GetOption().GetSkuCode()
	}
	if len(req.GetOption().GetGoodsName()) > 0 {
		orderRankWhere["name"] = req.GetOption().GetGoodsName()
	}
	inOrder := []string{"s_amount"}
	groupBy := "shop_id,sku_code,name"
	var err error
	var startTime time.Time
	var endTime time.Time
	if req.GetOption() != nil && req.GetOption().GetStartTime() != "" {
		startTime, err = util.GenTimeOfStr(req.GetOption().GetStartTime())
		if err != nil {
			retCode = code.InvalidParamTimeFormat
			return
		}
		startTime = startTime.UTC()
	}
	if req.GetOption() != nil && req.GetOption().GetEndTime() != "" {
		endTime, err = util.GenTimeOfStr(req.GetOption().GetEndTime())
		if err != nil {
			retCode = code.InvalidParamTimeFormat
			return
		}
		endTime = endTime.UTC()
	}

	ranks, err := repository.OrderSkuRank(orderSkuRankSql, orderRankWhere, groupBy, inOrder, startTime, endTime, req.GetPageMeta().GetPageSize(), req.GetPageMeta().GetPageNum())
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "OrderSkuRank err: %v, where: %+v", err, orderRankWhere)
		retCode = code.ErrorServer
		return
	}
	if len(ranks) == 0 {
		return
	}
	for _, v := range ranks {
		e := &order_business.OrderSkuRankEntry{
			ShopId:    v.ShopId,
			SkuCode:   v.SkuCode,
			GoodsName: v.Name,
			Amount:    int64(v.SAmount),
		}
		list = append(list, e)
	}
	return
}
