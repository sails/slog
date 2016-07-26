package slog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

const (
	// LevelStart 开始标识
	LevelStart = iota
	// LevelDebug debug级别，默认值
	LevelDebug
	// LevelInfo info级别
	LevelInfo
	// LevelWarning warn级别
	LevelWarning
	// LevelError error级别
	LevelError
	// LevelEnd 结束标识
	LevelEnd
)

const (
	// SplitStart 开始标识
	SplitStart = iota
	// NoneSplit 不拆分，默认值
	NoneSplit
	// SplitByDay 按天拆分
	SplitByDay
	// SplitByMonth 按月拆分
	SplitByMonth
	// SplitEnd 结束标识
	SplitEnd
)

var (
	levels = map[int]string{
		LevelDebug:   "DEBUG",
		LevelInfo:    "INFO",
		LevelWarning: "WARNING",
		LevelError:   "ERROR",
	}
	logConfigFile       string
	logConfig           LogConfig
	lastCheckConfigTime int64
	lock                sync.Mutex
	logs                map[string]*SLog
)

// SLog 自定义log
type SLog struct {
	name    string
	log     *log.Logger
	level   int
	model   int
	logfile string
}

// LogConfig 配置信息，程序中保存的
type LogConfig struct {
	DefaultLevel int
	DefaultSplit int
	LogDir       string
	LogLevels    map[string]*ConfigItem
}

// ConfigItem 针对某个单独的log的配置
type ConfigItem struct {
	LogName string
	Level   int
	Model   int
}

// LogFileConfig 用于配置文件中的，与程序中保存的不同，主要是json中更加人性化
type LogFileConfig struct {
	DefaultLevel int
	DefaultSplit int
	LogDir       string
	LogLevels    []ConfigItem
}

func init() {
	if logs == nil {
		logs = make(map[string]*SLog)
	}
	logConfig.DefaultLevel = LevelDebug
	logConfig.DefaultSplit = NoneSplit
	logConfig.LogDir = "./" // 当前目录
	logConfig.LogLevels = make(map[string]*ConfigItem)
}

// SetLogConfigFile 设置配置文件
func SetLogConfigFile(file string) {
	logConfigFile = file
}

// ParserConfig 解析配置
func parserConfig() {
	if len(logConfigFile) == 0 {
		return
	}
	data, err := ioutil.ReadFile(logConfigFile)
	if err != nil {
		return
	}
	config := LogFileConfig{}
	if err := json.Unmarshal(data, &config); err == nil {
		if config.DefaultLevel > LevelStart && config.DefaultLevel < LevelEnd {
			logConfig.DefaultLevel = config.DefaultLevel
		}
		if config.DefaultSplit > SplitStart && config.DefaultSplit < SplitEnd {
			logConfig.DefaultSplit = config.DefaultSplit
		}
		if len(config.LogDir) > 0 {
			logConfig.LogDir = config.LogDir
		}
		for i := 0; i < len(config.LogLevels); i++ {
			cfg := config.LogLevels[i]
			if cfg.Level <= LevelStart && cfg.Level >= LevelEnd {
				cfg.Level = logConfig.DefaultLevel
			}
			if cfg.Model <= SplitStart && cfg.Model >= SplitEnd {
				cfg.Model = logConfig.DefaultSplit
			}
			// 覆盖程序中的配置信息
			logConfig.LogLevels[cfg.LogName] = &cfg
		}

		// 应用到已经存在的日志中
		for name, log := range logs {
			if logConfig.LogLevels[name] == nil {
				continue
			}
			log.level = logConfig.LogLevels[name].Level
			log.level = logConfig.LogLevels[name].Model
		}
	}
}

func getLogFileName(slog *SLog) string {
	// 根据时间创建
	now := time.Now()
	newfile := logConfig.LogDir + slog.name + ".log"
	if slog.model == SplitByDay {
		newfile = fmt.Sprintf("%s%s_%d-%d-%d.log", logConfig.LogDir, slog.name, now.Year(), now.Month(), now.Day())
	} else if slog.model == SplitByMonth {
		newfile = fmt.Sprintf("%s%s_%d-%d.log", logConfig.LogDir, slog.name, now.Year(), now.Month())
	}
	return newfile
}

// GetSLog 根据名称, 时间和level创建log
func getSLog(logName string) *SLog {
	// 设置level，每10秒检查一次
	now := time.Now()
	if now.Unix()-lastCheckConfigTime > 10 {
		lastCheckConfigTime = now.Unix()
		parserConfig()
	}
	if logs[logName] != nil {
		newfile := getLogFileName(logs[logName])
		if newfile != logs[logName].logfile {
			logs[logName] = nil
		} else {
			return logs[logName]
		}
	}
	// 在这里，说明logs[logName]已经是空了
	lock.Lock()
	defer lock.Unlock()
	// 双重校验，防止多线程问题
	if logs[logName] != nil {
		return logs[logName]
	}
	// 创建新的logger
	slog := &SLog{}
	slog.name = logName
	if logConfig.LogLevels[logName] == nil {
		slog.level = logConfig.DefaultLevel
		slog.model = logConfig.DefaultSplit
	} else {
		slog.level = logConfig.LogLevels[logName].Level
		slog.model = logConfig.LogLevels[logName].Model
	}
	filename := getLogFileName(slog)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil
	}
	logger := log.New(file, "", log.LstdFlags|log.Lshortfile)
	slog.log = logger
	slog.logfile = filename
	logs[logName] = slog

	return logs[logName]
}

// SetLogLevel 设置日志输出级别，只有当没有配置文件时生效
// 如果有配置文件，那么就会定期从配置文件中进行覆盖
func SetLogLevel(logName string, level int) {
	if level > LevelError || level < LevelDebug {
		return
	}
	// 覆盖程序中的
	slog := getSLog(logName)
	if slog != nil {
		slog.level = level
	}
	// 覆盖配置信息，因为如果没有配置文件，配置信息不会主动的去更新程序中的logger信息
	if logConfig.LogLevels[logName] != nil {
		config := ConfigItem{}
		config.LogName = logName
		config.Level = level
		config.Model = logConfig.DefaultSplit
		logConfig.LogLevels[logName] = &config
	} else {
		logConfig.LogLevels[logName].Level = level
	}

}

// Debug debug 日志
func Debug(logName string, format string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelDebug {
		return
	}
	slog.log.Printf("[%s] %s", levels[LevelDebug], fmt.Sprintf(format, v...))
}

// DebugPrintln debug日志
func DebugPrintln(logName string, v ...interface{}) {
	// 不直接调用Debug(logName, "%s", fmt.Sprintln(v...))
	// 因为那样做，是先执行了fmt.Sprintln(v...)这个操作，但是
	// 如果日志级别不够的话，是不需要的，浪费性能
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelDebug {
		return
	}
	slog.log.Printf("[%s] %s", levels[LevelDebug], fmt.Sprintln(v...))
}

// Info info 日志
func Info(logName string, format string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelInfo {
		return
	}
	slog.log.Printf("[%s] %s", levels[LevelInfo], fmt.Sprintf(format, v...))
}

// InfoPrintln info 日志
func InfoPrintln(logName string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelInfo {
		return
	}
	slog.log.Printf("[%s] %s", levels[LevelInfo], fmt.Sprintln(v...))
}

// Warn warn 日志
func Warn(logName string, format string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelWarning {
		return
	}
	slog.log.Printf("[%s] %s", levels[LevelWarning], fmt.Sprintf(format, v...))
}

// WarnPrintln warn日志
func WarnPrintln(logName string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelWarning {
		return
	}
	slog.log.Printf("[%s] %s", levels[LevelWarning], fmt.Sprintln(v...))
}

// Error error 日志
func Error(logName string, format string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelError {
		return
	}
	slog.log.Printf("[%s] %s", levels[LevelError], fmt.Sprintf(format, v...))
}

// ErrorPrintln error日志
func ErrorPrintln(logName string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelError {
		return
	}
	slog.log.Printf("[%s] %s", levels[LevelError], fmt.Sprintln(v...))
}
