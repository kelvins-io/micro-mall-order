package args

type OrderSku struct {
	OrderCode string          `json:"order_code"`
	SkuList   []OrderSkuEntry `json:"sku_list"`
}

type OrderSkuEntry struct {
	SkuCode   string `json:"sku_code"`
	Amount    int    `json:"amount"`
	Name      string `json:"name"`
	Price     string `json:"price"`
	Reduction string `json:"reduction"`
}

type OrderSkuRsp struct {
	SkuList []OrderSku `json:"sku_list"`
}

type CreateOrderRsp struct {
	TxCode string
}

type ShopOrderDetail struct {
	ShopId      int64  `json:"shop_id"`
	OrderCode   string `json:"order_code"`
	Description string `json:"description"`
	Amount      string `json:"amount"`
}

type OrderDetailRsp struct {
	CoinType int               `json:"coin_type"`
	List     []ShopOrderDetail `json:"list"`
}

const (
	RpcServiceMicroMallUsers  = "micro-mall-users"
	RpcServiceMicroMallSku    = "micro-mall-sku"
	RpcServiceMicroMallSearch = "micro-mall-search"
)

const (
	TaskNameTradeOrderNotice    = "task_trade_order_notice"
	TaskNameTradeOrderNoticeErr = "task_trade_order_notice_err"
)

const (
	TaskNameTradeOrderPayCallback    = "task_trade_order_pay_callback"
	TaskNameTradeOrderPayCallbackErr = "task_trade_order_pay_callback_err"
)

const (
	TradeOrderInfoSearchNoticeTag    = "trade_order_info_search_notice"
	TradeOrderInfoSearchNoticeTagErr = "trade_order_info_search_notice_err"
)

const (
	TradeOrderInfoSearchNoticeType = 10001
)

type SearchTradeOrderInfo struct {
	Description   string                  `json:"description"`
	DeviceId      string                  `json:"device_id"`
	ShopOrderList []SearchTradeOrderEntry `json:"shop_order_list"`
}

type SearchTradeOrderEntry struct {
	Description string `json:"description"`
	DeviceId    string `json:"device_id"`
	ShopName    string `json:"shop_name"`
	ShopAddress string `json:"shop_address"`
	GoodsName   string `json:"goods_name"`
	OrderCode   string `json:"order_code"`
}

type CommonBusinessMsg struct {
	Type    int    `json:"type"`
	Tag     string `json:"tag"`
	UUID    string `json:"uuid"`
	Content string `json:"content"`
}

type TradeOrderNotice struct {
	Uid  int64  `json:"uid"`
	Time string `json:"time"`
	// 9-19 修改，直接通知交易号
	TxCode string `json:"tx_code"`
}

type TradeOrderPayCallback struct {
	Uid    int64  `json:"uid"`
	TxCode string `json:"tx_code"`
	PayId  string `json:"pay_id"`
}

const (
	Unknown                        = 0
	TradeOrderEventTypeCreate      = 10014
	TradeOrderEventTypeExpire      = 10015
	TradeOrderEventTypePayCallback = 10018
)

var MsgFlags = map[int]string{
	Unknown:                        "未知",
	TradeOrderEventTypeCreate:      "交易订单创建",
	TradeOrderEventTypeExpire:      "交易订单过期",
	TradeOrderEventTypePayCallback: "交易订单支付结果",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[Unknown]
}
