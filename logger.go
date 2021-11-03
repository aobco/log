package log

import (
	"fmt"
	"github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

const (
	RollingBySize = iota
	RollingByDate
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	// # configure specific data zone
	//loc, err := time.LoadLocation("Asia/Shanghai")
	//if err != nil {
	//	Errorf("time load location [Asia/Shanghai] fail %v", err)
	//	loc = time.FixedZone("CST", 8*3600)
	//}
	//enc.AppendString(t.In(loc).Format("2006-01-02 15:04:05.000"))
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func InitZapLog(filename string, logLevel string, maxSize int, maxBackups int, maxAge int, rollingBy int) {
	var fileWriterSyncer zapcore.WriteSyncer
	if rollingBy == RollingBySize {
		fileWriterSyncer = zapcore.AddSync(&lumberjack.Logger{
			Filename:   filename,
			MaxSize:    maxSize, // MB
			LocalTime:  true,
			MaxBackups: maxBackups, // file number
			MaxAge:     maxAge,     // day

		})
	} else {
		rotate, err := RotateLogs(filename)
		if err != nil {
			panic(err)
		}
		fileWriterSyncer = zapcore.AddSync(rotate)
	}
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.EncodeTime = TimeEncoder // zapcore.ISO8601TimeEncoder
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	devEncoderConfig := zap.NewDevelopmentEncoderConfig()
	devEncoderConfig.EncodeTime = TimeEncoder
	devEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // color

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(fileEncoderConfig), fileWriterSyncer, zap.NewAtomicLevel()),
		zapcore.NewCore(zapcore.NewConsoleEncoder(devEncoderConfig), zapcore.WriteSyncer(os.Stdout), zap.NewAtomicLevel()),
	)
	level := zapcore.DebugLevel
	switch string(logLevel) {
	case "debug", "DEBUG":
		level = zapcore.DebugLevel
	case "info", "INFO", "": // make the zero value useful
		level = zapcore.DebugLevel
	case "warn", "WARN":
		level = zapcore.DebugLevel
	case "error", "ERROR":
		level = zapcore.DebugLevel
	case "dpanic", "DPANIC":
		level = zapcore.DebugLevel
	case "panic", "PANIC":
		level = zapcore.DebugLevel
	case "fatal", "FATAL":
		level = zapcore.DebugLevel
	default:
		fmt.Printf("invalid log level %s", logLevel)
	}
	core.Enabled(level)
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Sugar = Logger.Sugar()
}

var once sync.Once

func Default() {
	once.Do(func() {
		println("init default logger to standard output")
		devEncoderConfig := zap.NewDevelopmentEncoderConfig()
		devEncoderConfig.EncodeTime = TimeEncoder
		devEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // color
		core := zapcore.NewCore(zapcore.NewConsoleEncoder(devEncoderConfig), zapcore.WriteSyncer(os.Stdout), zap.NewAtomicLevel())
		core.Enabled(zapcore.DebugLevel)
		Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
		Sugar = Logger.Sugar()
	})
}

func RotateLogs(filePath string) (*rotatelogs.RotateLogs, error) {
	filename := filePath + ".%Y%m%d"
	retate, err := rotatelogs.New(filename, rotatelogs.WithLinkName(filePath), rotatelogs.WithMaxAge(time.Hour*24*3), rotatelogs.WithRotationTime(time.Hour*24))
	return retate, err
}

func Debug(args ...interface{}) {
	if Sugar == nil {
		Default()
	}
	Sugar.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	if Sugar == nil {
		Default()
	}
	Sugar.Debugf(template, args...)
}

func Info(args ...interface{}) {
	if Sugar == nil {
		Default()
	}
	Sugar.Info(args...)
}

func Infof(template string, args ...interface{}) {
	if Sugar == nil {
		Default()
	}
	Sugar.Infof(template, args...)
}

func Warn(args ...interface{}) {
	if Sugar == nil {
		Default()
	}
	Sugar.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	if Sugar == nil {
		Default()
	}
	Sugar.Warnf(template, args...)
}

func Error(args ...interface{}) {
	if Sugar == nil {
		Default()
	}
	args = append(args, string(debug.Stack()))
	Sugar.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	if Sugar == nil {
		Default()
	}
	args = append(args, string(debug.Stack()))
	Sugar.Errorf(template+"\n", args...)
}

func Panic(args ...interface{}) {
	if Sugar == nil {
		Default()
	}
	args = append(args, string(debug.Stack()))
	Sugar.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	if Sugar == nil {
		Default()
	}
	args = append(args, string(debug.Stack()))
	Sugar.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	if Sugar == nil {
		Default()
	}
	args = append(args, string(debug.Stack()))
	Sugar.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	if Sugar == nil {
		Default()
	}
	args = append(args, string(debug.Stack()))
	Sugar.Fatalf(template, args...)
}
