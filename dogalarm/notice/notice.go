package notice

import (
	"context"
	"errors"

	"github.com/hedon-go-road/go-apm/dogalarm/thirdparty/feishu"
	"github.com/hedon-go-road/go-apm/dogapm"
)

type alarm struct{}

var Alarmer = &alarm{}

type NoticeType int

const (
	DingTalk NoticeType = 1
	Phone    NoticeType = 2
	Feishu   NoticeType = 3
)

func (a *alarm) Send(noticeType NoticeType, msg, webhook string) {
	switch noticeType {
	case Phone:
		a.sendPhone(msg, webhook)
	case DingTalk:
		a.sendDingTalk(msg, webhook)
	case Feishu:
		a.sendFeishu(msg, webhook)
	default:
		dogapm.Logger.Error(context.Background(),
			"send alarm notice failed",
			map[string]any{"noticeType": noticeType},
			errors.New("unknown notice type"))
	}
}

func (a *alarm) sendPhone(msg, webhook string) {
	// TODO: send phone message
	dogapm.Logger.Info(context.Background(), "notice", map[string]any{
		"noticeType": Phone,
		"msg":        msg,
		"webhook":    webhook,
	})
}

func (a *alarm) sendDingTalk(msg, webhook string) {
	// TODO: send dingtalk message
	dogapm.Logger.Info(context.Background(), "notice", map[string]any{
		"noticeType": DingTalk,
		"msg":        msg,
		"webhook":    webhook,
	})
}

func (a *alarm) sendFeishu(msg, webhook string) {
	if err := feishu.SendTextMsg(webhook, msg); err != nil {
		dogapm.Logger.Error(context.Background(), "send feishu text msg failed", map[string]any{"msg": msg, "webhook": webhook}, err)
	}
}
