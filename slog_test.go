package slog

import (
	"testing"
	"time"
)

func TestDebugLog(t *testing.T) {
	SetLogConfigFile("./test.json")
	Debug("test", "hello %s", "sails")
	Info("test", "hello %s", "sails")
	Warn("test", "hello %s", "sails")
	Error("test", "hello %s", "sails")
	for i := 0; i < 100; i++ {
		Debug("tt", "hello %s", "sails")
		Info("tt", "hello %s", "sails")
		Warn("tt", "hello %s", "sails")
		Error("tt", "hello %s", "sails")
		Debug("t", "hello %s", "sails")
		Info("t", "hello %s", "sails")
		Warn("t", "hello %s", "sails")
		Error("t", "hello %s", "sails")
		time.Sleep(2 * time.Second)
	}

}
