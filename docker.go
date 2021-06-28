package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
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

func (m *TelegrafInstanceManager) RunTelegrafInstance(configUrl string) (string, error) {
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

	if err := m.client.ContainerStart(m.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return resp.ID, nil
	}

	fmt.Println("Docker instance started:", resp.ID)

	return resp.ID, nil
}