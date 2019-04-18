package logging

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
var Durations = zap.Durations
var Float32 = zap.Float32
var Float32s = zap.Float32s
var Float64 = zap.Float64
var Float64s = zap.Float64s
var Int = zap.Int
var Ints = zap.Ints
var Int8 = zap.Int8
var Int8s = zap.Int8s
var Int16 = zap.Int16
var Int16s = zap.Int16s
var Int32 = zap.Int32
var Int32s = zap.Int32s
var Int64 = zap.Int64
var Int64s = zap.Int64s
var Error = zap.Error
var Errors = zap.Errors
var NamedError = zap.NamedError
var Reflect = zap.Reflect
var String = zap.String
var Strings = zap.Strings
var Time = zap.Time
var Times = zap.Times
var Stack = zap.Stack
var Uint = zap.Uint
var Uints = zap.Uints
var Uint8 = zap.Uint8
var Uint8s = zap.Uint8s
var Uint16 = zap.Uint16
var Uint16s = zap.Uint16s
var Uint32 = zap.Uint32
var Uint32s = zap.Uint32s
var Uint64 = zap.Uint64
var Uint64s = zap.Uint64s

type Level = zapcore.Level

const (
	// DebugLevel logs are usually disabled in production.
	DebugLevel = zapcore.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zapcore.InfoLevel
	// WarnLevel logs are more important than Info.
	WarnLevel = zapcore.WarnLevel
	// ErrorLevel logs are high-priority.
	ErrorLevel = zapcore.ErrorLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zapcore.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zapcore.FatalLevel
)

var defaultConfig string = `{
	"level": "info",
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
	GetEnabled() bool
	GetFormat() string
	GetLevel() string
	GetOutputPaths() []string
	GetErrorOutputPaths() []string
}

type Logger struct {
	zapLogger *zap.Logger
}

func NewLogger(opts LoggerOptions) (*Logger, error) {
	var cfg zap.Config
	if err := json.Unmarshal([]byte(defaultConfig), &cfg); err != nil {
		return nil, err
	}

	if opts != nil && !opts.GetEnabled() {
		l := new(Logger)
		l.zapLogger = zap.NewNop()
		return l, nil
	}

	if opts != nil {
		switch(opts.GetFormat()) {
		case "json":
			cfg.Encoding = "json"
		}
		switch(opts.GetLevel()) {
		case "debug":
			cfg.Level.SetLevel(zapcore.DebugLevel)
		case "info":
			cfg.Level.SetLevel(zapcore.InfoLevel)
		case "warn":
			cfg.Level.SetLevel(zapcore.WarnLevel)
		case "error":
			cfg.Level.SetLevel(zapcore.ErrorLevel)
		case "panic":
			cfg.Level.SetLevel(zapcore.PanicLevel)
		case "fatal":
			cfg.Level.SetLevel(zapcore.FatalLevel)
		}
		// customize OutputPaths
		outputPaths := opts.GetOutputPaths()
		if outputPaths != nil && len(outputPaths) > 0 {
			cfg.OutputPaths = outputPaths
		}
		// customize ErrorOutputPaths
		errorOutputPaths := opts.GetErrorOutputPaths()
		if errorOutputPaths != nil && len(errorOutputPaths) > 0 {
			cfg.ErrorOutputPaths = errorOutputPaths
		}
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

func (l *Logger) Log(level Level, msg string, fields ...Field) {
	l.assertReady()
	switch(level) {
	case DebugLevel:
		l.zapLogger.Debug(msg, fields...)
	case InfoLevel:
		l.zapLogger.Info(msg, fields...)
	case WarnLevel:
		l.zapLogger.Warn(msg, fields...)
	case ErrorLevel:
		l.zapLogger.Error(msg, fields...)
	case PanicLevel:
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
