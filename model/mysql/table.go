package mysql

import "time"

const (
	TableOrder          = "order"
	TableOrderSku       = "order_sku"
	TableOrderSceneShop = "order_scene_shop"
	TableConfigKv       = "config_kv_store"
)

type Order struct {
	Id                  int64     `xorm:"pk autoincr comment('自增ID') BIGINT"`
	TxCode              string    `xorm:"not null comment('交易号') unique(tx_code_order_code_index) CHAR(40)"`
	OrderCode           string    `xorm:"not null comment('订单code') unique unique(tx_code_order_code_index) CHAR(40)"`
	Uid                 int64     `xorm:"not null comment('用户UID') index BIGINT"`
	OrderTime           time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('下单时间') index DATETIME"`
	Description         string    `xorm:"comment('订单描述') index VARCHAR(255)"`
	ClientIp            string    `xorm:"comment('客户端IP') CHAR(16)"`
	DeviceCode          string    `xorm:"comment('客户端设备code') VARCHAR(512)"`
	ShopId              int64     `xorm:"not null comment('门店ID') index BIGINT"`
	State               int       `xorm:"not null default 0 comment('订单状态，0-有效，1-锁定中，2-无效') TINYINT"`
	PayExpire           time.Time `xorm:"not null comment('支付有效期，默认30分钟内有效') DATETIME"`
	PayState            int       `xorm:"not null default 0 comment('支付状态，0-未支付，1-支付中，2-支付失败，3-已支付，4-支付过期取消') TINYINT"`
	Amount              int       `xorm:"comment('订单关联商品数量') INT"`
	Money               string    `xorm:"not null default 0.0000000000000000 comment('订单总金额') DECIMAL(32,16)"`
	CoinType            int       `xorm:"default 0 comment(' 订单币种，0-CNY，1-USD') TINYINT"`
	LogisticsDeliveryId int       `xorm:"comment('物流投递ID') INT"`
	CreateTime          time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('创建时间') DATETIME"`
	UpdateTime          time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('修改时间') DATETIME"`
	InventoryVerify     int       `xorm:"default 0 comment('库存核实，0-未核实，1-核实') TINYINT"`
}

type OrderEstimate struct {
	Id           int64     `xorm:"pk autoincr comment('自增ID') BIGINT"`
	EstimateCode string    `xorm:"not null comment('评论code') unique(estimate_code_shop_id) CHAR(40)"`
	SkuCode      string    `xorm:"comment('商品sku') CHAR(40)"`
	OrderCode    string    `xorm:"comment('订单code') index CHAR(40)"`
	Uid          int64     `xorm:"comment('用户uid') index BIGINT"`
	ShopId       int64     `xorm:"comment('店铺ID') unique(estimate_code_shop_id) index BIGINT"`
	Content      string    `xorm:"comment('内容') TEXT"`
	Star         int       `xorm:"comment('星级') INT"`
	State        int       `xorm:"default 0 comment('状态，0-有效') TINYINT"`
	CreateTime   time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('创建时间') DATETIME"`
	UpdateTime   time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('修改时间') DATETIME"`
}

type OrderSceneShop struct {
	Id           int64     `xorm:"pk autoincr comment('自增ID') BIGINT"`
	OrderCode    string    `xorm:"comment('订单code') CHAR(40)"`
	ShopId       int64     `xorm:"comment('店铺ID') BIGINT"`
	ShopName     string    `xorm:"comment('店铺名') VARCHAR(512)"`
	ShopAreaCode string    `xorm:"comment('店铺区域code') VARCHAR(255)"`
	ShopAddress  string    `xorm:"comment('店铺地址') TEXT"`
	CreateTime   time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('创建时间') DATETIME"`
	UpdateTime   time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('更新时间') DATETIME"`
}

type OrderSku struct {
	Id         int64     `xorm:"pk autoincr comment('自增ID') BIGINT"`
	OrderCode  string    `xorm:"not null comment('对应订单code') unique(order_unique) CHAR(64)"`
	ShopId     int64     `xorm:"not null comment('店铺ID') unique(order_unique) index BIGINT"`
	SkuCode    string    `xorm:"not null comment('商品sku') unique(order_unique) index CHAR(64)"`
	Price      string    `xorm:"not null default 0.0000000000000000 comment('商品单价') DECIMAL(32,16)"`
	Amount     int       `xorm:"not null comment('商品数量') INT"`
	Name       string    `xorm:"comment('商品名称') index VARCHAR(255)"`
	CreateTime time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('创建时间') DATETIME"`
	UpdateTime time.Time `xorm:"not null default CURRENT_TIMESTAMP comment('修改时间') DATETIME"`
}
