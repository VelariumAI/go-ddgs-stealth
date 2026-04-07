package goddgs

import (
	"context"
	"errors"
	"testing"
)

func TestTracingHelpers(t *testing.T) {
	ctx, span := startSpan(context.Background(), "test.span")
	if ctx == nil || span == nil {
		t.Fatal("expected span")
	}
	endSpan(span, nil)
	_, span2 := startSpan(context.Background(), "test.span.err")
	endSpan(span2, errors.New("boom"))
}
