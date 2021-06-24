package main

import (
	// "context"
	"fmt"
	"net/http"
	// "os"
	// "os/signal"
	// "syscall"

	"github.com/gin-gonic/gin"
)

const PORT = 8080
const ADDR = "0.0.0.0"

type ApiServer interface {
	Start()
}

type TelegrafanaApiServer struct {
	Addr string
	Port int
	G *gin.Engine
}

func DefaultTelegrafanaApiServer() *TelegrafanaApiServer {
	return NewTelegrafanaApiServer ("0.0.0.0", PORT)
}

func NewTelegrafanaApiServer(addr string, port int) *TelegrafanaApiServer {
	// Initializa routers
	r := gin.Default()
	r.LoadHTMLGlob("html/index.html")
	r.Static("/assets", "./html/assets")
	r.GET("/", renderConsolePage)
	r.GET("/instance/:name/start", startTelegrafInstance)
	r.GET("/instances", getInstances)

	return &TelegrafanaApiServer {
		Addr: addr,
		Port: port,
		G: r,
	}
}

func (apiServer *TelegrafanaApiServer) Start() {
	addr := fmt.Sprintf("%s:%d", apiServer.Addr, apiServer.Port)
	apiServer.G.Run(addr)
}

func getInstances(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func startTelegrafInstance(c *gin.Context) {
	// instanceID := c.Param("name")
	// err := runAgent(instanceID)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	// } else {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	// }
}

func renderConsolePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}