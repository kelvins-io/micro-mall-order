package vars

import (
	"gitee.com/kelvins-io/common/queue"
	"gitee.com/kelvins-io/kelvins/config/setting"
)

var (
	EmailConfigSetting               *EmailConfigSettingS
	TradeOrderPayCallbackSetting     *setting.QueueAMQPSettingS
	TradeOrderPayCallbackQueueServer *queue.MachineryQueue
)
