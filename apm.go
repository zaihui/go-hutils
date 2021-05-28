package hutils

import (
	"context"
	"time"

	"go.elastic.co/apm"
)

// NewApmSpan 根据context，在当前transaction中生成新的span记录
func NewApmSpan(ctx context.Context, name, spanType string) *apm.Span {
	tx := apm.TransactionFromContext(ctx)
	opts := apm.SpanOptions{
		Start:  time.Now(),
		Parent: apm.SpanFromContext(ctx).TraceContext(),
	}
	return tx.StartSpanOptions(name, spanType, opts)
}
