package slog

import (
	"testing"
)

func TestDebugLog(t *testing.T) {
	SetLogConfigFile("./test.json")
	DebugLog("test", "hello %s", "sails")
	InfoLog("test", "hello %s", "sails")
	WarnLog("test", "hello %s", "sails")
	ErrorLog("test", "hello %s", "sails")

	DebugLog("tt", "hello %s", "sails")
	InfoLog("tt", "hello %s", "sails")
	WarnLog("tt", "hello %s", "sails")
	ErrorLog("tt", "hello %s", "sails")
}
