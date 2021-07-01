// API Server

package main

import (
	// "context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	// "os"
	// "os/signal"
	// "syscall"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const PORT = 8080
const ADDR = "0.0.0.0"

type ApiServer struct {
	Host   string
	Addr   string
	Port   int
	Engine *gin.Engine
	Srv    *Telegrafana
}

// Create standard response
func createApiResp(err error, data interface{}) interface{} {
	resp := map[string]interface{}{}

	if err != nil {
		resp["status"] = 1
		resp["message"] = err.Error()
	} else {
		resp["status"] = 0
		resp["data"] = data
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
	r.StaticFile("/", "html/index.html")
	r.Static("/assets", "./html/assets")
	//r.GET("/", renderConsolePage)

	return &ApiServer{
		Host:   "",
		Addr:   addr,
		Port:   port,
		Engine: r,
		Srv:    srv,
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
	t.Engine.POST("/instance/start/:name", t.startTelegrafInstance)
	t.Engine.POST("/instance/stop/:name", t.stopTelegrafInstance)
	t.Engine.POST("/instance/restart/:name", t.restartTelegrafInstance)
	t.Engine.POST("/instance/delete/:name", t.removeTelegrafInstance)
	t.Engine.GET("/instance/config/:name", t.getTelegrafConfig)
	t.Engine.GET("/instance/detail/:name", t.getTelegrafInstance)
	t.Engine.GET("/instance/list", t.getInstances)

	return t.Engine.Run(addr)
}

func (t *ApiServer) getInstances(c *gin.Context) {
	// Get all the instance information list
	instances, err := t.Srv.StorageManager.GetInstanceList()
	total := len(instances)
	if err != nil || total < 1 {
		data := map[string]interface{}{
			"total": total,
			"data":  make([]interface{}, 0),
		}
		c.JSON(http.StatusOK, createApiResp(nil, data))
		return
	}

	// Group by ID
	instanceStatsByID := map[string]*TelegrafInstanceStat{}
	instanceStats, err := t.Srv.InstanceManager.GetTelegrafInstances()
	if err == nil && len(instanceStats) > 0 {
		for _, s := range instanceStats {
			instanceStatsByID[s.ID] = &s
		}
	}

	// Append docker container status one by one
	data := make([]TelegrafInstanceResp, 0, total)
	for _, i := range instances {
		s, ok := instanceStatsByID[i.DockerContainerID]
		if ok {
			data = append(data, t.createInstanceResp(&i, s))
		} else {
			data = append(data, t.createInstanceResp(&i, nil))
		}
	}

	// Send response
	ret := map[string]interface{}{
		"total": total,
		"data":  data,
	}
	c.JSON(http.StatusOK, createApiResp(nil, ret))
}

type HandlerFunc func(string) error

func (t *ApiServer) handleTelegrafInstance(c *gin.Context, handler HandlerFunc) {
	instanceID := c.Param("name")
	instanceInfo, err := t.Srv.StorageManager.GetInstance(instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, createApiResp(err, nil))
		return
	}

	if instanceInfo == nil {
		c.JSON(http.StatusNotFound, createApiResp(err, nil))
		return
	}

	err = handler(instanceInfo.DockerContainerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, createApiResp(err, nil))
		return
	}

	c.JSON(http.StatusOK, createApiResp(err, nil))
}

func (t *ApiServer) startTelegrafInstance(c *gin.Context) {
	t.handleTelegrafInstance(c, t.Srv.InstanceManager.StartTelegrafInstance)
}

func (t *ApiServer) stopTelegrafInstance(c *gin.Context) {
	t.handleTelegrafInstance(c, t.Srv.InstanceManager.StopTelegrafInstance)
}

func (t *ApiServer) restartTelegrafInstance(c *gin.Context) {
	t.handleTelegrafInstance(c, t.Srv.InstanceManager.RestartTelegrafInstance)
}

func (t *ApiServer) createTelegrafConfig(c *gin.Context) (string, error) {
	b, err := ioutil.ReadFile("./test001.conf")
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (t *ApiServer) removeTelegrafInstance(c *gin.Context) {
	instanceID := c.Param("name")
	fmt.Println("ID:", instanceID)

	instanceInfo, err := t.Srv.StorageManager.GetInstance(instanceID)
	if err != nil || instanceInfo == nil {
		c.JSON(http.StatusNotFound, createApiResp(nil, nil))
		return
	}

	// Remove docker instance
	err = t.Srv.InstanceManager.RemoveTelegrafInstance(instanceInfo.DockerContainerID)
	if err != nil {
		c.JSON(http.StatusNotFound, createApiResp(err, nil))
		return
	}

	// Remove information from database
	err = t.Srv.StorageManager.RemoveInstance(instanceID)

	// Return Response
	c.JSON(http.StatusNotFound, createApiResp(nil, nil))
}

func (t *ApiServer) createTelegrafInstance(c *gin.Context) {
	instanceID := uuid.New().String()
	configUrl := fmt.Sprintf("http://%s:%d/instance/config/%s", t.Host, t.Port, instanceID)
	fmt.Println("Config URL:", configUrl)

	// Create Telegraf Config and Get the ID
	configStr, err := t.createTelegrafConfig(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, createApiResp(err, nil))
		return
	}

	// Create docker container but not start
	containerID, err := t.Srv.InstanceManager.CreateTelegrafInstance(configUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, createApiResp(err, nil))
		return
	}

	// Add information into database
	instanceObj := &TelegrafInstanceInfo{
		ID:                instanceID,
		Name:              "",
		DockerContainerID: containerID,
		Config:            configStr,
		Description:       "",
		Created:           GetCurrentTimeString(),
	}
	err = t.Srv.StorageManager.PutInstance(instanceObj)
	if err != nil {
		// Remove the create docker container
		t.Srv.InstanceManager.RemoveTelegrafInstance(containerID)
		c.JSON(http.StatusInternalServerError, createApiResp(err, nil))
		return
	}

	// Start the container
	err = t.Srv.InstanceManager.StartTelegrafInstance(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, createApiResp(err, nil))
	} else {
		data := map[string]interface{}{
			"id": instanceID,
		}
		c.JSON(http.StatusOK, createApiResp(nil, data))
	}
}

func (t *ApiServer) getTelegrafConfig(c *gin.Context) {
	instanceID := c.Param("name")
	fmt.Println("ID:", instanceID)
	instanceInfo, err := t.Srv.StorageManager.GetInstance(instanceID)
	// err := runAgent(instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, createApiResp(err, nil))
		return
	}

	if instanceInfo == nil {
		c.JSON(http.StatusNotFound, createApiResp(nil, nil))
		return
	}
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	// } else {
	// c.JSON(http.StatusOK, gin.H{"status": "OK"})
	// }

	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.toml\"", instanceID))
	c.Data(http.StatusOK, "application/octet-stream", []byte(instanceInfo.Config))
}

type TelegrafInstanceResp struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Created string `json:"created"`
	Config  string `json:"config"`
}

func (t *ApiServer) createInstanceResp(instanceInfo *TelegrafInstanceInfo, containerStat *TelegrafInstanceStat) TelegrafInstanceResp {
	status := "Removed"
	if containerStat != nil {
		status = containerStat.Status
	}
	return TelegrafInstanceResp{
		ID:      instanceInfo.ID,
		Name:    instanceInfo.Name,
		Config:  instanceInfo.Config,
		Created: instanceInfo.Created,
		Status:  status,
	}
}

func (t *ApiServer) getTelegrafInstance(c *gin.Context) {
	instanceID := c.Param("name")
	fmt.Println("ID:", instanceID)
	instanceInfo, err := t.Srv.StorageManager.GetInstance(instanceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, createApiResp(err, nil))
		return
	}

	if instanceInfo == nil {
		c.JSON(http.StatusNotFound, createApiResp(nil, nil))
		return
	}

	containerInfo, err := t.Srv.InstanceManager.GetTelegrafInstanceStat(instanceInfo.DockerContainerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, createApiResp(err, nil))
		return
	}

	// Send response
	var data interface{}
	data = t.createInstanceResp(instanceInfo, containerInfo)
	c.JSON(http.StatusOK, createApiResp(nil, data))
}

func renderConsolePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}
