package log

import (
	"context"
	"errors"
	"fmt"
	"io"
	//"mailganer_test_task/internal/services"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ConfigLog конфигурация для логгера
type ConfigLog struct {
	Mode	string	`env:"LOG_MODE"`
	Level	string	`env:"LOG_LEVEL"`
	Output	string	`env:"LOG_OUTPUT"`
}

// DevConfig Конфигурация конфига для тестов
var DevConfig = &ConfigLog{
	Mode:	"development",
	Level:	"debug",
	Output:	"stdout",
}

// Log - logger с обвязкой zap
type Log struct {
	*zap.Logger
}

// e Ядро логгера
var e *zap.Logger

// InitLogger ...
func InitLogger(c *ConfigLog) (*Log, error) {
	pe, err := SetMod(c.Mode)
	if err != nil {
		return nil, err
	}

	consoleEncoder := zapcore.NewJSONEncoder(pe)

	level, err := SetLevel(c.Level)
	if err != nil {
		return nil, err
	}

	output, err := SetOutput(c.Output)
	if err != nil {
		return nil, err
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(output), level),
	)

	e = zap.New(core)
	return &Log{
		e,
	}, nil
}

// SetMod возвращает конфигурацию для zap
func SetMod(mod string) (zapcore.EncoderConfig, error) {
	switch mod {
	case "development":
		return zap.NewDevelopmentEncoderConfig(), nil
	case "production":
		pe := zap.NewProductionEncoderConfig()
		pe.TimeKey = "time"
		pe.EncodeTime = zapcore.TimeEncoderOfLayout("02-01-2006 15:04:05 MST")
		return pe, nil
	default:
		return zapcore.EncoderConfig{}, fmt.Errorf("unknown mod - %s. Sopported only: development or production mod", mod)
	}
}

// SetLevel возвращает уровень, который передан в конфиге
func SetLevel(level string) (zapcore.Level, error) {
	switch level {
	case "panic":
		return zap.PanicLevel, nil
	case "fatal":
		return zap.FatalLevel, nil
	case "error":
		return zap.ErrorLevel, nil
	case "warn":
		return zap.WarnLevel, nil
	case "info":
		return zap.InfoLevel, nil
	case "debug":
		return zap.DebugLevel, nil
	default:
		return 0, fmt.Errorf("unknown logging level - %s. Supported only: panic, fatal, error, warn, info, debug", level)
	}
}

// SetOutput назначает источник вывода логов
func SetOutput(output string) (io.Writer, error) {
	switch output {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		return nil, fmt.Errorf("unknown output - %s. Supported only stdout or stderr", output)
	}
}

// LogGracefulShutdown - очищает буферы и корректно завершает работу
func (l *Log) LogGracefulShutdown() error {
	err := l.Sync()
	if err != nil {
		return err
	}

	return nil
}

// WithContext - метод берет информацию по span_id, parent_span_id и trace_id из контекста и передает в logger
func (l *Log) WithContext(ctx context.Context) *Log {
	info := getFromContext(ctx)
	e := l.With(msg(info))
	return &Log{e}
}

// Debug - отладочный вывод в лог
func (l *Log) Debug(msg string) {
	l.Logger.Debug(msg)
}

// Debugf - отладочный форматированный вывод в лог
func (l *Log) Debugf(format string, v ...interface{}) {
	l.Debug(fmt.Sprintf(format, v...))
}

// Info - вывод в лог уровня info
func (l *Log) Info(msg string) {
	l.Logger.Info(msg)
}

// Infof - форматированный вывод в лог уровня info
func (l *Log) Infof(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v...))
}

// Warn - вывод в лог уровня warn
func (l *Log) Warn(msg string) {
	l.Logger.Warn(msg)
}

// Warnf - форматированный вывод в лог уровня warn
func (l *Log) Warnf(format string, v ...interface{}) {
	l.Warn(fmt.Sprintf(format, v...))
}

// Error - вывод в лог уровня error
func (l *Log) Error(msg string) {
	l.Logger.Error(msg)
}

// Errorf - форматированный вывод в лог уровня error
func (l *Log) Errorf(format string, v ...interface{}) {
	l.Error(fmt.Sprintf(format, v...))
}

// RError - вывод в лог уровня error и возврат ошибки
func (l *Log) RError(msg string) error {
	l.Error(msg)
	return errors.New(msg)
}

// RErrorf - форматированный вывод в лог уровня error и возврат ошибки
func (l *Log) RErrorf(format string, v ...interface{}) error {
	l.Error(fmt.Sprintf(format, v...))
	return fmt.Errorf(format, v...)
}

// Fatal - вывод в лог уровня fatal (приведёт к вызову os.Exit(1) и завершению работы программы)
func (l *Log) Fatal(msg string) {
	l.Logger.Fatal(msg)
}

// Fatalf - форматированный вывод в лог уровня fatal (приведёт к вызову os.Exit(1) и завершению работы программы)
func (l *Log) Fatalf(format string, v ...interface{}) {
	l.Fatal(fmt.Sprintf(format, v...))
}

// Вспомогательная функция для добавления к выводу trace_id, span_id и parent_span_id
func msg(info *ContextInfo) zap.Field {
	field := zap.Inline(info)
	return field
}
