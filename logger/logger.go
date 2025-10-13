package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(isDev bool) (*zap.Logger, error) {
	var cfg zap.Config
	if isDev {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig.StacktraceKey = "stacktrace"
	}

	// Tambahin caller info supaya bisa tahu dari mana log berasal
	cfg.EncoderConfig.CallerKey = "caller"
	logger, err := cfg.Build(zap.AddCaller())
	if err != nil {
		return nil, err
	}

	return logger, nil
}
