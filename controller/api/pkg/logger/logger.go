package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const FilePerm600 = 0x600

// InitZap provides logging with zap.
func InitZap(isDebug bool) error {
	var cors zapcore.Core

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	errorFile, err := setFile("/var/log/controller-error.json")
	if err != nil {
		return fmt.Errorf("setFile function failed: %w", err)
	}

	errCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(config()),
		zapcore.AddSync(errorFile),
		highPriority,
	)

	if isDebug {
		debugFile, err := setFile("/var/log/controller-debug.json")
		if err != nil {
			return fmt.Errorf("setFile function failed: %w", err)
		}

		debugCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(config()),
			zapcore.AddSync(debugFile),
			lowPriority,
		)
		cors = zapcore.NewTee(
			debugCore,
			errCore,
		)
	} else {
		cors = zapcore.NewTee(
			errCore,
		)
	}

	zap.ReplaceGlobals(zap.New(cors))

	return nil
}

func setFile(filePath string) (*os.File, error) {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			if _, err := os.Create(filePath); err != nil {
				return nil, fmt.Errorf("can't create log file: %w", err)
			}
		}
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, FilePerm600)
	if err != nil {
		return nil, fmt.Errorf("can't open log file: %w", err)
	}

	return file, nil
}

func config() zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()

	cfg.MessageKey = "msg"
	cfg.LevelKey = "level"
	cfg.NameKey = "name"
	cfg.TimeKey = "timestamp"
	cfg.CallerKey = "caller"
	cfg.FunctionKey = "func"
	cfg.StacktraceKey = "stacktrace"
	cfg.LineEnding = "\n"
	cfg.EncodeTime = zapcore.EpochTimeEncoder
	cfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	cfg.EncodeDuration = zapcore.SecondsDurationEncoder
	cfg.EncodeCaller = zapcore.ShortCallerEncoder

	return cfg
}

func LogDebug(msg string, kv ...interface{}) {
	zap.S().Debugw(msg, kv...)
}

func LogErr(msg string, kv ...interface{}) {
	zap.S().Errorw(msg, kv...)
}

func LogInfo(msg string, kv ...interface{}) {
	zap.S().Infow(msg, kv...)
}
