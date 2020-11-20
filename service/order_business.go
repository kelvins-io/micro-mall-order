package service

import (
	"context"
	"gitee.com/cristiane/micro-mall-order/model/args"
	"gitee.com/cristiane/micro-mall-order/model/mysql"
	"gitee.com/cristiane/micro-mall-order/pkg/code"
	"gitee.com/cristiane/micro-mall-order/pkg/util"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_order_proto/order_business"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_shop_proto/shop_business"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_sku_proto/sku_business"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_users_proto/users"
	"gitee.com/cristiane/micro-mall-order/repository"
	"gitee.com/cristiane/micro-mall-order/vars"
	"gitee.com/kelvins-io/common/errcode"
	"gitee.com/kelvins-io/common/json"
	"gitee.com/kelvins-io/kelvins"
	"github.com/shopspring/decimal"
	"time"
	"xorm.io/xorm"
)

func GenOrderTxCode(ctx context.Context, req *order_business.GenOrderTxCodeRequest) (orderTxCode string, retCode int) {
	return util.GetUUID(), code.Success
}

func CheckOrderExist(ctx context.Context, req *order_business.CheckOrderExistRequest) (isExist bool, retCode int) {
	retCode = code.Success
	// 检查是否重复下单
	isExist, retCode = checkTradeOrderExist(ctx, req.OrderTxCode)
	if retCode != code.Success {
		return
	}
	if isExist {
		retCode = code.OrderExist
		return
	} else {
		retCode = code.OrderNotExist
	}

	return
}

func createOrderCheckPriceVersion(ctx context.Context, req *order_business.CreateOrderRequest) int {
	setList := make([]*sku_business.SkuPriceVersionSet, 0)
	for i := 0; i < len(req.Detail.ShopDetail); i++ {
		set := &sku_business.SkuPriceVersionSet{
			ShopId:    req.Detail.ShopDetail[i].ShopId,
			EntryList: nil,
		}
		entryList := make([]*sku_business.SkuPriceVersionEntry, 0)
		for j := 0; j < len(req.Detail.ShopDetail[i].Goods); j++ {
			entry := &sku_business.SkuPriceVersionEntry{
				SkuCode: req.Detail.ShopDetail[i].Goods[j].SkuCode,
				Price:   req.Detail.ShopDetail[i].Goods[j].Price,
				Version: req.Detail.ShopDetail[i].Goods[j].Version,
			}
			entryList = append(entryList, entry)
		}
		set.EntryList = entryList
		setList = append(setList, set)
	}
	serverName := args.RpcServiceMicroMallSku
	conn, err := util.GetGrpcClient(serverName)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetGrpcClient %v,err: %v", serverName, err)
		return code.ErrorServer
	}
	defer conn.Close()
	client := sku_business.NewSkuBusinessServiceClient(conn)
	filtrateReq := &sku_business.FiltrateSkuPriceVersionRequest{
		SetList:    setList,
		PolicyType: sku_business.SkuPricePolicyFiltrateType_VERSION_SECTION,
		LimitUpper: 3,
	}
	filtrateRsp, err := client.FiltrateSkuPriceVersion(ctx, filtrateReq)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetGrpcClient %v,err: %v, req: %+v", serverName, err, filtrateReq)
		return code.ErrorServer
	}
	if filtrateRsp.Common.Code == sku_business.RetCode_SUCCESS {
		return code.Success
	}
	kelvins.ErrLogger.Errorf(ctx, "FiltrateSkuPriceVersion req: %+v, rsp: %+v", filtrateReq, filtrateRsp)
	switch filtrateRsp.Common.Code {
	case sku_business.RetCode_SKU_PRICE_VERSION_NOT_EXIST:
		return code.SkuPriceVersionNotExist
	case sku_business.RetCode_ERROR:
		return code.ErrorServer
	}
	return code.ErrorServer
}

func createOrderCheckUserDelivery(ctx context.Context, req *order_business.CreateOrderRequest) int {
	if req.Uid <= 0 {
		return code.UserNotExist
	}
	if req.OrderTxCode == "" {
		return code.OrderTxCodeEmpty
	}
	if req.DeliveryInfo == nil || req.DeliveryInfo.UserDeliveryId <= 0 {
		return code.OrderDeliveryNotExist
	}
	// 检查订单交付信息是否存在
	serverName := args.RpcServiceMicroMallUsers
	conn, err := util.GetGrpcClient(serverName)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetGrpcClient %v,err: %v", serverName, err)
		return code.ErrorServer
	}
	defer conn.Close()
	userClient := users.NewUsersServiceClient(conn)
	userDeliveryInfoReq := &users.GetUserDeliveryInfoRequest{
		Uid:            req.Uid,
		UserDeliveryId: req.DeliveryInfo.UserDeliveryId,
	}
	userDeliveryInfoRsp, err := userClient.GetUserDeliveryInfo(ctx, userDeliveryInfoReq)
	if err != nil || userDeliveryInfoRsp.Common.Code != users.RetCode_SUCCESS {
		kelvins.ErrLogger.Errorf(ctx, "GetGrpcClient %v,err: %v", serverName, err)
		return code.ErrorServer
	}
	if userDeliveryInfoRsp.Info[0].Id <= 0 {
		return code.OrderDeliveryNotExist
	}

	return code.Success
}

func CreateOrder(ctx context.Context, req *order_business.CreateOrderRequest) (result *args.CreateOrderRsp, retCode int) {
	result = &args.CreateOrderRsp{
		TxCode: "",
	}
	retCode = code.Success
	// 参数检查
	retCode = createOrderCheckUserDelivery(ctx, req)
	if retCode != code.Success {
		return
	}
	// 检查用户身份
	retCode = tradeOrderCheckUserIdentity(ctx, req)
	if retCode != code.Success {
		return
	}
	// 检查是否重复下单
	isExist, retCode := checkTradeOrderExist(ctx, req.OrderTxCode)
	if retCode != code.Success {
		return
	}
	if isExist {
		retCode = code.OrderExist
		result.TxCode = req.OrderTxCode
		return
	}
	// 检查商品价格版本是否符合预期
	retCode = createOrderCheckPriceVersion(ctx, req)
	if retCode != code.Success {
		return
	}
	//txCode := util.GetUUID()
	txCode := req.OrderTxCode
	// 扣减库存，构造订单
	orderList, orderSkuList, orderSceneShopList, retCode := tradeOrderDeductInventory(ctx, req, txCode)
	if retCode != code.Success {
		return
	}
	tx := kelvins.XORM_DBEngine.NewSession()
	err := tx.Begin()
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "CreateOrder Begin err: %v", err)
		retCode = code.ErrorServer
		return
	}
	// 创建订单
	retCode = tradeOrderCreate(ctx, tx, orderList, orderSkuList, orderSceneShopList)
	if retCode != code.Success {
		errRollback := tx.Rollback()
		if errRollback != nil {
			kelvins.ErrLogger.Errorf(ctx, "CreateOrder Rollback err: %v", errRollback)
		}
		return
	}
	result.TxCode = txCode
	// 触发订单事件
	retCode = tradeOrderEventNotice(ctx, req, txCode)
	if retCode != code.Success {
		errRollback := tx.Rollback()
		if errRollback != nil {
			kelvins.ErrLogger.Errorf(ctx, "CreateOrder Rollback err: %v", errRollback)
		}
		return
	}
	err = tx.Commit()
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "CreateOrder commit err: %v", err)
		retCode = code.ErrorServer
		return
	}
	return
}

func checkTradeOrderExist(ctx context.Context, orderTxCode string) (bool, int) {
	if orderTxCode == "" {
		return true, code.OrderTxCodeEmpty
	}
	orderOne, err := repository.GetOrderExist(orderTxCode)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetOrderExist  err: %v, txCode: %v", err, orderTxCode)
		return true, code.ErrorServer
	}
	if orderOne.OrderCode != "" {
		return true, code.Success
	}
	return false, code.Success
}

func tradeOrderCreate(ctx context.Context, tx *xorm.Session, orderList []mysql.Order, orderSkuList []mysql.OrderSku, orderSceneShopList []mysql.OrderSceneShop) int {
	// 创建订单
	err := repository.CreateOrder(tx, orderList)
	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			kelvins.ErrLogger.Errorf(ctx, "CreateOrder Rollback err: %v", errRollback)
		}
		kelvins.ErrLogger.Errorf(ctx, "CreateOrder err: %v, orderList: %+v", err, orderList)
		return code.ErrorServer
	}
	// 创建订单明细
	err = repository.CreateOrderSku(tx, orderSkuList)
	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			kelvins.ErrLogger.Errorf(ctx, "CreateOrder Rollback err: %v", errRollback)
		}
		kelvins.ErrLogger.Errorf(ctx, "CreateOrderSku err: %v, orderSkuList: %+v", err, orderSkuList)
		return code.ErrorServer
	}
	// 创建订单场景信息
	err = repository.CreateOrderSceneShop(tx, orderSceneShopList)
	if err != nil {
		errRollback := tx.Rollback()
		if errRollback != nil {
			kelvins.ErrLogger.Errorf(ctx, "CreateOrderSceneShop Rollback err: %v", errRollback)
		}
		kelvins.ErrLogger.Errorf(ctx, "CreateOrderSceneShop err: %v, orderSceneShopList: %+v", err, orderSceneShopList)
		return code.ErrorServer
	}
	return code.Success
}

func tradeOrderDeductInventory(ctx context.Context, req *order_business.CreateOrderRequest, txCode string) ([]mysql.Order, []mysql.OrderSku, []mysql.OrderSceneShop, int) {
	// 初始订单和订单明细
	shops := req.Detail.ShopDetail
	orderList := make([]mysql.Order, len(shops))
	orderSkuList := make([]mysql.OrderSku, 0)
	orderSceneShopList := make([]mysql.OrderSceneShop, 0)
	deductInventoryList := make([]*sku_business.InventoryEntryShop, 0)
	for i := 0; i < len(shops); i++ {
		orderCode := util.GetUUID()
		totalAmount := decimal.NewFromInt(0)
		deductEntryShop := &sku_business.InventoryEntryShop{
			ShopId:     shops[i].ShopId,
			OutTradeNo: orderCode,
			Detail:     nil,
		}
		deductEntryList := make([]*sku_business.InventoryEntryDetail, 0)
		var skuAmount int64
		for j := 0; j < len(shops[i].Goods); j++ {
			// 统计订单包含商品个数
			skuAmount += shops[i].Goods[j].Amount
			goods := shops[i].Goods[j]
			if shops[i].Goods[j].Price == "" {
				shops[i].Goods[j].Price = "0"
			}
			price, err := decimal.NewFromString(shops[i].Goods[j].Price)
			if err != nil {
				kelvins.ErrLogger.Errorf(ctx, "decimal NewFromString err: %v, Price: %v", err, shops[i].Goods[j].Price)
				return nil, nil, nil, code.ErrorServer
			}
			if shops[i].Goods[j].Reduction == "" {
				shops[i].Goods[j].Reduction = "0"
			}
			reduction, err := decimal.NewFromString(shops[i].Goods[j].Reduction)
			if err != nil {
				kelvins.ErrLogger.Errorf(ctx, "decimal NewFromString err: %v, Reduction: %v", err, shops[i].Goods[j].Reduction)
				return nil, nil, nil, code.ErrorServer
			}
			price = util.DecimalSub(price, reduction)
			temp := util.DecimalMul(price, decimal.NewFromInt(shops[i].Goods[j].Amount))
			totalAmount = util.DecimalAdd(totalAmount, temp)
			orderSku := mysql.OrderSku{
				OrderCode:  orderCode,
				ShopId:     shops[i].ShopId,
				SkuCode:    goods.SkuCode,
				Price:      price.String(), // 满减后的价格
				Amount:     int(goods.Amount),
				Name:       goods.Name,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			}
			deductEntry := &sku_business.InventoryEntryDetail{
				SkuCode: goods.SkuCode,
				Amount:  goods.Amount,
			}
			deductEntryList = append(deductEntryList, deductEntry)
			orderSkuList = append(orderSkuList, orderSku)
		}
		deductEntryShop.Detail = deductEntryList
		payExpire := time.Now().Add(30 * time.Minute)
		order := mysql.Order{
			TxCode:              txCode, // 同一个批次下单的订单对应同一个交易code
			OrderCode:           orderCode,
			Uid:                 req.Uid,
			OrderTime:           time.Now(),
			Description:         req.Description,
			ClientIp:            req.PayerClientIp,
			DeviceCode:          req.DeviceId,
			ShopId:              shops[i].ShopId,
			State:               0,
			PayExpire:           payExpire,
			PayState:            1,
			Amount:              int(skuAmount),
			TotalAmount:         totalAmount.String(),
			CoinType:            int(shops[i].CoinType),
			LogisticsDeliveryId: int(req.DeliveryInfo.UserDeliveryId),
			CreateTime:          time.Now(),
			UpdateTime:          time.Now(),
		}
		orderSceneShop := mysql.OrderSceneShop{
			OrderCode:    orderCode,
			ShopId:       shops[i].ShopId,
			ShopName:     shops[i].SceneInfo.StoreInfo.Name,
			ShopAreaCode: shops[i].SceneInfo.StoreInfo.AreaCode,
			ShopAddress:  shops[i].SceneInfo.StoreInfo.Address,
			CreateTime:   time.Time{},
			UpdateTime:   time.Time{},
		}
		orderSceneShopList = append(orderSceneShopList, orderSceneShop)
		orderList[i] = order
		deductInventoryList = append(deductInventoryList, deductEntryShop)
	}

	// 扣减库存
	serverName := args.RpcServiceMicroMallSku
	conn, err := util.GetGrpcClient(serverName)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetGrpcClient %v,err: %v", serverName, err)
		return nil, nil, nil, code.ErrorServer
	}
	defer conn.Close()
	skuClient := sku_business.NewSkuBusinessServiceClient(conn)
	skuR := sku_business.DeductInventoryRequest{
		List: deductInventoryList,
		OperationMeta: &sku_business.OperationMeta{
			OpUid: req.Uid,
			OpIp:  req.PayerClientIp,
		},
	}
	skuRsp, err := skuClient.DeductInventory(ctx, &skuR)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "DeductInventory %v,err: %v", serverName, err)
		return nil, nil, nil, code.ErrorServer
	}
	if skuRsp == nil || skuRsp.Common == nil || skuRsp.Common.Code == sku_business.RetCode_ERROR {
		return nil, nil, nil, code.ErrorServer
	}
	if skuRsp.Common.Code == sku_business.RetCode_SKU_AMOUNT_NOT_ENOUGH {
		return nil, nil, nil, code.SkuAmountNotEnough
	}
	if skuRsp.Common.Code == sku_business.RetCode_TRANSACTION_FAILED {
		return nil, nil, nil, code.TransactionFailed
	}
	return orderList, orderSkuList, orderSceneShopList, code.Success
}

func tradeOrderCheckUserIdentity(ctx context.Context, req *order_business.CreateOrderRequest) int {
	// 检查用户
	serverName := args.RpcServiceMicroMallUsers
	conn, err := util.GetGrpcClient(serverName)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetGrpcClient %v,err: %v", serverName, err)
		return code.ErrorServer
	}
	defer conn.Close()
	client := users.NewUsersServiceClient(conn)
	r := users.GetUserInfoRequest{
		Uid: req.Uid,
	}
	rsp, err := client.GetUserInfo(ctx, &r)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetUserInfo %v,err: %v", serverName, err)
		return code.ErrorServer
	}
	if rsp == nil || rsp.Common.Code != users.RetCode_SUCCESS {
		return code.ErrorServer
	}
	if rsp.Info == nil || rsp.Info.Uid <= 0 {
		return code.UserNotExist
	}
	return code.Success
}

func tradeOrderEventNotice(ctx context.Context, req *order_business.CreateOrderRequest, txCode string) int {
	// 触发订单事件
	pushSer := NewPushNoticeService(vars.TradeOrderQueueServer, PushMsgTag{
		DeliveryTag:    args.TaskNameTradeOrderNotice,
		DeliveryErrTag: args.TaskNameTradeOrderNoticeErr,
		RetryCount:     kelvins.QueueAMQPSetting.TaskRetryCount,
		RetryTimeout:   kelvins.QueueAMQPSetting.TaskRetryTimeout,
	})
	businessMsg := args.CommonBusinessMsg{
		Type: args.TradeOrderEventTypeCreate,
		Tag:  args.GetMsg(args.TradeOrderEventTypeCreate),
		UUID: util.GetUUID(),
		Msg: json.MarshalToStringNoError(args.TradeOrderNotice{
			Uid:    req.Uid,
			Time:   util.ParseTimeOfStr(time.Now().Unix()),
			TxCode: txCode,
		}),
	}
	taskUUID, retCode := pushSer.PushMessage(ctx, businessMsg)
	if retCode != code.Success {
		kelvins.ErrLogger.Errorf(ctx, "trade order businessMsg: %+v  notice send err: ", businessMsg, errcode.GetErrMsg(retCode))
	} else {
		kelvins.BusinessLogger.Infof(ctx, "trade order businessMsg businessMsg: %+v  taskUUID :%v", businessMsg, taskUUID)
	}
	return retCode
}

func GetOrderDetail(ctx context.Context, req *order_business.GetOrderDetailRequest) (result *args.OrderDetailRsp, retCode int) {
	result = &args.OrderDetailRsp{}
	result.List = make([]args.ShopOrderDetail, 0)
	retCode = code.Success
	// 通过交易号获取订单详细
	where := map[string]interface{}{
		"tx_code": req.TxCode, // 订单事务号
		//"state":     0,           // 有效
		//"pay_state": []int{0, 2}, // 支付失败或未支付
	}
	orderList, err := repository.GetOrderListByTxCode(where)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetOrderListByTxCode err: %v, where: %+v", err, where)
		retCode = code.ErrorServer
		return
	}
	if len(orderList) <= 0 {
		return
	}
	for i := 0; i < len(orderList); i++ {
		if orderList[i].State != 0 {
			retCode = code.OrderStateInvalid
			return
		}
		if orderList[i].PayState != 0 && orderList[i].PayState != 2 {
			retCode = code.OrderPayCompleted
			return
		}
		if orderList[i].PayExpire.Sub(time.Now()) <= 0 {
			retCode = code.OrderExpire
			return
		}
	}
	uid := orderList[0].Uid
	coinType := orderList[0].CoinType
	// 获取订单用户code
	serverName := args.RpcServiceMicroMallUsers
	conn, err := util.GetGrpcClient(serverName)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetGrpcClient %v,err: %v", serverName, err)
		retCode = code.ErrorServer
		return
	}
	defer conn.Close()
	serve := users.NewUsersServiceClient(conn)
	r := users.GetUserInfoRequest{
		Uid: uid,
	}
	rsp, err := serve.GetUserInfo(ctx, &r)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetUserInfo %v,err: %v", serverName, err)
		retCode = code.ErrorServer
		return
	}
	if rsp == nil || rsp.Common == nil || rsp.Common.Code == users.RetCode_ERROR {
		kelvins.ErrLogger.Errorf(ctx, "GetUserInfo %v, rsp: %v", serverName, rsp.Common.Msg)
		retCode = code.ErrorServer
		return
	}
	if rsp.Common.Code == users.RetCode_USER_EXIST {
		retCode = code.UserNotExist
		return
	}
	if rsp.Info.AccountId == "" {
		retCode = code.UserExist
		return
	}
	result.UserCode = rsp.Info.AccountId
	result.CoinType = coinType
	// 获取店铺code
	shopIdList := make([]int64, len(orderList))
	for i := 0; i < len(orderList); i++ {
		shopIdList[i] = orderList[i].ShopId
	}
	serverName = args.RpcServiceMicroMallShop
	conn, err = util.GetGrpcClient(serverName)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetGrpcClient %v,err: %v", serverName, err)
		retCode = code.ErrorServer
		return
	}
	defer conn.Close()
	serveShop := shop_business.NewShopBusinessServiceClient(conn)
	rShop := shop_business.GetShopInfoRequest{
		ShopIds: shopIdList,
	}
	rspShop, err := serveShop.GetShopInfo(ctx, &rShop)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetShopInfo %v,err: %v", serverName, err)
		retCode = code.ErrorServer
		return
	}
	if rspShop == nil || rspShop.Common == nil || rspShop.Common.Code == shop_business.RetCode_ERROR {
		kelvins.ErrLogger.Errorf(ctx, "GetShopInfo %v,rspShop: %v", serverName, rspShop.Common.Code)
		retCode = code.ErrorServer
		return
	}
	// 店铺ID和店铺code映射关系
	shopIdToShopCode := make(map[int64]string)
	for i := 0; i < len(rspShop.InfoList); i++ {
		shopIdToShopCode[rspShop.InfoList[i].ShopId] = rspShop.InfoList[i].ShopCode
	}
	//key := args.ConfigKvShopOrderNotifyUrl
	//config, err := repository.GetConfigKv(key)
	//if err != nil {
	//	kelvins.ErrLogger.Errorf(ctx, "GetConfigKv err: %v ,key: %v", err, key)
	//	retCode = code.ErrorServer
	//	return
	//}
	result.List = make([]args.ShopOrderDetail, len(orderList))
	for i := 0; i < len(orderList); i++ {
		detail := args.ShopOrderDetail{
			ShopCode:    shopIdToShopCode[orderList[i].ShopId],
			OrderCode:   orderList[i].OrderCode,
			TimeExpire:  util.ParseTimeOfStr(orderList[i].PayExpire.Unix()),
			Description: orderList[i].Description,
			Amount:      orderList[i].TotalAmount,
			CoinType:    orderList[i].CoinType,
			//NotifyUrl:   config.ConfigValue,
		}
		result.List[i] = detail
	}
	return
}

func GetOrderSku(ctx context.Context, req *order_business.GetOrderSkuRequest) (*args.OrderSkuRsp, int) {
	result := &args.OrderSkuRsp{SkuList: make([]args.OrderSku, 0)}
	retCode := code.Success
	where := map[string]interface{}{
		"tx_code":   req.TxCode,  // 订单事务号
		"state":     0,           // 有效
		"pay_state": []int{0, 2}, // 支付失败或未支付
	}
	orderList, err := repository.GetOrderListByTxCode(where)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetOrderListByTxCode err: %v ,tx-code: %v", err, req.TxCode)
		retCode = code.ErrorServer
		return result, retCode
	}
	result.SkuList = make([]args.OrderSku, len(orderList))
	for i := 0; i < len(orderList); i++ {
		orderSkuList, err := repository.GetOrderSkuListByOrderCode([]string{orderList[i].OrderCode})
		if err != nil {
			kelvins.ErrLogger.Errorf(ctx, "GetOrderSkuListByOrderCode err: %v ,orderCodeList: %v", err, orderList[i].OrderCode)
			retCode = code.ErrorServer
			return result, retCode
		}
		orderSku := args.OrderSku{
			OrderCode: orderList[i].OrderCode,
			SkuList:   make([]args.OrderSkuEntry, len(orderSkuList)),
		}
		for j := 0; j < len(orderSkuList); j++ {
			entry := args.OrderSkuEntry{
				SkuCode: orderSkuList[j].SkuCode,
				Amount:  orderSkuList[j].Amount,
				Name:    orderSkuList[j].Name,
				Price:   orderSkuList[j].Price,
			}
			orderSku.SkuList[j] = entry
		}
		result.SkuList[i] = orderSku
	}

	return result, retCode
}

func UpdateOrderState(ctx context.Context, req *order_business.UpdateOrderStateRequest) (retCode int) {
	retCode = code.Success
	tx := kelvins.XORM_DBEngine.NewSession()
	err := tx.Begin()
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "UpdateOrder Begin err: %v", err)
		retCode = code.ErrorServer
		return
	}
	for i := 0; i < len(req.GetEntryList()); i++ {
		row := req.EntryList[i]
		where := map[string]interface{}{
			"order_code": row.OrderCode,
		}
		maps := map[string]interface{}{
			"update_time": time.Now(),
			"state":       row.State,
			"pay_state":   row.PayState,
		}
		rowsAffected, err := repository.UpdateOrderByTx(tx, where, maps)
		if err != nil {
			errRollback := tx.Rollback()
			if errRollback != nil {
				kelvins.ErrLogger.Errorf(ctx, "UpdateOrder Rollback err: %v, where: %+v, maps: %+v", errRollback, where, maps)
			}
			kelvins.ErrLogger.Errorf(ctx, "UpdateOrder err: %v, where: %+v, maps: %+v", err, where, maps)
			retCode = code.ErrorServer
			return
		}
		if rowsAffected <= 0 {
			errRollback := tx.Rollback()
			if errRollback != nil {
				kelvins.ErrLogger.Errorf(ctx, "UpdateOrder Rollback err: %v, where: %+v, maps: %+v", errRollback, where, maps)
			}
			retCode = code.OperationNotEffect
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "UpdateOrder Commit err: %v", err)
		retCode = code.ErrorServer
		return
	}

	return retCode
}

func OrderTradeNotice(ctx context.Context, req *order_business.OrderTradeNoticeRequest) int {
	retCode := code.Success
	if req.OrderTxCode == "" {
		return code.OrderTxCodeEmpty
	}
	order, err := repository.GetOrderExist(req.OrderTxCode)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetOrderExist err: %v, txCode: %v", err, req.OrderTxCode)
		return code.ErrorServer
	}
	if order.OrderCode == "" {
		return code.OrderTxCodeNotExist
	}
	// 触发订单事件
	pushSer := NewPushNoticeService(vars.TradeOrderPayQueueServer, PushMsgTag{
		DeliveryTag:    args.TaskNameTradeOrderPayCallback,
		DeliveryErrTag: args.TaskNameTradeOrderPayCallbackErr,
		RetryCount:     kelvins.QueueAMQPSetting.TaskRetryCount,
		RetryTimeout:   kelvins.QueueAMQPSetting.TaskRetryTimeout,
	})
	businessMsg := args.CommonBusinessMsg{
		Type: args.TradeOrderEventTypePayCallback,
		Tag:  args.GetMsg(args.TradeOrderEventTypePayCallback),
		UUID: util.GetUUID(),
		Msg: json.MarshalToStringNoError(args.TradeOrderPayCallback{
			Uid:    req.Uid,
			TxCode: req.OrderTxCode,
			PayId:  req.PayId,
		}),
	}
	taskUUID, pushCode := pushSer.PushMessage(ctx, businessMsg)
	if pushCode != code.Success {
		kelvins.ErrLogger.Errorf(ctx, "trade order businessMsg: %+v  notice send err: ", businessMsg, errcode.GetErrMsg(retCode))
		return code.ErrorServer
	}
	kelvins.BusinessLogger.Infof(ctx, "trade order businessMsg businessMsg: %+v  taskUUID :%v", businessMsg, taskUUID)

	return retCode
}
