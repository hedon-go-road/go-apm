package liveprobe

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hedon-go-road/go-apm/dogalarm/dao"
	"github.com/hedon-go-road/go-apm/dogalarm/notice"
	"github.com/hedon-go-road/go-apm/dogapm"
	"github.com/spf13/cast"
)

type probe struct{}

var Prober = &probe{}

func (p *probe) Enable() {
	dogapm.Logger.Info(context.Background(), "live probe enabled", nil)
	go func() {
		for {
			apps := dao.DeployInfo.All()
			for _, app := range apps {
				port, ips, appName, liveProbe := getAppInfo(app)
				if liveProbe == "" {
					continue
				}
				for _, ip := range ips {
					checkUrl := fmt.Sprintf("http://%s:%d/%s", ip, port, liveProbe)
					p.checkLive(checkUrl, appName, ip,
						cast.ToString(app["phone_webhook"]),
						cast.ToString(app["phone"]),
						3)
				}
			}
			time.Sleep(time.Second * 10)
		}
	}()
}

func getAppInfo(appCopy map[string]any) (port int, ips []string, appName, liveProbe string) {
	port = cast.ToInt(appCopy["port"])
	ips = strings.Split(cast.ToString(appCopy["hosts"]), ",")
	appName = cast.ToString(appCopy["app"])
	liveProbe = cast.ToString(appCopy["liveprobe"])
	return
}

func (p *probe) checkLive(checkUrl, appName, host, alarmURL, phone string, retryCnt int) {
	failCnt := 0
	for i := 0; i < retryCnt; i++ {
		req, err := http.NewRequest(http.MethodGet, checkUrl, http.NoBody)
		if err != nil {
			failCnt++
			continue
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			failCnt++
			if resp != nil {
				resp.Body.Close()
			}
			dogapm.Logger.Warn(context.Background(), "check live probe failed", map[string]any{"checkUrl": checkUrl, "err": err})
			continue
		}
		failCnt = 0
		break
	}
	if failCnt > 0 {
		notice.Alarmer.Send(notice.Phone, alarmURL, fmt.Sprintf("service %s on host %s is down", appName, host), phone)
	}
}
