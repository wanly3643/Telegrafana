// This is module to handle the docker 

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
)

const TELEGRAF_DOCKER_IMAGE_TAG = "1.19.0-alpine"
const TELEGRAF_DOCKER_IMAGE_NAME = "telegraf"
const TELEGRAF_DOCKER_IMAGE = TELEGRAF_DOCKER_IMAGE_NAME + ":" + TELEGRAF_DOCKER_IMAGE_TAG
const TELEGRAF_DOCKER_CONTAINER_PREFIX = "Telegrafana-"
const DOCKER_CONTAINER_RESTART_POLICY = "on-failure"
const DOCKER_CONTAINER_RESTART_LIMIT = 5

type TelegrafInstanceManager struct {
	client *client.Client
	ctx context.Context
}

type TelegrafInstanceStat struct {
	ID string
	Name string
	Status string
}

func NewTelegrafIntanceManager() *TelegrafInstanceManager {
	return &TelegrafInstanceManager{
		client: nil,
		ctx: nil,
	}
}

func GetInstanceUniqueName() string {
	return TELEGRAF_DOCKER_CONTAINER_PREFIX + time.Now().Format("20060102150405999999999")
}

// Check if the docker image of telegraf exists and pull the image if it does
// not exist
func (m *TelegrafInstanceManager) checkTelegrafDockerImage() error {
	images, err := m.client.ImageList(m.ctx, types.ImageListOptions{})
	if err != nil {
		return err
	}

	found := false
	for _, image := range images {
		for _, t := range image.RepoTags {
			if TELEGRAF_DOCKER_IMAGE == t {
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	// If the image does not exist, pull it
	if !found {
		fmt.Println("Telegraf Docker image does not exist, will pull it ...")
		out, err := m.client.ImagePull(m.ctx, TELEGRAF_DOCKER_IMAGE, types.ImagePullOptions{})
		if err != nil {
			return err
		}
	
		defer out.Close()
	
		io.Copy(os.Stdout, out)
	} else {
		fmt.Println("Telegraf Docker image is ready.")
	}

	return nil
}

func (m *TelegrafInstanceManager) Start() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	m.client = cli
	m.ctx = context.Background()

	if err := m.checkTelegrafDockerImage(); err != nil {
		return err
	}

	return nil
}

func (m *TelegrafInstanceManager) CreateTelegrafInstance(configUrl string) (string, error) {
	if m.client == nil || m.ctx == nil {
		return "", errors.New("Please Start first")
	}

	// Create container
	resp, err := m.client.ContainerCreate(m.ctx, &container.Config{
		Image: TELEGRAF_DOCKER_IMAGE,
		Cmd: strslice.StrSlice{"telegraf", "--config", configUrl},
	}, &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: DOCKER_CONTAINER_RESTART_POLICY,
			MaximumRetryCount: DOCKER_CONTAINER_RESTART_LIMIT,
		},
	}, nil, nil, GetInstanceUniqueName())
	if err != nil {
		return "", err
	}

	// err = m.StartTelegrafInstance(resp.ID)
	// if err != nil {
	// 	return resp.ID, err
	// }

	// fmt.Println("Docker instance started:", resp.ID)

	return resp.ID, nil
}

func (m *TelegrafInstanceManager) GetTelegrafInstances() ([]TelegrafInstanceStat, error) {
	ret := make([]TelegrafInstanceStat, 0, 100)
	if m.client == nil || m.ctx == nil {
		return ret, errors.New("Please Start first")
	}

	containers, err :=  m.client.ContainerList(m.ctx, types.ContainerListOptions{All: true,})
	if err != nil {
		return ret, err
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if strings.HasPrefix(name, "/") {
				name = name[1:]
			}

			if strings.HasPrefix(name, TELEGRAF_DOCKER_CONTAINER_PREFIX) {
				ret = append(ret, TelegrafInstanceStat{
					ID: container.ID,
					Name: name,
					Status: container.Status,
				})
			}
		}
	}

	return ret, nil
}

func (m *TelegrafInstanceManager) GetTelegrafInstanceStat(containerID string) (*TelegrafInstanceStat, error) {
	if m.client == nil || m.ctx == nil {
		return nil, errors.New("Please Start first")
	}

	containers, err :=  m.client.ContainerList(m.ctx, types.ContainerListOptions{All: true,})
	if err != nil {
		return nil, err
	}

	for _, container := range containers {
		if container.ID == containerID {
			name := ""
			if len(container.Names) > 0 {
				name = container.Names[0]
			}
			return &TelegrafInstanceStat{
				ID: container.ID,
				Name: name,
				Status: container.Status,
			}, nil
		}
	}

	return nil, nil
}

// Start the existing docker container of telegraf
func (m *TelegrafInstanceManager) StartTelegrafInstance(instanceID string) error {
	return m.client.ContainerStart(m.ctx, instanceID, types.ContainerStartOptions{}); 
}

// Stop the existing docker container of telegraf
func (m *TelegrafInstanceManager) StopTelegrafInstance(instanceID string) error {
	return m.client.ContainerStop(m.ctx, instanceID, nil);
}

// Restart the existing docker container of telegraf
func (m *TelegrafInstanceManager) RestartTelegrafInstance(instanceID string) error {
	err := m.StopTelegrafInstance(instanceID)
	if err != nil {
		return err
	}

	return m.StartTelegrafInstance(instanceID)
}

// Remove the existing docker container of telegraf
func (m *TelegrafInstanceManager) RemoveTelegrafInstance(instanceID string) error {
	err := m.StopTelegrafInstance(instanceID)
	if err != nil && !strings.HasPrefix(err.Error(), "Error response from daemon: No such container: ") {
		return err
	}

	err = m.client.ContainerRemove(m.ctx, instanceID, types.ContainerRemoveOptions{Force: true,});
	if err != nil && !strings.HasPrefix(err.Error(), "Error: No such container: ") {
		return err
	} else {
		return nil
	}
}