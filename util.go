package main

import (
	"time"
)

func GetCurrentTimeString() string {
	return time.Now().Format("2006-01-02T15:04:05Z07:00")
}