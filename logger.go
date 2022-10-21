package log

import (
	"fmt"
	"github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

const (
	dataRollingSuffix = ".%Y%m%d"

	RollingBySize = iota
	RollingByDate
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
	once   sync.Once
)

// zapcore.ISO8601TimeEncoder
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	//loc, err := time.LoadLocation("Asia/Shanghai")
	//if err != nil {
	//	Errorf("time load location [Asia/Shanghai] fail %v", err)
	//	loc = time.FixedZone("CST", 8*3600)
	//}
	//enc.AppendString(t.In(loc).Format("2006-01-02 15:04:05.000"))
	enc.AppendString(t.Format(time.RFC3339))
}

func DateRolling(filename string, logLevel string, maxBackups, maxAge int, stdout ...bool) {
	rotate, err := RotateLogs(filename, uint(maxBackups), maxAge)
	if err != nil {
		panic(err)
	}
	level := logLv(logLevel)
	cores := make([]zapcore.Core, 0)
	fileWriterSyncer := zapcore.AddSync(rotate)
	logCore(fileWriterSyncer, level, &cores)
	devCore(stdout, level, &cores)
	core := zapcore.NewTee(cores...)
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Sugar = Logger.Sugar()
}

func SizeRolling(filename string, logLevel string, maxSize, maxBackups, maxAge int, stdout ...bool) {
	level := logLv(logLevel)
	cores := make([]zapcore.Core, 0)
	fileWriterSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize, // MB
		LocalTime:  true,
		MaxBackups: maxBackups,
		MaxAge:     maxAge, // Day

	})
	logCore(fileWriterSyncer, level, &cores)
	devCore(stdout, level, &cores)
	core := zapcore.NewTee(cores...)
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Sugar = Logger.Sugar()
}

func devCore(stdout []bool, level zapcore.Level, cores *[]zapcore.Core) {
	if len(stdout) > 0 && stdout[0] {
		devEncoderConfig := zap.NewDevelopmentEncoderConfig()
		devEncoderConfig.EncodeTime = timeEncoder
		devEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		c := zapcore.NewCore(zapcore.NewConsoleEncoder(devEncoderConfig), zapcore.WriteSyncer(os.Stdout), level)
		*cores = append(*cores, c)
	}
}

func logCore(fileWriterSyncer zapcore.WriteSyncer, level zapcore.Level, cores *[]zapcore.Core) {
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.EncodeTime = timeEncoder
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	c := zapcore.NewCore(zapcore.NewConsoleEncoder(fileEncoderConfig), fileWriterSyncer, level)
	*cores = append(*cores, c)
}

func logLv(logLevel string) zapcore.Level {
	level := zapcore.InfoLevel
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO", "": // make the zero value useful
		level = zapcore.InfoLevel
	case "WARN":
		level = zapcore.WarnLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	case "DPANIC":
		level = zapcore.DPanicLevel
	case "PANIC":
		level = zapcore.PanicLevel
	case "FATAL":
		level = zapcore.FatalLevel
	default:
		fmt.Printf("invalid log level %s", logLevel)
	}
	return level
}

func Init(filename string, logLevel string, maxSize int, maxBackups int, maxAge int, rollingBy int, stdout ...bool) {
	switch rollingBy {
	case RollingBySize:
		SizeRolling(filename, logLevel, maxSize, maxBackups, maxAge, stdout...)
	default:
		DateRolling(filename, logLevel, maxBackups, maxAge, stdout...)
	}
}

func Default() {
	once.Do(func() {
		println("pipe log to stdout")
		devEncoderConfig := zap.NewDevelopmentEncoderConfig()
		devEncoderConfig.EncodeTime = timeEncoder
		devEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // color
		core := zapcore.NewCore(zapcore.NewConsoleEncoder(devEncoderConfig), zapcore.WriteSyncer(os.Stdout), zap.DebugLevel)
		Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
		Sugar = Logger.Sugar()
	})
}

func RotateLogs(filePath string, maxBackups uint, maxAge int) (*rotatelogs.RotateLogs, error) {
	var filename string
	ext := filepath.Ext(filePath)
	if len(ext) > 0 {
		filename = strings.TrimSuffix(filePath, ext) + dataRollingSuffix + ext
	} else {
		filename = filePath + dataRollingSuffix
	}
	options := []rotatelogs.Option{rotatelogs.WithLinkName(filePath), rotatelogs.WithRotationTime(time.Hour * 24)}
	if int(maxBackups) > maxAge {
		options = append(options, rotatelogs.WithMaxAge(time.Hour*24*time.Duration(maxAge)))
	} else {
		options = append(options, rotatelogs.WithRotationCount(maxBackups))
	}
	return rotatelogs.New(filename, options...)
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

/*
	1. print err msg
	2. exit application
	3. defer won't be executed
*/
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
