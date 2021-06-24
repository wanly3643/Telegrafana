package main

import (
	//"github.com/burntsushi/toml"
)

type TelegrafConfigManager struct {
}

func NewTelegrafConfigManager() *TelegrafConfigManager {
	return &TelegrafConfigManager{
	}
}

func (m *TelegrafConfigManager) Start() error {
	return nil
}