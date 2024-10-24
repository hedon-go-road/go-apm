package metric

import "github.com/prometheus/client_golang/prometheus"

var (
	DropAlarmCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "drop_dogalarm_total",
		Help: "The total number of dropped alarms",
	}, []string{"alarm_app", "alarm_host", "notice_type"})

	LiveProbeGuage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "live_probe",
		Help: "The total number of live probes",
	}, []string{"alarm_app", "alarm_host"})
)

type LiveStatus int

const (
	Living   LiveStatus = 1
	Shutdown LiveStatus = 2
)

func All() []prometheus.Collector {
	return []prometheus.Collector{
		DropAlarmCounter, LiveProbeGuage,
	}
}
