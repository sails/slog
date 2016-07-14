package slog

import (
	"testing"
)

func TestDebugLog(t *testing.T) {
	SetLogConfigFile("./test.json")
	Debug("test", "hello %s", "sails")
	Info("test", "hello %s", "sails")
	Warn("test", "hello %s", "sails")
	Error("test", "hello %s", "sails")

	Debug("tt", "hello %s", "sails")
	Info("tt", "hello %s", "sails")
	Warn("tt", "hello %s", "sails")
	Error("tt", "hello %s", "sails")
}
