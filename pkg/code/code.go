package code

import "gitee.com/kelvins-io/common/errcode"

const (
	Success                  = 29000000
	ErrorServer              = 29000001
	UserNotExist             = 29000005
	UserExist                = 29000006
	DBDuplicateEntry         = 29000007
	MerchantExist            = 29000008
	MerchantNotExist         = 29000009
	ShopBusinessExist        = 29000010
	ShopBusinessNotExist     = 29000011
	SkuCodeEmpty             = 29000012
	SkuCodeNotExist          = 29000013
	SkuCodeExist             = 29000014
	SkuAmountNotEnough       = 29000015
	TransactionFailed        = 29000016
	OrderNotExist            = 29000017
	OrderExist               = 29000018
	OrderStateLocked         = 29000019
	OrderStateProhibit       = 29000020
	OperationNotEffect       = 29000021
	OrderTxCodeEmpty         = 29000022
	OrderDeliveryNotExist    = 29000023
	OrderTxCodeNotExist      = 29000024
	SkuPriceVersionNotExist  = 29000025
	OrderPayCompleted        = 29000026
	OrderExpire              = 29000027
	OrderStateInvalid        = 29000028
	RequestDataTooMuch       = 29000029
	ErrRequestDataFormat     = 29000030
	UserDeliveryInfoNotExist = 29000031
	UserStateNotVerify       = 29000032
	OrderPayIng              = 29000033
)

var ErrMap = make(map[int]string)

func init() {
	dict := map[int]string{
		Success:                  "OK",
		ErrorServer:              "服务器错误",
		UserNotExist:             "用户不存在",
		DBDuplicateEntry:         "Duplicate entry",
		UserExist:                "已存在用户记录，请勿重复创建",
		MerchantExist:            "商户认证材料已存在",
		MerchantNotExist:         "商户未提交材料",
		ShopBusinessExist:        "店铺申请材料已存在",
		ShopBusinessNotExist:     "商户未提交店铺材料",
		SkuCodeEmpty:             "商品唯一code为空",
		SkuCodeNotExist:          "商品唯一code在系统找不到",
		SkuCodeExist:             "商品唯一code已存在系统",
		SkuAmountNotEnough:       "商品库存不够",
		TransactionFailed:        "事务执行失败",
		OrderNotExist:            "订单不存在",
		OrderExist:               "订单已存在",
		OrderStateLocked:         "订单被锁定",
		OrderStateProhibit:       "订单状态禁止更改",
		OperationNotEffect:       "操作未生效",
		OrderTxCodeEmpty:         "订单事务号为空",
		OrderDeliveryNotExist:    "订单交付信息不存在",
		OrderTxCodeNotExist:      "订单交易号不存在",
		SkuPriceVersionNotExist:  "商品价格版本不存在或不符合规则",
		OrderPayCompleted:        "订单支付完成",
		OrderPayIng:              "订单正在支付中",
		OrderExpire:              "订单过期",
		OrderStateInvalid:        "订单无效或被锁定",
		RequestDataTooMuch:       "请求数据过多",
		ErrRequestDataFormat:     "请求数据格式不正确",
		UserDeliveryInfoNotExist: "用户物流信息不存在",
		UserStateNotVerify:       "用户身份未验证或审核或被锁定",
	}
	errcode.RegisterErrMsgDict(dict)
	for key, _ := range dict {
		ErrMap[key] = dict[key]
	}
}
