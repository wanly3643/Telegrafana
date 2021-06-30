package main

import (
	// "context"
	"fmt"
	// "net/http"
	"os"
	"strconv"
	"os/signal"
	"syscall"

	// "github.com/gin-gonic/gin"
)

type Service interface {
	Start() error
	Stop() error
}

func AddHandlerBeforeExit(server *Telegrafana) {
    c := make(chan os.Signal)
    signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
    go func() {
        for s := range c {
            switch s {
            case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
                fmt.Println("Exiting", s)
				// Some clear work
				server.Stop()

				os.Exit(0)
            default:
                fmt.Println("other", s)
            }
        }
    }()
}

func main() {
	// Start Api Server
	strPort := os.Getenv("PORT")
	port := PORT
	if strPort != "" {
		portVal, err := strconv.Atoi(strPort)
		if err != nil {
			fmt.Printf("Invalid PORT: %s", strPort)
		} else {
			port = portVal
		}
	}
	
	server := NewTelegrafana(ADDR, port)

	AddHandlerBeforeExit(server)
	
	server.Start()
}
