package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hedon-go-road/go-apm/dogalarm/dao"
	"github.com/hedon-go-road/go-apm/dogalarm/model"
	"github.com/hedon-go-road/go-apm/dogalarm/notice"
	"github.com/hedon-go-road/go-apm/dogapm"
)

type alarm struct{}

var Alarm = &alarm{}

func (a *alarm) MetricWebHook(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	alert := model.GrafanaAlert{}
	err = json.Unmarshal(data, &alert)
	if err != nil {
		dogapm.Logger.Error(context.Background(), "unmarshal metric data failed", map[string]any{"data": string(data)}, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, alert := range alert.Alerts {
		app, msg := alert.Labels.App, alert.Labels.Alertname
		msg = fmt.Sprintf("content=%s, app=%s, value=%v", msg, app, alert.Values)
		deployInfo := dao.DeployInfo.GetInfoByApp(app)
		phoneWebHook := deployInfo["phone_webhook"].(string)
		if phoneWebHook != "" {
			notice.Alarmer.Send(notice.Phone, msg, phoneWebHook)
		}
	}
}

func (a *alarm) LogWebHook(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log := model.LogstashLog{}
	err = json.Unmarshal(data, &log)
	if err != nil {
		dogapm.Logger.Error(context.Background(), "unmarshal log data failed", map[string]any{"data": string(data)}, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app, host, msg := log.AppName, log.AppHost, log.Msg
	deployInfo := dao.DeployInfo.GetInfoByApp(app)
	feishuWebHook := deployInfo["feishu_webhook"].(string)
	if feishuWebHook == "" {
		return
	}

	msg = fmt.Sprintf("content=%s, app=%s, host=%s", msg, app, host)
	if filterDuplicate(msg) {
		return
	}
	notice.Alarmer.Send(notice.Feishu, msg, feishuWebHook)
}

const (
	msgCutLen   = 1024
	minuteLimit = 10
)

func filterDuplicate(msg string) (duplicate bool) {
	// we just read the first 1kb of msg to check if it's duplicate,
	// and we would check if it's duplicate by its hash plus original msg length.
	msgLen := len(msg)
	if msgLen > msgCutLen {
		msg = msg[:msgCutLen]
	}
	hashStr := getMd5Hash(msg)
	str := fmt.Sprintf("%s_%d", hashStr, msgLen)
	return dogapm.RedisLimiter.IsLimit(dogapm.Infra.RDB, str, minuteLimit, time.Minute)
}

func getMd5Hash(msg string) string {
	hash := sha256.Sum256([]byte(msg))
	return hex.EncodeToString(hash[:])
}
