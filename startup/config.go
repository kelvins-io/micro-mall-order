package startup

import (
	"gitee.com/cristiane/micro-mall-order/vars"
	"gitee.com/kelvins-io/kelvins/config"
	"gitee.com/kelvins-io/kelvins/config/setting"
	"log"
)

const (
	SectionEmailConfig               = "email-config"
	SectionAMQPOrderTradePayCallback = "trade-order-pay-callback"
)

// LoadConfig 加载配置对象映射
func LoadConfig() error {
	// 加载email数据源
	log.Printf("[info] Load custom config %s", SectionEmailConfig)
	vars.EmailConfigSetting = new(vars.EmailConfigSettingS)
	config.MapConfig(SectionEmailConfig, vars.EmailConfigSetting)
	// 加载mq配置
	log.Printf("[info] Load custom config %s", SectionAMQPOrderTradePayCallback)
	vars.TradeOrderPayCallbackSetting = new(setting.QueueAMQPSettingS)
	config.MapConfig(SectionAMQPOrderTradePayCallback, vars.TradeOrderPayCallbackSetting)

	return nil
}
