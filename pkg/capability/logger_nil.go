package capability

import (
	"context"

	"go.uber.org/fx"
)

func WithNilLogger() fx.Option {
	provide := func() Logger {
		return NilLogger{}
	}
	return fx.Provide(provide)
}

type NilLogger struct {
}

func (logger NilLogger) GetLogLevel() LogLevel {
	return DebugLogLevel
}

func (logger NilLogger) Debug(message string, fields map[string]interface{}) {

}

func (logger NilLogger) Info(message string, fields map[string]interface{}) {
}

func (logger NilLogger) Warn(message string, fields map[string]interface{}) {
}

func (logger NilLogger) Error(message string, fields map[string]interface{}) {
}

func (logger NilLogger) Fatal(message string, fields map[string]interface{}) {
}

func (logger NilLogger) DebugCtx(ctx context.Context, message string, fields map[string]interface{}) {

}

func (logger NilLogger) InfoCtx(ctx context.Context, message string, fields map[string]interface{}) {
}

func (logger NilLogger) WarnCtx(ctx context.Context, message string, fields map[string]interface{}) {
}

func (logger NilLogger) ErrorCtx(ctx context.Context, message string, fields map[string]interface{}) {
}

func (logger NilLogger) FatalCtx(ctx context.Context, message string, fields map[string]interface{}) {
}

func (logger NilLogger) WithName(string) Logger {
	return logger
}

func (logger NilLogger) WithFields(map[string]interface{}) Logger {
	return logger
}
