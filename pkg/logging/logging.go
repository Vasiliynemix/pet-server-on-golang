package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type Logger struct {
	*zap.Logger
}

func NewLogger(
	logger *zap.Logger,
) *Logger {
	return &Logger{
		Logger: logger,
	}
}

func InitLogger(
	structDateFormat string,
	pathToInfoLog string,
	pathToDebugLog string,
	logLevel string,
) *zap.Logger {
	log := getFileLogger(structDateFormat, pathToInfoLog, pathToDebugLog, logLevel)
	defer log.Sync()
	return log
}

func getFileLogger(
	structDateFormat string,
	pathToInfoLog string,
	pathToDebugLog string,
	levelString string,
) *zap.Logger {
	cfgLogger := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		CallerKey:      "caller",
		MessageKey:     "message",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(structDateFormat),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	fileEncoder := zapcore.NewJSONEncoder(cfgLogger)
	consoleEncoder := zapcore.NewConsoleEncoder(cfgLogger)
	consoleWriter := zapcore.AddSync(os.Stdout)

	logFileInfo, _ := os.OpenFile(pathToInfoLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	logFileDebug, _ := os.OpenFile(pathToDebugLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	logLevelCfg := levelString

	var writer zapcore.WriteSyncer
	var logLevel zapcore.Level

	switch logLevelCfg {
	case "debug":
		writer = zapcore.AddSync(logFileDebug)
		logLevel = zapcore.DebugLevel
	case "info":
		writer = zapcore.AddSync(logFileInfo)
		logLevel = zapcore.InfoLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, writer, logLevel),
		zapcore.NewCore(consoleEncoder, consoleWriter, logLevel),
	)

	log := zap.New(core, zap.AddCaller())

	return log
}
