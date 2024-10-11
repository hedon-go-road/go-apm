package internal

import (
	"os"
)

var (
	hostname string
	appName  string
)

func init() {
	hostname, _ = os.Hostname()
	appName = os.Getenv("APP_NAME")
}

type buildInfo struct{}

var BuildInfo = &buildInfo{}

func (b *buildInfo) Hostname() string {
	return hostname
}

func (b *buildInfo) AppName() string {
	return appName
}
