package mysql

import "time"

const (
	TableOrder    = "order"
	TableOrderSku = "order_sku"
)

type Order struct {
	Id           int64     `xorm:"pk autoincr comment('自增ID') BIGINT"`
	OrderCode    string    `xorm:"not null comment('订单code') unique CHAR(40)"`
	Uid          int64     `xorm:"not null comment('用户UID') index BIGINT"`
	OrderTime    time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('下单时间') index DATETIME"`
	Description  string    `xorm:"comment('订单描述') index VARCHAR(255)"`
	ClientIp     string    `xorm:"comment('客户端IP') CHAR(16)"`
	DeviceCode   string    `xorm:"comment('客户端设备code') VARCHAR(512)"`
	ShopId       int64     `xorm:"not null comment('门店ID') index BIGINT"`
	ShopName     string    `xorm:"not null comment('门店名称') index VARCHAR(255)"`
	ShopAreaCode string    `xorm:"comment('门店区域编号') VARCHAR(255)"`
	ShopAddress  string    `xorm:"comment('门店地址') TEXT"`
	State        int       `xorm:"not null default 0 comment('订单状态，0-有效，1-锁定中，2-无效') TINYINT"`
	PayExpire    time.Time `xorm:"not null comment('支付有效期，默认30分钟内有效') DATETIME"`
	PayState     int       `xorm:"not null default 0 comment('支付状态，0-未支付，1-支付中，2-支付失败，3-已支付') TINYINT"`
	TotalAmount  string    `xorm:"not null default 0.0000000000000000 comment('订单总金额') DECIMAL(32,16)"`
	CreateTime   time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('创建时间') DATETIME"`
	UpdateTime   time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('修改时间') DATETIME"`
}

type OrderSku struct {
	Id         int64     `xorm:"pk autoincr comment('自增ID') BIGINT"`
	OrderCode  string    `xorm:"not null comment('对应订单code') unique(order_code_sku_index) CHAR(64)"`
	SkuCode    string    `xorm:"not null comment('商品sku') unique(order_code_sku_index) CHAR(64)"`
	Price      string    `xorm:"not null default 0.0000000000000000 comment('商品单价') DECIMAL(32,16)"`
	Amount     int       `xorm:"not null comment('商品数量') INT"`
	Name       string    `xorm:"comment('商品名称') index VARCHAR(255)"`
	CreateTime time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('创建时间') DATETIME"`
	UpdateTime time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('修改时间') DATETIME"`
}
