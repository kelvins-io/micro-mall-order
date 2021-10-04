package startup

import (
	"gitee.com/cristiane/micro-mall-order/model/args"
	"gitee.com/cristiane/micro-mall-order/vars"
	"gitee.com/kelvins-io/kelvins"
	"gitee.com/kelvins-io/kelvins/setup"
	"gitee.com/kelvins-io/kelvins/util/queue_helper"
)

// SetupVars 加载变量
func SetupVars() error {
	var err error
	err = setupTradeOrderPayCall()
	if err != nil {
		return err
	}

	err = setupTradeOrderInfoSearchNotice()
	if err != nil {
		return err
	}

	return err
}

func setupTradeOrderPayCall() error {
	var err error
	vars.TradeOrderPayCallbackQueueServer, err = setup.NewAMQPQueue(vars.TradeOrderPayCallbackSetting, nil)
	return err
}

func setupTradeOrderInfoSearchNotice() error {
	var err error
	if vars.TradeOrderInfoSearchNoticeSetting != nil {
		vars.TradeOrderInfoSearchNoticeServer, err = setup.NewAMQPQueue(vars.TradeOrderInfoSearchNoticeSetting, nil)
		if err != nil {
			return err
		}
		vars.TradeOrderInfoSearchNoticePusher, err = queue_helper.NewPublishService(
			vars.TradeOrderInfoSearchNoticeServer, &queue_helper.PushMsgTag{
				DeliveryTag:    args.TradeOrderInfoSearchNoticeTag,
				DeliveryErrTag: args.TradeOrderInfoSearchNoticeTagErr,
				RetryCount:     vars.TradeOrderInfoSearchNoticeSetting.TaskRetryCount,
				RetryTimeout:   vars.TradeOrderInfoSearchNoticeSetting.TaskRetryTimeout,
			}, kelvins.BusinessLogger)
		if err != nil {
			return err
		}
	}
	return err
}
