package main

import (
	// "context"
	"fmt"
	// "net/http"
	// "os"
	// "os/signal"
	// "syscall"
)

type Telegrafana struct {
	Server *ApiServer
	InstanceManager *TelegrafInstanceManager
}

func DefaultTelegrafana() *Telegrafana {
	return NewTelegrafana("0.0.0.0", PORT)
}

func NewTelegrafana(addr string, port int) *Telegrafana {
	return &Telegrafana {
		Server: NewApiServer(addr, port),
		InstanceManager: NewTelegrafIntanceManager(),
	}
}

func (t *Telegrafana) startInstanceManager() error {
	// Check Docker environment
	if err := t.InstanceManager.Start(); err != nil {
		// fmt.Printf("Docker Environment is not avaiable: %s", err)
		return err
	}

	return nil
}

func (t *Telegrafana) startApiServer() error {
	return t.Server.Start()
}

func (t *Telegrafana) Start() error {
	if err:= t.startInstanceManager(); err != nil {
		fmt.Printf("Docker Environment is not avaiable: %s", err)
		return err
	}

	if err:= t.startApiServer(); err != nil {
		fmt.Printf("Failed to Api Server: %s", err)
		return err
	}

	return nil
}