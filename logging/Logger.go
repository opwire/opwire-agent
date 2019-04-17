package logging

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
)

type Field = zap.Field
var Any = zap.Any
var Array = zap.Array
var Binary = zap.Binary
var Bool = zap.Bool
var Bools = zap.Bools
var ByteString = zap.ByteString
var ByteStrings = zap.ByteStrings
var Complex128 = zap.Complex128
var Complex128s = zap.Complex128s
var Complex64 = zap.Complex64
var Complex64s = zap.Complex64s
var Duration = zap.Duration
var Error = zap.Error
var Errors = zap.Errors
var String = zap.String

type LogLevel int

const (
	_ LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	PANIC
	FATAL
)

var defaultConfig string = `{
	"level": "debug",
	"encoding": "console",
	"outputPaths": ["stdout"],
	"errorOutputPaths": ["stderr"],
	"encoderConfig": {
		"messageKey": "message",
		"levelKey": "level",
		"levelEncoder": "capital",
		"timeKey": "logtime",
		"timeEncoder": "ISO8601"
	}
}`

type LoggerOptions interface {
}

type Logger struct {
	zapLogger *zap.Logger
}

func NewLogger(opts LoggerOptions) (*Logger, error) {
	var cfg zap.Config
	if err := json.Unmarshal([]byte(defaultConfig), &cfg); err != nil {
		return nil, err
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	l := new(Logger)
	l.zapLogger = logger

	return l, nil
}

func (l *Logger) assertReady() *Logger {
	if l.zapLogger == nil {
		panic(fmt.Errorf("Internal Logger must not be nil"))
	}
	return l
}

func (l *Logger) Log(level LogLevel, msg string, fields ...Field) {
	l.assertReady()
	switch(level) {
	case DEBUG:
		l.zapLogger.Debug(msg, fields...)
	case INFO:
		l.zapLogger.Info(msg, fields...)
	case WARN:
		l.zapLogger.Warn(msg, fields...)
	case ERROR:
		l.zapLogger.Error(msg, fields...)
	case PANIC:
		l.zapLogger.Panic(msg, fields...)
	default:
		l.zapLogger.Fatal(msg, fields...)
	}
}

func (l *Logger) With(fields ...Field) *Logger {
	l.assertReady()
	branch := &Logger{}
	branch.zapLogger = l.zapLogger.With(fields...)
	return branch
}

func (l *Logger) Sync() error {
	l.assertReady()
	return l.zapLogger.Sync()
}
