package service

import (
	"context"
	"gitee.com/cristiane/micro-mall-order/model/args"
	"gitee.com/cristiane/micro-mall-order/model/mysql"
	"gitee.com/cristiane/micro-mall-order/pkg/code"
	"gitee.com/cristiane/micro-mall-order/pkg/util"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_order_proto/order_business"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_sku_proto/sku_business"
	"gitee.com/cristiane/micro-mall-order/proto/micro_mall_users_proto/users"
	"gitee.com/cristiane/micro-mall-order/repository"
	"gitee.com/cristiane/micro-mall-order/vars"
	"gitee.com/kelvins-io/common/errcode"
	"gitee.com/kelvins-io/common/json"
	"gitee.com/kelvins-io/kelvins"
	"github.com/shopspring/decimal"
	"time"
)

func CreateOrder(ctx context.Context, req *order_business.CreateOrderRequest) (result *args.CreateOrderRsp, retCode int) {
	var err error
	result = &args.CreateOrderRsp{
		OrderEntryList: make([]args.OrderEntry, len(req.Detail.ShopDetail)),
	}
	retCode = code.Success
	// 检查用户
	serverName := args.RpcServiceMicroMallUsers
	conn, err := util.GetGrpcClient(serverName)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetGrpcClient %v,err: %v", serverName, err)
		retCode = code.ErrorServer
		return
	}
	defer conn.Close()
	client := users.NewUsersServiceClient(conn)
	r := users.GetUserInfoRequest{
		Uid: req.Uid,
	}
	rsp, err := client.GetUserInfo(ctx, &r)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetUserInfo %v,err: %v", serverName, err)
		retCode = code.ErrorServer
		return
	}
	if rsp == nil || rsp.Common.Code != users.RetCode_SUCCESS {
		retCode = code.ErrorServer
		return
	}
	if rsp.Info == nil || rsp.Info.Uid <= 0 {
		retCode = code.UserNotExist
		return
	}
	// 初始订单和订单明细
	shops := req.Detail.ShopDetail
	orderList := make([]mysql.Order, len(shops))
	orderSkuList := make([]mysql.OrderSku, 0)
	tradeOrderDetail := make([]args.TradeOrderDetail, len(shops))
	deductInventoryList := make([]*sku_business.DeductEntryShop, 0)
	for i := 0; i < len(shops); i++ {
		orderCode := util.GetUUID()
		totalAmount := decimal.NewFromInt(0)
		deductEntryShop := &sku_business.DeductEntryShop{
			ShopId: shops[i].ShopId,
			Detail: nil,
		}
		deductEntryList := make([]*sku_business.DeductEntryDetail, 0)
		for j := 0; j < len(shops[i].Goods); j++ {
			goods := shops[i].Goods[j]
			price, err := decimal.NewFromString(shops[i].Goods[j].Price)
			if err != nil {
				kelvins.ErrLogger.Errorf(ctx, "decimal NewFromString err: %v, data: %v", err, shops[i].Goods[j].Price)
				retCode = code.ErrorServer
				return
			}
			temp := util.DecimalMul(price, decimal.NewFromInt(shops[i].Goods[j].Amount))
			totalAmount = util.DecimalAdd(totalAmount, temp)
			orderSku := mysql.OrderSku{
				OrderCode:  orderCode,
				ShopId:     shops[i].ShopId,
				SkuCode:    goods.SkuCode,
				Price:      goods.Price,
				Amount:     int(goods.Amount),
				Name:       goods.Name,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			}
			deductEntry := &sku_business.DeductEntryDetail{
				SkuCode: goods.SkuCode,
				Amount:  goods.Amount,
			}
			deductEntryList = append(deductEntryList, deductEntry)
			orderSkuList = append(orderSkuList, orderSku)
		}
		deductEntryShop.Detail = deductEntryList
		payExpire := time.Now().Add(30 * time.Minute)
		order := mysql.Order{
			OrderCode:    orderCode,
			Uid:          req.Uid,
			OrderTime:    time.Now(),
			Description:  req.Description,
			ClientIp:     req.PayerClientIp,
			DeviceCode:   req.DeviceId,
			ShopId:       shops[i].ShopId,
			ShopName:     shops[i].SceneInfo.StoreInfo.Name,
			ShopAreaCode: shops[i].SceneInfo.StoreInfo.AreaCode,
			ShopAddress:  shops[i].SceneInfo.StoreInfo.Address,
			State:        0,
			PayExpire:    payExpire,
			PayState:     0,
			Amount:       len(shops[i].Goods),
			TotalAmount:  totalAmount.String(),
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		orderList[i] = order
		result.OrderEntryList[i] = args.OrderEntry{
			OrderCode:  orderCode,
			TimeExpire: util.ParseTimeOfStr(payExpire.Unix()),
		}
		tradeOrderDetail[i] = args.TradeOrderDetail{
			ShopId:    shops[i].ShopId,
			OrderCode: orderCode,
		}
		deductInventoryList = append(deductInventoryList, deductEntryShop)
	}

	// 扣减库存
	serverName = args.RpcServiceMicroMallSku
	conn, err = util.GetGrpcClient(serverName)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "GetGrpcClient %v,err: %v", serverName, err)
		retCode = code.ErrorServer
		return
	}
	defer conn.Close()
	skuSer := sku_business.NewSkuBusinessServiceClient(conn)
	skuR := sku_business.DeductInventoryRequest{
		List: deductInventoryList,
	}
	skuRsp, err := skuSer.DeductInventory(ctx, &skuR)
	if err != nil {
		kelvins.ErrLogger.Errorf(ctx, "DeductInventory %v,err: %v", serverName, err)
		retCode = code.ErrorServer
		return
	}
	if skuRsp == nil || skuRsp.Common == nil || skuRsp.Common.Code == sku_business.RetCode_ERROR {
		retCode = code.ErrorServer
		return
	}
	if skuRsp.Common.Code == sku_business.RetCode_SKU_AMOUNT_NOT_ENOUGH {
		retCode = code.SkuAmountNotEnough
		return
	}

	tx := kelvins.XORM_DBEngine.NewSession()
	// 创建订单
	err = repository.CreateOrder(tx, orderList)
	if err != nil {
		tx.Rollback()
		kelvins.ErrLogger.Errorf(ctx, "CreateOrder err: %v, orderList: %+v", err, orderList)
		retCode = code.ErrorServer
		return
	}
	// 创建订单明细
	err = repository.CreateOrderSku(tx, orderSkuList)
	if err != nil {
		tx.Rollback()
		kelvins.ErrLogger.Errorf(ctx, "CreateOrderSku err: %v, orderSkuList: %+v", err, orderSkuList)
		retCode = code.ErrorServer
		return
	}
	tx.Commit()

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
			Detail: tradeOrderDetail,
		}),
	}
	taskUUID, retCode := pushSer.PushMessage(ctx, businessMsg)
	if retCode != code.Success {
		kelvins.ErrLogger.Errorf(ctx, "trade order businessMsg: %+v  notice send err: ", businessMsg, errcode.GetErrMsg(retCode))
	} else {
		kelvins.BusinessLogger.Infof(ctx, "trade order businessMsg businessMsg: %+v  taskUUID :%v", businessMsg, taskUUID)
	}

	return
}
