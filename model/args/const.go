package args

type CreateOrderRsp struct {
	OrderEntryList []OrderEntry
}

type OrderEntry struct {
	OrderCode  string `json:"order_code"`
	TimeExpire string `json:"time_expire"`
}

const (
	RpcServiceMicroMallUsers = "micro-mall-users"
	RpcServiceMicroMallShop  = "micro-mall-shop"
)

const (
	TaskNameTradeOrderNotice    = "task_trade_order_notice"
	TaskNameTradeOrderNoticeErr = "task_trade_order_notice_err"
)

type CommonBusinessMsg struct {
	Type int    `json:"type"`
	Tag  string `json:"tag"`
	UUID string `json:"uuid"`
	Msg  string `json:"msg"`
}

type TradeOrderDetail struct {
	ShopId    int64  `json:"shop_id"`
	OrderCode string `json:"order_code"`
}

type TradeOrderNotice struct {
	Uid    int64              `json:"uid"`
	Time   string             `json:"time"`
	Detail []TradeOrderDetail `json:"detail"`
}

const (
	Unknown                   = 0
	TradeOrderEventTypeCreate = 10014
	TradeOrderEventTypeExpire = 10015
)

var MsgFlags = map[int]string{
	Unknown:                   "未知",
	TradeOrderEventTypeCreate: "交易订单创建",
	TradeOrderEventTypeExpire: "交易订单过期",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[Unknown]
}
