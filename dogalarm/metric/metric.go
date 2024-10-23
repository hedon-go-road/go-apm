package metric

import "github.com/prometheus/client_golang/prometheus"

var (
	DropAlarmCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dogalarm_drop_alarm_total",
		Help: "The total number of dropped alarms",
	}, []string{"app"})
)

func All() []prometheus.Collector {
	return []prometheus.Collector{
		DropAlarmCounter,
	}
}
