package main

import (
	// "context"
	"fmt"
	// "net/http"
	"os"
	"strconv"
	// "os/signal"
	// "syscall"

	// "github.com/gin-gonic/gin"
)

type Service interface {
	Start() error
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
	
	NewTelegrafana(ADDR, port).Start()
}
