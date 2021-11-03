package startup

import (
	"gitee.com/cristiane/micro-mall-order/vars"
	"gitee.com/kelvins-io/kelvins/config"
	"gitee.com/kelvins-io/kelvins/config/setting"
)

const (
	SectionEmailConfig                = "email-config"
	SectionAMQPOrderTradePayCallback  = "trade-order-pay-callback"
	SectionTradeOrderInfoSearchNotice = "trade-order-info-search-notice"
	SectionG2Cache                    = "micro-mall-g2cache"
)

// LoadConfig 加载配置对象映射
func LoadConfig() error {
	// 加载email数据源
	vars.EmailConfigSetting = new(vars.EmailConfigSettingS)
	config.MapConfig(SectionEmailConfig, vars.EmailConfigSetting)
	// 加载mq配置
	vars.TradeOrderPayCallbackSetting = new(setting.QueueAMQPSettingS)
	config.MapConfig(SectionAMQPOrderTradePayCallback, vars.TradeOrderPayCallbackSetting)

	// 订单搜索通知
	vars.TradeOrderInfoSearchNoticeSetting = new(setting.QueueAMQPSettingS)
	config.MapConfig(SectionTradeOrderInfoSearchNotice, vars.TradeOrderInfoSearchNoticeSetting)

	//加载G2Cache二级缓存配置
	vars.G2CacheSetting = new(vars.G2CacheSettingS)
	config.MapConfig(SectionG2Cache, vars.G2CacheSetting)

	return nil
}
