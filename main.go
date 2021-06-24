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

func main() {
	// Check Docker environment
	if err := CheckDockerEnv(); err != nil {
		fmt.Printf("Docker Environment is not avaiable: %s", err)
		os.Exit(-1)
	}

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
	
	NewTelegrafanaApiServer(ADDR, port).Start()
}
