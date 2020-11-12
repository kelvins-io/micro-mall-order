package vars

import (
	"gitee.com/kelvins-io/common/queue"
	"gitee.com/kelvins-io/kelvins/config/setting"
)

var (
	EmailConfigSetting           *EmailConfigSettingS
	AppName                      = ""
	TradeOrderQueueServer        *queue.MachineryQueue
	TradeOrderPayCallbackSetting *setting.QueueAMQPSettingS
	TradeOrderPayQueueServer     *queue.MachineryQueue
)
