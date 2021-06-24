package main

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const TELEGRAF_DOCKER_IMAGE_TAG = "1.18.3-alpine"
const TELEGRAF_DOCKER_IMAGE_NAME = "telegraf"

type TelegrafInstanceManager struct {
	Client *client.Client
	ctx context.Context
}

func NewTelegrafIntanceManager() *TelegrafInstanceManager {
	return &TelegrafInstanceManager{
		Client: nil,
	}
}

func (m *TelegrafInstanceManager) Start() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	m.Client = cli

	m.ctx = context.Background()

	if _, err := cli.ContainerList(m.ctx, types.ContainerListOptions{}); err != nil {
		return err
	}

	return nil
}