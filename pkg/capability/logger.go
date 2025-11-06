package capability

import "context"

type LogLevel uint8

const (
	DebugLogLevel = LogLevel(0)
	InfoLogLevel  = LogLevel(1)
	WarnLogLevel  = LogLevel(2)
	ErrorLogLevel = LogLevel(3)
	FatalLogLevel = LogLevel(4)
)

type Logger interface {
	GetLogLevel() LogLevel
	Debug(message string, fields map[string]interface{})
	Info(message string, fields map[string]interface{})
	Warn(message string, fields map[string]interface{})
	Error(message string, fields map[string]interface{})
	Fatal(message string, fields map[string]interface{})
	DebugCtx(ctx context.Context, message string, fields map[string]interface{})
	InfoCtx(ctx context.Context, message string, fields map[string]interface{})
	WarnCtx(ctx context.Context, message string, fields map[string]interface{})
	ErrorCtx(ctx context.Context, message string, fields map[string]interface{})
	FatalCtx(ctx context.Context, message string, fields map[string]interface{})
	WithName(string) Logger
	WithFields(map[string]interface{}) Logger
}

type LogFieldMarshaller interface {
	MarshalLogField() interface{}
}
