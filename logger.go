package logger

import (
	"fmt"
	"log"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logLevels = map[string]zapcore.Level{
		"DEBUG":  zapcore.DebugLevel,
		"INFO":   zapcore.InfoLevel,
		"WARN":   zapcore.WarnLevel,
		"ERROR":  zapcore.ErrorLevel,
		"DPANIC": zapcore.DPanicLevel,
		"PANIC":  zapcore.PanicLevel,
		"FATAL":  zapcore.FatalLevel,
	}
)

type loggerOpt func(*zap.Config) error

// New constructs a Sugared Logger that writes to stdout and
// provides human-readable timestamps.
func New(service string, opts ...loggerOpt) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]any{
		"service": service,
	}
	config.OutputPaths = []string{"stdout"}
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	for _, opt := range opts {
		if err := opt(&config); err != nil {
			return nil, err
		}
	}

	log, err := config.Build(zap.WithCaller(true))
	if err != nil {
		return nil, err
	}

	return log.Sugar(), nil
}

func NewStdLogger(log *zap.SugaredLogger) *log.Logger {
	return zap.NewStdLog(log.Desugar())
}

func WithLevel(level string) loggerOpt {
	return func(cfg *zap.Config) error {
		key := strings.ToUpper(level)
		lvl, ok := logLevels[key]
		if !ok {
			return fmt.Errorf("unknown log level %q", level)
		}
		cfg.Level = zap.NewAtomicLevelAt(lvl)
		return nil
	}
}

// WithZapConfig will overwrite the standard configurations provided by `New()`
// any loggerOpt provided AFTER this function when calling `New()` will
// continue to modify this provided config.
func WithZapConfig(config zap.Config) loggerOpt {
	return func(cfg *zap.Config) error {
		cfg = &config
		return nil
	}
}

// WithOutputPaths overrides the default OutputPaths of os.Stdout. Multiple
// files, URLs, can also be included in this function. For example:
// `WithOutputPaths("stdout", "/var/logs/myapp.log")` will print to a file and
// the standard output
func WithOutputPaths(outputPaths ...string) loggerOpt {
	return func(cfg *zap.Config) error {
		cfg.OutputPaths = outputPaths
		return nil
	}
}

// WithGCPMapping rewrites the zap config to utilize encoding values to conform
// to the standards used on Google Cloud logging systems. For more information
// refer to the following Github Issue/Discussion
// https://github.com/uber-go/zap/discussions/1110#discussioncomment-2955566
func WithGCPMapping() loggerOpt {
	return func(cfg *zap.Config) error {
		cfg.EncoderConfig.TimeKey = "time"
		cfg.EncoderConfig.LevelKey = "severity"
		cfg.EncoderConfig.NameKey = "logger"
		cfg.EncoderConfig.CallerKey = "caller"
		cfg.EncoderConfig.MessageKey = "message"
		cfg.EncoderConfig.StacktraceKey = "stacktrace"
		cfg.EncoderConfig.LineEnding = zapcore.DefaultLineEnding
		cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
		cfg.EncoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
		cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		cfg.EncoderConfig.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			switch l {
			case zapcore.DebugLevel:
				enc.AppendString("DEBUG")
			case zapcore.InfoLevel:
				enc.AppendString("INFO")
			case zapcore.WarnLevel:
				enc.AppendString("WARNING")
			case zapcore.ErrorLevel:
				enc.AppendString("ERROR")
			case zapcore.DPanicLevel:
				enc.AppendString("CRITICAL")
			case zapcore.PanicLevel:
				enc.AppendString("ALERT")
			case zapcore.FatalLevel:
				enc.AppendString("EMERGENCY")
			}
		}
		return nil
	}
}
