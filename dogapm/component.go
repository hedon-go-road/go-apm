package dogapm

import (
	"os"
	"os/signal"
	"syscall"
)

type starter interface {
	Start()
}

type closer interface {
	Close()
}

var (
	globalStarters = make([]starter, 0)
	globalClosers  = make([]closer, 0)
)

type endPoint struct {
	stop chan struct{}
}

var EndPoint = &endPoint{
	stop: make(chan struct{}, 1),
}

func (e *endPoint) Start() {
	for _, s := range globalStarters {
		go s.Start()
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
		<-quit
		e.Shutdown()
	}()

	<-e.stop
}

func (e *endPoint) Shutdown() {
	for _, c := range globalClosers {
		c.Close()
	}
	e.stop <- struct{}{}
}
