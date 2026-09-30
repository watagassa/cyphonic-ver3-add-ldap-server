package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// FilePerm600 is the permission for log file.
const FilePerm600 = 0x600

// InitZap provides logging with zap.
func InitZap(cfg *Config) (func(), error) {
	borderLogLevel := zap.ErrorLevel

	instanceID, err := getInstanceID()
	if err != nil {
		return nil, fmt.Errorf("failed to create instance_id: %w", err)
	}

	// Setting for standard output log
	stdCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig(true)),
		zapcore.AddSync(os.Stdout),
		map[bool]zapcore.LevelEnabler{true: zap.DebugLevel, false: zap.InfoLevel}[cfg.DebugMode],
	)

	// Setting of file log for error
	errorfile, err := setFile(cfg.ErrorLogFilePath)
	if err != nil {
		return nil, fmt.Errorf("file creation failed: %w", err)
	}

	errorCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig(false)),
		zapcore.AddSync(errorfile),
		borderLogLevel,
	)

	// Create common fields
	commonFields := zap.Fields(
		zap.String("service_type", cfg.ServiceType),
		zap.String("instance_id", instanceID),
		zap.String("tls_port", cfg.TLSPort),
		zap.String("quic_port", cfg.QUICPort),
	)

	// The setting of file log for debugging
	debugfile, err := setFile(cfg.DebugLogFilePath)
	if err != nil {
		return nil, fmt.Errorf("file creation failed: %w", err)
	}

	debugCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig(false)),
		zapcore.AddSync(debugfile),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			if cfg.DebugMode {
				return lvl < borderLogLevel
			}

			return lvl < borderLogLevel && lvl != zap.DebugLevel
		}),
	)

	// Create logger
	cores := zapcore.NewTee(stdCore, errorCore, debugCore)
	logger := zap.New(
		cores,
		commonFields,
	)

	if cfg.DebugMode {
		logger = zap.New(
			cores,
			zap.AddStacktrace(zap.ErrorLevel),
			commonFields,
		)
	}

	reset := zap.ReplaceGlobals(logger)

	return reset, nil
}

func encoderConfig(isConsole bool) zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()

	cfg.MessageKey = "msg"
	cfg.LevelKey = "level"
	cfg.NameKey = "name"
	cfg.TimeKey = "timestamp"
	cfg.CallerKey = "caller"
	cfg.FunctionKey = "func"
	cfg.StacktraceKey = "stacktrace"
	cfg.LineEnding = "\n"
	cfg.EncodeTime = map[bool]zapcore.TimeEncoder{true: zapcore.ISO8601TimeEncoder, false: zapcore.EpochTimeEncoder}[isConsole]
	cfg.EncodeLevel = map[bool]zapcore.LevelEncoder{true: zapcore.CapitalColorLevelEncoder, false: zapcore.LowercaseLevelEncoder}[isConsole]
	cfg.EncodeDuration = zapcore.SecondsDurationEncoder
	cfg.EncodeCaller = zapcore.ShortCallerEncoder

	return cfg
}

func setFile(filePath string) (*os.File, error) {
	// Get Parent dir
	dir := filepath.Dir(filePath)

	// Create log dir
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("can't create directory: %w", err)
	}

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

func getInstanceID() (string, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return "", fmt.Errorf("can't generate uuid: %w", err)
	}

	return id.String(), nil
}

// LogDebug outputs log with DEBUG LEVEL.
// The arguments kv... are key-value pairs of additional information.
func LogDebug(msg string, kv ...interface{}) {
	zap.S().Debugw(msg, kv...)
}

// LogInfo outputs log with INFO LEVEL.
// The arguments kv... are key-value pairs of additional information.
func LogInfo(msg string, kv ...interface{}) {
	zap.S().Infow(msg, kv...)
}

// LogWarn outputs log with WARN LEVEL.
// The arguments kv... are key-value pairs of addtional information.
func LogWarn(msg string, kv ...interface{}) {
	zap.S().Warnw(msg, kv...)
}

// LogErr outputs log with ERROR LEVEL.
// The arguments kv... are key-value pairs of additional information.
func LogErr(msg string, kv ...interface{}) {
	zap.S().Errorw(msg, kv...)
}
