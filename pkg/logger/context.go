package log

import (
	"context"
	"go.uber.org/zap/zapcore"
)

//
const (
	traceStr      = "trace_id"
	spanStr       = "span_id"
	parentSpanStr = "parent_span_id"
)

// Ключи для хранения идентификаторов вызовов

// TraceID - ключ для чтения/записи сквозного идентификатора вызова
type TraceID struct{}

// SpanID - ключ для чтения/записи идентификатора вызова
type SpanID struct{}

// ParentSpanID - ключ для чтения/записи родительского идентификатора вызова
type ParentSpanID struct{}

// ContextInfo - хранит информацию из контекста в виде структуры
type ContextInfo struct {
	TraceID      string `json:"trace_id"`
	SpanID       string `json:"span_id"`
	ParentSpanID string `json:"parent_span_id"`
}

// ContextWithTrace
func ContextWithTrace(ctx context.Context, trace string) context.Context {
	ctx = context.WithValue(ctx, TraceID{}, trace)
	return ctx
}

// ContextWithSpan
func ContextWithSpan(ctx context.Context, span string) context.Context {
	ctx = contextWithParentSpan(ctx)
	ctx = context.WithValue(ctx, SpanID{}, span)
	return ctx
}

// contextWithParentSpan
func contextWithParentSpan(ctx context.Context) context.Context {
	span, ok := ctx.Value(SpanID{}).(string)
	if !ok {
		return ctx
	}

	if span == "" {
		return ctx
	}

	ctx = context.WithValue(ctx, ParentSpanID{}, span)

	return ctx
}

// getFromContext Берет информацию из контекста
func getFromContext(ctx context.Context) *ContextInfo {
	traceID, _ := ctx.Value(TraceID{}).(string)
	spanID, _ := ctx.Value(SpanID{}).(string)
	parentSpanID, _ := ctx.Value(ParentSpanID{}).(string)

	return &ContextInfo{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
	}
}

// MarshalLogObject Записывает ключи в JSON объект
func (c *ContextInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString(traceStr, c.TraceID)
	enc.AddString(spanStr, c.SpanID)
	enc.AddString(parentSpanStr, c.ParentSpanID)
	return nil
}
