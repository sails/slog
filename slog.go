package slog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// OutputConsole 向stdout中输出
	OutputConsole = 1
	// OutputFile 向file中输出
	OutputFile = 2
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

const (
	// DefaltLevel 默认输出级别
	DefaltLevel = LevelDebug
	// DefaultSplit 默认日志文件拆分方式
	DefaultSplit = NoneSplit
	// DefaultLogDir 默认日志文件输出目录
	DefaultLogDir = "./log/"
)

var (
	levels = map[int]string{
		LevelDebug:   "DEBUG",
		LevelInfo:    "INFO",
		LevelWarning: "WARNING",
		LevelError:   "ERROR",
	}
	logConfigFile       = "log.json"
	logConfig           LogConfig
	lastCheckConfigTime int64
	lock                sync.Mutex
	logs                map[string]*SLog
)

// SLog 自定义log
type SLog struct {
	name string
	// file logger
	log       *log.Logger
	level     int
	filesplit int
	// 保存的文件名，如果没有指定，默认是name
	filename string
	logfile  string
	file     *os.File

	// console
	console *log.Logger
}

// LogConfig 配置信息，程序中保存的
type LogConfig struct {
	Out       int
	Level     int
	FileSplit int
	FileDir   string
	LogLevels map[string]*ConfigItem
}

// ConfigItem 针对某个单独的log的配置
type ConfigItem struct {
	LogName   string
	Level     int
	FileSplit int
	FileName  string
}

// LogFileConfig 用于配置文件中的，与程序中保存的不同，主要是json中更加人性化
type LogFileConfig struct {
	Out       string
	Level     int
	FileSplit int
	FileDir   string
	LogLevels []ConfigItem
}

func init() {
	if logs == nil {
		logs = make(map[string]*SLog)
	}
	logConfig.Out = OutputFile
	logConfig.Level = DefaltLevel
	logConfig.FileSplit = DefaultSplit
	logConfig.FileDir = DefaultLogDir
	logConfig.LogLevels = make(map[string]*ConfigItem)
	// 创建对应的目录
	if !Exist(logConfig.FileDir) {
		os.MkdirAll(logConfig.FileDir, 0777)
	}
}

// 检查文件或目录是否存在
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

// SetLogConfigFile 设置配置文件
func SetLogConfigFile(file string) {
	logConfigFile = file
}

// ParserConfig 解析配置
func parserConfig() {
	lock.Lock()
	defer lock.Unlock()
	if len(logConfigFile) == 0 {
		return
	}
	data, err := ioutil.ReadFile(logConfigFile)
	if err != nil {
		return
	}
	config := LogFileConfig{}
	if err := json.Unmarshal(data, &config); err == nil {
		logConfig.Out = 0
		if len(config.Out) > 0 { // 有配置
			out := strings.ToUpper(config.Out)
			if strings.Contains(out, "FILE") {
				logConfig.Out = OutputFile
			}
			if strings.Contains(out, "CONSOLE") {
				logConfig.Out = logConfig.Out | OutputConsole
			}
		}

		if logConfig.Out == 0 {
			logConfig.Out = OutputFile
		}
		if config.Level > LevelStart && config.Level < LevelEnd {
			logConfig.Level = config.Level
		} else {
			logConfig.Level = DefaltLevel
		}
		if config.FileSplit > SplitStart && config.FileSplit < SplitEnd {
			logConfig.FileSplit = config.FileSplit
		} else {
			logConfig.FileSplit = DefaultSplit
		}
		if len(config.FileDir) > 0 {
			logConfig.FileDir = config.FileDir
			// 创建对应的目录
			if !Exist(logConfig.FileDir) {
				os.MkdirAll(logConfig.FileDir, 0777)
			}
		} else {
			logConfig.FileDir = DefaultLogDir
		}
		loglevels := make(map[string]*ConfigItem)
		for i := 0; i < len(config.LogLevels); i++ {
			cfg := config.LogLevels[i]
			if len(cfg.LogName) == 0 {
				continue
			}
			if cfg.Level <= LevelStart || cfg.Level >= LevelEnd {
				cfg.Level = logConfig.Level
			}
			if cfg.FileSplit <= SplitStart || cfg.FileSplit >= SplitEnd {
				cfg.FileSplit = logConfig.FileSplit
			}
			if len(cfg.FileName) == 0 {
				cfg.FileName = cfg.LogName
			}
			loglevels[cfg.LogName] = &cfg
		}
		// 覆盖程序中的配置信息
		logConfig.LogLevels = loglevels
	}
}

func applyConfig() {
	lock.Lock()
	defer lock.Unlock()
	// 应用到已经存在的日志中
	for name, log := range logs {
		if log == nil {
			continue
		}
		if logConfig.LogLevels[name] == nil {
			log.level = logConfig.Level
			log.filesplit = logConfig.FileSplit
			log.filename = log.name
		} else {
			log.level = logConfig.LogLevels[name].Level
			log.filesplit = logConfig.LogLevels[name].FileSplit
			log.filename = logConfig.LogLevels[name].FileName
		}
	}
}

func getLogFileFullPath(slog *SLog) string {
	// 根据时间创建
	now := time.Now()

	if len(slog.filename) == 0 {
		slog.filename = slog.name
	}
	newfile := logConfig.FileDir + "/" + slog.filename + ".log"
	if slog.filesplit == SplitByDay {
		newfile = fmt.Sprintf("%s/%s_%d-%d-%d.log", logConfig.FileDir, slog.filename, now.Year(), now.Month(), now.Day())
	} else if slog.filesplit == SplitByMonth {
		newfile = fmt.Sprintf("%s/%s_%d-%d.log", logConfig.FileDir, slog.filename, now.Year(), now.Month())
	}
	return newfile
}

func newSlog(logName string) *SLog {
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
		slog.level = logConfig.Level
		slog.filesplit = logConfig.FileSplit
		slog.filename = slog.name
	} else {
		slog.level = logConfig.LogLevels[logName].Level
		slog.filesplit = logConfig.LogLevels[logName].FileSplit
		slog.filename = logConfig.LogLevels[logName].FileName
	}
	if (logConfig.Out & OutputFile) == OutputFile {
		filename := getLogFileFullPath(slog)

		file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		logger := log.New(file, "", log.LstdFlags|log.Lshortfile)
		slog.log = logger
		slog.logfile = filename
		slog.file = file
	}
	if (logConfig.Out & OutputConsole) == OutputConsole {
		console := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
		slog.console = console
	}

	return slog
}

// GetSLog 根据名称, 时间和level创建log
func getSLog(logName string) *SLog {
	// 设置level，每10秒检查一次
	now := time.Now()
	if now.Unix()-lastCheckConfigTime > 10 {
		lastCheckConfigTime = now.Unix()
		parserConfig()
		applyConfig()
	}
	if logs[logName] != nil {
		// 只要配置变更或者输出文件改变了，就重新创建
		reCreate := false
		// 控制台改变
		if (logConfig.Out & OutputConsole) == OutputConsole {
			if logs[logName].console == nil {
				reCreate = true
			}
		} else {
			logs[logName].console = nil
		}
		// 文件日志格式或者日期导致的文件名改变
		if !reCreate {
			if (logConfig.Out & OutputFile) == OutputFile {
				newfile := getLogFileFullPath(logs[logName])
				if newfile != logs[logName].logfile {
					reCreate = true
				}
			} else {
				if logs[logName].log != nil && logs[logName].file != nil {
					logs[logName].file.Close()
				}
				logs[logName].file = nil
				logs[logName].logfile = ""
				logs[logName].log = nil
			}
		}

		if !reCreate {
			return logs[logName]
		}

		// 需要重建
		if logs[logName].log != nil && logs[logName].file != nil {
			logs[logName].file.Close()
		}
		logs[logName] = nil
	}
	// 说明找不到日志，则新建一个
	logs[logName] = newSlog(logName)
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
		config.FileSplit = logConfig.FileSplit
		logConfig.LogLevels[logName] = &config
	} else {
		logConfig.LogLevels[logName].Level = level
	}

}

func output(slog *SLog, levelstr string, content string) {
	if slog.log != nil {
		slog.log.Printf("[%s] %s", levelstr, content)
	}
	if slog.console != nil {
		slog.console.Printf("[%s] [%s] %s", slog.name, levelstr, content)
	}
}

// Debug debug 日志
func Debug(logName string, format string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelDebug {
		return
	}
	output(slog, levels[LevelDebug], fmt.Sprintf(format, v...))
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
	output(slog, levels[LevelDebug], fmt.Sprintln(v...))
}

// Info info 日志
func Info(logName string, format string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelInfo {
		return
	}
	output(slog, levels[LevelInfo], fmt.Sprintf(format, v...))
}

// InfoPrintln info 日志
func InfoPrintln(logName string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelInfo {
		return
	}
	output(slog, levels[LevelInfo], fmt.Sprintln(v...))
}

// Warn warn 日志
func Warn(logName string, format string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelWarning {
		return
	}
	output(slog, levels[LevelWarning], fmt.Sprintf(format, v...))
}

// WarnPrintln warn日志
func WarnPrintln(logName string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelWarning {
		return
	}
	output(slog, levels[LevelWarning], fmt.Sprintln(v...))
}

// Error error 日志
func Error(logName string, format string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelError {
		return
	}
	output(slog, levels[LevelError], fmt.Sprintf(format, v...))
}

// ErrorPrintln error日志
func ErrorPrintln(logName string, v ...interface{}) {
	slog := getSLog(logName)
	if slog == nil || slog.level > LevelError {
		return
	}
	output(slog, levels[LevelError], fmt.Sprintln(v...))
}
