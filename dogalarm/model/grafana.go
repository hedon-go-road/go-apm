package model

const (
	StatusFiring   = "firing"
	StatusResolved = "resolved"
	StatusAlert    = "alert"
)

type GrafanaAlert struct {
	Receiver          string      `json:"receiver"`
	Status            string      `json:"status"`
	Alerts            []*Alert    `json:"alerts"`
	GroupLabels       GroupLabels `json:"groupLabels"`
	CommonLabels      Labels      `json:"commonLabels"`
	CommonAnnotations Annotations `json:"commonAnnotations"`
	ExternalURL       string      `json:"externalURL"`
	Version           string      `json:"version"`
	GroupKey          string      `json:"groupKey"`
	TruncatedAlerts   int64       `json:"truncatedAlerts"`
	OrgID             int64       `json:"orgId"`
	Title             string      `json:"title"`
	State             string      `json:"state"`
	Message           string      `json:"message"`
}

type Alert struct {
	Status       string      `json:"status"`
	Labels       Labels      `json:"labels"`
	Annotations  Annotations `json:"annotations"`
	StartsAt     string      `json:"startsAt"`
	EndsAt       string      `json:"endsAt"`
	GeneratorURL string      `json:"generatorURL"`
	Fingerprint  string      `json:"fingerprint"`
	SilenceURL   string      `json:"silenceURL"`
	DashboardURL string      `json:"dashboardURL"`
	PanelURL     string      `json:"panelURL"`
	Values       Values      `json:"values"`
	ValueString  string      `json:"valueString"`
}

type Annotations struct {
}

type Labels struct {
	Alertname     string `json:"alertname"`
	App           string `json:"app"`
	Dogapm        string `json:"dogapm"`
	GrafanaFolder string `json:"grafana_folder"`
}

type Values struct {
	B float64 `json:"B"`
	C int64   `json:"C"`
}

type GroupLabels struct {
	Alertname     string `json:"alertname"`
	GrafanaFolder string `json:"grafana_folder"`
}
