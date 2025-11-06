package capability

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func WithZapLogger() fx.Option {
	type provideIn struct {
		fx.In
		Loader Loader `name:"configuration"`
	}

	type provideOut struct {
		fx.Out
		Logger Logger
	}

	provide := func(in provideIn) (out provideOut, err error) {
		configuration := ZapLoggerConfiguration{}
		if err = in.Loader.Load("logger.zap", &configuration); err != nil {
			return
		}
		if err = configuration.Validate(); err != nil {
			return
		}

		out.Logger, err = NewZapLogger(configuration)
		return
	}

	return fx.Provide(provide)
}

var zapLogLevels = map[string]LogLevel{
	"debug": DebugLogLevel,
	"info":  InfoLogLevel,
	"warn":  WarnLogLevel,
	"error": ErrorLogLevel,
	"fatal": FatalLogLevel,
}

type ZapLoggerConfiguration struct {
	Level       string
	Annotations struct {
		Caller     bool
		Stacktrace bool
	}
	OutputPaths      []string
	ErrorOutputPaths []string
	Labels           struct {
		Container string
		Bundle    string
	}
}

func (configuration ZapLoggerConfiguration) Validate() error {
	return validation.Errors{
		"Level": validation.Validate(configuration.Level, validation.Required, validation.In("debug", "info", "warn", "error", "fatal")),
		"Labels": validation.Errors{
			"Container": validation.Validate(configuration.Labels.Container, validation.Required),
			"Bundle":    validation.Validate(configuration.Labels.Bundle, validation.Required),
		}.Filter(),
	}.Filter()
}

var _ Logger = &zapLogger{}

type zapLogger struct {
	level LogLevel
	*zap.SugaredLogger
	fields map[string]interface{}
}

func NewZapLogger(configuration ZapLoggerConfiguration) (*zapLogger, error) {
	zapConfiguration := zap.NewProductionConfig()
	zapConfiguration.InitialFields = map[string]interface{}{
		"container": configuration.Labels.Container,
		"bundle":    configuration.Labels.Bundle,
	}
	if err := zapConfiguration.Level.UnmarshalText([]byte(configuration.Level)); err != nil {
		return nil, err
	}
	zapConfiguration.DisableCaller = !configuration.Annotations.Caller
	zapConfiguration.DisableStacktrace = !configuration.Annotations.Stacktrace
	zapConfiguration.OutputPaths = configuration.OutputPaths
	zapConfiguration.ErrorOutputPaths = configuration.ErrorOutputPaths

	if logger, err := zapConfiguration.Build(); err != nil {
		return nil, err
	} else {
		return &zapLogger{
			level:         zapLogLevels[configuration.Level],
			SugaredLogger: logger.Sugar(),
			fields:        make(map[string]interface{}),
		}, nil
	}
}

func (logger zapLogger) flatten(fields map[string]interface{}) []interface{} {
	mergedFields := make(map[string]interface{})
	for k, v := range logger.fields {
		mergedFields[k] = v
	}

	for k, v := range fields {
		if marshaller, ok := v.(LogFieldMarshaller); ok {
			mergedFields[k] = marshaller.MarshalLogField()
		} else {
			mergedFields[k] = v
		}
	}

	kv := make([]interface{}, 0, len(mergedFields)*2)
	for k, v := range mergedFields {
		kv = append(kv, k, v)
	}
	return kv
}

func (logger zapLogger) Match(filter map[string]string) bool {
	return true
}

func (logger zapLogger) Debug(message string, fields map[string]interface{}) {
	kv := logger.flatten(fields)
	logger.Debugw(message, kv...)
}

func (logger zapLogger) Info(message string, fields map[string]interface{}) {
	kv := logger.flatten(fields)
	logger.Infow(message, kv...)
}

func (logger zapLogger) Warn(message string, fields map[string]interface{}) {
	kv := logger.flatten(fields)
	logger.Warnw(message, kv...)
}

func (logger zapLogger) Error(message string, fields map[string]interface{}) {
	kv := logger.flatten(fields)
	logger.Errorw(message, kv...)
}

func (logger zapLogger) Fatal(message string, fields map[string]interface{}) {
	kv := logger.flatten(fields)
	logger.Debugw(message, kv...)
}

func (logger zapLogger) DebugCtx(ctx context.Context, message string, fields map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	fields["span"] = map[string]interface{}{
		"traceId": span.SpanContext().TraceID().String(),
		"spanId":  span.SpanContext().SpanID().String(),
	}
	logger.Debug(message, fields)
}

func (logger zapLogger) InfoCtx(ctx context.Context, message string, fields map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	fields["span"] = map[string]interface{}{
		"traceId": span.SpanContext().TraceID().String(),
		"spanId":  span.SpanContext().SpanID().String(),
	}
	logger.Info(message, fields)
}

func (logger zapLogger) WarnCtx(ctx context.Context, message string, fields map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	fields["span"] = map[string]interface{}{
		"traceId": span.SpanContext().TraceID().String(),
		"spanId":  span.SpanContext().SpanID().String(),
	}
	logger.Warn(message, fields)
}

func (logger zapLogger) ErrorCtx(ctx context.Context, message string, fields map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	fields["span"] = map[string]interface{}{
		"traceId": span.SpanContext().TraceID().String(),
		"spanId":  span.SpanContext().SpanID().String(),
	}
	logger.Error(message, fields)
}

func (logger zapLogger) FatalCtx(ctx context.Context, message string, fields map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	fields["span"] = map[string]interface{}{
		"traceId": span.SpanContext().TraceID().String(),
		"spanId":  span.SpanContext().SpanID().String(),
	}
	logger.Fatal(message, fields)
}

func (logger zapLogger) GetLogLevel() LogLevel {
	return logger.level
}

func (logger zapLogger) WithFields(fields map[string]interface{}) Logger {
	newFields := make(map[string]interface{}, len(logger.fields)+len(fields))
	for k, v := range logger.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	return &zapLogger{
		level:         logger.level,
		SugaredLogger: logger.SugaredLogger,
		fields:        newFields,
	}
}

func (logger zapLogger) WithName(name string) Logger {
	sugaredLogger := logger.SugaredLogger.Named(name)
	return &zapLogger{
		level:         logger.level,
		SugaredLogger: sugaredLogger,
		fields:        logger.fields,
	}
}
