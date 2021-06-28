// API Server

package main

import (
	// "context"
	"errors"
	"fmt"
	"net/http"
	"net"
	// "os"
	// "os/signal"
	// "syscall"

	"github.com/gin-gonic/gin"
)

const PORT = 8080
const ADDR = "0.0.0.0"

type ApiServer struct {
	Host string
	Addr string
	Port int
	Engine *gin.Engine
	Srv *Telegrafana
}

// Create standard response
func createApiResp(err error, data map[string]interface{}) map[string]interface{} {
	resp := map[string]interface{}{
	}

	if err != nil {
		resp["status"] = 1
		resp["message"] = ""
	} else {
		resp["status"] = 0
	}

	return resp
}


// Get the IP of Api Server
func externalIP() (string, error) {
    ifaces, err := net.Interfaces()
    if err != nil {
        return "", err
    }
    for _, iface := range ifaces {
        if iface.Flags&net.FlagUp == 0 {
            continue // interface down
        }
        if iface.Flags&net.FlagLoopback != 0 {
            continue // loopback interface
        }
        addrs, err := iface.Addrs()
        if err != nil {
            return "", err
        }
        for _, addr := range addrs {
            ip := getIpFromAddr(addr)
            if ip == "" {
                continue
            }
            return ip, nil
        }
    }
    return "", errors.New("connected to the network?")
}

func getIpFromAddr(addr net.Addr) string {
    var ip net.IP
    switch v := addr.(type) {
    case *net.IPNet:
        ip = v.IP
    case *net.IPAddr:
        ip = v.IP
    }
    if ip == nil || ip.IsLoopback() {
        return ""
    }
    ip = ip.To4()
    if ip == nil {
        return "" // not an ipv4 address
    }

    return ip.String()
}


func DefaultApiServer(srv *Telegrafana) *ApiServer {
	return NewApiServer("0.0.0.0", PORT, srv)
}

func NewApiServer(addr string, port int, srv *Telegrafana) *ApiServer {
	// Initializa routers
	r := gin.Default()

	// Web console
	r.LoadHTMLGlob("html/index.html")
	r.Static("/assets", "./html/assets")
	r.GET("/", renderConsolePage)

	return &ApiServer {
		Host: "",
		Addr: addr,
		Port: port,
		Engine: r,
		Srv: srv,
	}
}

func (t *ApiServer) Start() error {
	// Check the host
	if hostname, err := externalIP(); err != nil {
		return err
	} else {
		fmt.Printf("The hostname is: %s\n", hostname)
		t.Host = hostname
	}

	addr := fmt.Sprintf("%s:%d", t.Addr, t.Port)

	// Web Apis
	t.Engine.POST("/instance/create", t.createTelegrafInstance)
	t.Engine.GET("/instance/:name/start", t.startTelegrafInstance)
	t.Engine.GET("/instance/:name/config", t.getTelegrafConfig)
	t.Engine.GET("/instances", t.getInstances)
	t.Engine.GET("/config/:name", t.getTelegrafConfig)

	return t.Engine.Run(addr)
}

func (t *ApiServer)getInstances(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func (t *ApiServer)startTelegrafInstance(c *gin.Context) {
	// instanceID := c.Param("name")
	// err := runAgent(instanceID)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	// } else {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	//
}

func (t *ApiServer)createTelegrafInstance(c *gin.Context) {
	// Create Telegraf Config and Get the ID
	configID := "ac164"
	configUrl := fmt.Sprintf("http://%s:%d/config/%s", t.Host, t.Port, configID)
	fmt.Println("Config URL:", configUrl)
	instanceID, err := t.Srv.InstanceManager.RunTelegrafInstance(configUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "OK", "data": map[string]interface{}{"ID": instanceID}})
	}
}

func (t *ApiServer)getTelegrafConfig(c *gin.Context) {
	configID := c.Param("name")
	// err := runAgent(instanceID)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	// } else {
	// c.JSON(http.StatusOK, gin.H{"status": "OK"})
	// }

	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.toml\"", configID))
    c.Writer.Header().Add("Content-Type", "application/octet-stream")
	c.File("./test001.conf")
}

func renderConsolePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}