package logger

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	Level        string `yaml:"level"`         //日志记录等级
	Filename     string `yaml:"filename"`      //文件名称
	MaxSize      int    `yaml:"max_size"`      //文件大小,单位MB
	MaxAge       int    `yaml:"max_age"`       //保留旧文件的最大天数
	MaxBackups   int    `yaml:"max_backups"`   //保留旧文件的最大个数
	LocalTime    bool   `yaml:"local_time"`    //是否使用本地时间,默认UTC
	Compress     bool   `yaml:"compress"`      //日志是否压缩
	AsyncConsole bool   `yaml:"async_console"` //是否同步输出控制台
}

var Logger *zap.Logger

func SetupLogger(logConfig *LogConfig) (err error) {
	var (
		NewLogWriter = func() zapcore.WriteSyncer {
			writer := &lumberjack.Logger{
				Filename:   logConfig.Filename,
				MaxSize:    logConfig.MaxSize,
				MaxAge:     logConfig.MaxAge,
				MaxBackups: logConfig.MaxBackups,
				LocalTime:  logConfig.LocalTime,
				Compress:   logConfig.Compress,
			}
			if logConfig.AsyncConsole {
				return zapcore.AddSync(io.MultiWriter(writer, os.Stdout))
			} else {
				return zapcore.AddSync(writer)
			}
		}
		NewEncoder = func() zapcore.Encoder {
			return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
				TimeKey:        "time",
				LevelKey:       "level",
				NameKey:        "logger",
				CallerKey:      "caller",
				FunctionKey:    zapcore.OmitKey,
				MessageKey:     "msg",
				StacktraceKey:  "stacktrace",
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.CapitalLevelEncoder,
				EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			})
		}
		level = new(zapcore.Level)
	)
	if err = level.UnmarshalText([]byte(logConfig.Level)); err != nil {
		return
	}
	zapCore := zapcore.NewCore(NewEncoder(), NewLogWriter(), level)
	Logger = zap.New(zapCore, zap.AddCaller())
	zap.ReplaceGlobals(Logger)
	return
}
