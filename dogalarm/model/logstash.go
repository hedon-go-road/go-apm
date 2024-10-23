package model

type LogstashLog struct {
	Input     Input    `json:"input"`
	Ecs       Ecs      `json:"ecs"`
	Version   string   `json:"@version"`
	Time      string   `json:"time"`
	Agent     Agent    `json:"agent"`
	Log       Log      `json:"log"`
	Host      Host     `json:"host"`
	Level     string   `json:"level"`
	AppName   string   `json:"app_name"`
	AppHost   string   `json:"app_host"`
	Timestamp string   `json:"@timestamp"`
	Tags      []string `json:"tags"`
	Msg       string   `json:"msg"`
	Message   string   `json:"message"`
}

type Agent struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Hostname    string `json:"hostname"`
	Type        string `json:"type"`
	EphemeralID string `json:"ephemeral_id"`
	ID          string `json:"id"`
}

type Ecs struct {
	Version string `json:"version"`
}

type Host struct {
	Name string `json:"name"`
}

type Input struct {
	Type string `json:"type"`
}

type Log struct {
	Offset int64 `json:"offset"`
	File   File  `json:"file"`
}

type File struct {
	Path string `json:"path"`
}
