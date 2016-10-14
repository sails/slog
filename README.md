# slog
golang log with file split by day and logLevel controll

## usage
```
import slog

slog.Debug("test", "hello %s", "sails")
slog.Info("test", "hello %s", "sails")
slog.Warn("test", "hello %s", "sails")
slog.Error("test", "hello %s", "sails")
```
日志的第一个参数是日志名，第二个参数是格式化字符串，后面的参数是对应的值


slog 默认会在当前目录下创建一个log目录用于保存日志文件，比如:
```
slog.Debug("test", "hello %s", "sails")
```
将在当前目录创建test.log，把所有的内容输出到文件中

## configure
slog可以通过指定一个json的配置文件进行配置，有几个选项可提供配置:
+ Out，用于指定输出的类型，有两个值"file,console"，可以同时指定.默认是FILE。
+ Level, 用于指定输入的级别，有1,2,3,4分别对应debug,info, warn,error；默认是debug。
+ FileSplit, 当Out指定了FILE类型时，用于表明文件的拆分形式，分三种：不拆分，按天，按月；默认是不拆分。
+ FileDir，用于指定输入的目录，默认是当前目录。
+ LogLevels，它是一个json结构，用于单独指定日志配置,有以下几个配置：
    - LogName，要单独配置的日志名
    - Level，单独配置的输出级别
    - FileSplit，单独的文件拆分形式
    - FileName，日志对应的文件名，默认文件名也文件名对应

example, config.json:
```
{
    "Out":"console, file",
    "Level":2,
    "FileSplit":2,
    "FileDir":"./log/",
    "LogLevels":[
		{"LogName":"test","Level":2, "FileSplit":2, "FileName":"test"}
    ]
}
```
```
slog.SetLogConfigFile("./config.json")
```
