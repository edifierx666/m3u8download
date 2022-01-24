/*
   logPath 日志文件路径
   logLevel 日志级别 debug/info/warn/error，debug输出：debug/info/warn/error日志。 info输出：info/warn/error日志。 warn输出：warn/error日志。 error输出：error日志。
   maxSize 单个文件大小,MB
   maxBackups 保存的文件个数
   maxAge 保存的天数， 没有的话不删除
   compress 压缩
   jsonFormat 是否输出为json格式
   shoowLine 显示代码行
   logInConsole 是否同时输出到控制台
*/

package utils

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

var Logger *zap.Logger

func init() {
	initLogger("", "debug", 0, 10, 0, false, false, true, true)
}

func initLogger(logPath string, logLevel string, maxSize, maxBackups, maxAge int, compress, jsonFormat, showLine, logInConsole bool) {
	// 设置日志级别
	var level zapcore.Level
	switch logLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	var (
		syncer zapcore.WriteSyncer
		// 自定义时间输出格式
		customTimeEncoder = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString("[" + t.Format("2006-01-02 15:04:05.000") + "]")
		}
		// 自定义日志级别显示
		customLevelEncoder = func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString("[" + level.CapitalString() + "]")
		}
	)

	// 定义日志切割配置
	hook := lumberjack.Logger{
		Filename:   logPath,    // 日志文件的位置
		MaxSize:    maxSize,    // 在进行切割之前，日志文件的最大大小（以MB为单位）
		MaxBackups: maxBackups, // 保留旧文件的最大个数
		Compress:   compress,   // 是否压缩 disabled by default
	}
	if maxAge > 0 {
		hook.MaxAge = maxAge // days
	}

	// 判断是否控制台输出日志
	if logInConsole {
		syncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook))
	} else {
		syncer = zapcore.AddSync(&hook)
	}

	// 定义zap配置信息
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     customTimeEncoder,          // 自定义时间格式
		EncodeLevel:    customLevelEncoder,         // 小写编码器
		EncodeCaller:   zapcore.ShortCallerEncoder, // 全路径编码器
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	var encoder zapcore.Encoder
	// 判断是否json格式输出
	if jsonFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		syncer,
		level,
	)

	Logger = zap.New(core)
	// 判断是否显示代码行号
	if showLine {
		Logger = Logger.WithOptions(zap.AddCaller())
	}
}
